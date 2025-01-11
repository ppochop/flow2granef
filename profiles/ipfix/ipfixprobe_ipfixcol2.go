package ipfix

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/dgraph-io/dgo/v240"
	"github.com/ppochop/flow2granef/flowutils"
	"github.com/ppochop/flow2granef/profiles"
	dgraphhelpers "github.com/ppochop/flow2granef/profiles/dgraph_helpers"
	"github.com/satta/gommunityid"
)

type IpfixprobeTransformer struct {
	commIdGen gommunityid.CommunityID
	cache     profiles.Cache
	dgoClient *dgo.Dgraph
	stats     profiles.TransformerStats
}

type IpfixprobeTransformerDuplCheck struct {
	instanceName string
	commIdGen    gommunityid.CommunityID
	cache        profiles.CacheDuplCheck
}

func init() {
	profiles.RegisterTransformer("ipfixprobe", InitIpfixprobeTransformer)
	profiles.RegisterDuplCheckTransformer("ipfixprobe", InitIpfixprobeTransformerDuplCheck)
	profiles.RegisterPreHandler("ipfixprobe", IpfixPreHandle)
}

func InitIpfixprobeTransformer(cache profiles.Cache, dgoClient *dgo.Dgraph, stats profiles.TransformerStats) profiles.Transformer {
	commId, _ := gommunityid.GetCommunityIDByVersion(1, 0)
	return &IpfixprobeTransformer{
		commIdGen: commId,
		cache:     cache,
		dgoClient: dgoClient,
		stats:     stats,
	}
}

func InitIpfixprobeTransformerDuplCheck(cache profiles.CacheDuplCheck, name string) profiles.Transformer {
	commId, _ := gommunityid.GetCommunityIDByVersion(1, 0)
	return &IpfixprobeTransformerDuplCheck{
		instanceName: name,
		commIdGen:    commId,
		cache:        cache,
	}
}

func IpfixPreHandle(data []byte) (uint32, error) {
	preflow := IpfixprobePreFlow{}
	err := json.Unmarshal(data, &preflow)
	if err != nil {
		return 0, err
	}
	id := preflow.GetIpfixPreflowId()
	return id, nil
}

func (s *IpfixprobeTransformer) Handle(ctx context.Context, data []byte) error {
	event := IpfixprobeFlow{}
	err := json.Unmarshal(data, &event)
	if err != nil {
		return err
	}
	s.stats.EventsProcessed.Inc()
	flow := event.GetGranefFlowRec("ipfixprobe")
	ft := flow.GetFlowTuple()

	commIdHash := s.commIdGen.Hash(ft)
	commId := s.commIdGen.RenderBase64(commIdHash)
	flowId := bytes.Buffer{}
	err = binary.Write(&flowId, binary.LittleEndian, event.FlowId)
	if err != nil {
		return fmt.Errorf("failed to parse flowid %d into byte slice", event.FlowId)
	}
	flowIdHash := commIdHash.Sum(flowId.Bytes())
	xid := base64.StdEncoding.EncodeToString(flowIdHash)
	hit := false

	switch flow.FlushReason {
	case flowutils.ActiveTimeout:
		xid, hit = s.cache.AddOrGet(commId, xid, flow.LastTs)
	default:
		foundXid, hit := s.cache.Get(commId)
		if hit {
			xid = foundXid
		}
	}
	dgraphhelpers.HandleFlow(ctx, s.dgoClient, flow, xid, hit, &s.stats)
	return nil
}

func (s *IpfixprobeTransformerDuplCheck) Handle(ctx context.Context, data []byte) error {
	event := IpfixprobeFlow{}
	err := json.Unmarshal(data, &event)
	if err != nil {
		return err
	}
	flow := event.GetGranefFlowRec("ipfixprobe")
	ft := flow.GetFlowTuple()
	commId := s.commIdGen.CalcBase64(ft)
	sourceHit, found := s.cache.DuplHandle(commId, flow.FirstTs, flow.LastTs, s.instanceName)
	if found && sourceHit != s.instanceName {
		slog.Warn("Duplicate flow record found.", "community-id", commId, "found_record_source", sourceHit, "attempt_source", s.instanceName)
	}
	return nil
}
