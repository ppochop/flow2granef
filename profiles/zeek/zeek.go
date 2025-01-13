package zeek

/*
Flow records made from Zeek's conn log will use Zeek's uid as xid in dgraph.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/dgraph-io/dgo/v240"
	"github.com/ppochop/flow2granef/flowutils"
	"github.com/ppochop/flow2granef/profiles"
	dgraphhelpers "github.com/ppochop/flow2granef/profiles/dgraph_helpers"
	"github.com/satta/gommunityid"
)

type ZeekTransformer struct {
	commIdGen gommunityid.CommunityID
	cache     profiles.Cache
	dgoClient *dgo.Dgraph
	stats     profiles.TransformerStats
}

type ZeekTransformerDuplCheck struct {
	instanceName string
	commIdGen    gommunityid.CommunityID
	cache        profiles.CacheDuplCheck
}

func init() {
	profiles.RegisterPreHandler("zeek", PreHandle)
	profiles.RegisterTransformer("zeek", InitZeekTransformer)
	profiles.RegisterDuplCheckTransformer("zeek", InitZeekTransformerDuplCheck)
}

func InitZeekTransformer(cache profiles.Cache, dgoClient *dgo.Dgraph, stats profiles.TransformerStats) profiles.Transformer {
	commId, _ := gommunityid.GetCommunityIDByVersion(1, 0)
	return &ZeekTransformer{
		commIdGen: commId,
		cache:     cache,
		dgoClient: dgoClient,
		stats:     stats,
	}
}

func InitZeekTransformerDuplCheck(cache profiles.CacheDuplCheck, name string) profiles.Transformer {
	commId, _ := gommunityid.GetCommunityIDByVersion(1, 0)
	return &ZeekTransformerDuplCheck{
		instanceName: name,
		commIdGen:    commId,
		cache:        cache,
	}
}

func PreHandle(data []byte) (uint32, error) {
	preflow := ZeekPre{}
	err := json.Unmarshal(data, &preflow)
	if err != nil {
		return 0, err
	}
	id := preflow.GetPreflowId()
	return id, nil
}

func (s *ZeekTransformer) handleFlow(ctx context.Context, event *ZeekConn) error {
	flow := event.GetGranefFlowRec("zeek")
	ft := flow.GetFlowTuple()

	commId := s.commIdGen.CalcBase64(ft)
	flow.CommId = commId
	xid := event.Uid
	hit := false

	switch flow.FlushReason {
	case flowutils.ActiveTimeout:
		xid, hit = s.cache.AddOrGet(commId, xid, flow.FirstTs, flow.LastTs)
	default:
		foundXid, hit := s.cache.Get(commId, flow.FirstTs)
		if hit {
			xid = foundXid
		}
	}
	dgraphhelpers.HandleFlow(ctx, s.dgoClient, flow, xid, hit, &s.stats)
	return nil
}

func (z *ZeekTransformer) Handle(ctx context.Context, data []byte) error {
	zeekBase := ZeekBase{}
	err := json.Unmarshal(data, &zeekBase)
	if err != nil {
		return err
	}
	zType := zeekBase.decideType()
	switch zType {
	case ZeekLogConn:
		event := ZeekConn{}
		err := json.Unmarshal(data, &event)
		if err != nil {
			return err
		}
		return z.handleFlow(ctx, &event)
	default:
		return fmt.Errorf("unsupported zeek event type")
	}
}

func (z *ZeekTransformerDuplCheck) Handle(ctx context.Context, data []byte) error {
	zeekBase := ZeekBase{}
	err := json.Unmarshal(data, &zeekBase)
	if err != nil {
		return err
	}
	switch zeekBase.decideType() {
	case ZeekLogConn:
		zC := ZeekConn{}
		err := json.Unmarshal(data, &zC)
		if err != nil {
			return err
		}
		flow := zC.GetGranefFlowRec("zeek")
		ft := flow.GetFlowTuple()
		commId := z.commIdGen.CalcBase64(ft)
		sourceHit, found := z.cache.DuplHandle(commId, flow.FirstTs, flow.LastTs, z.instanceName)
		if found && sourceHit != z.instanceName {
			slog.Warn("Duplicate flow record found.", "community-id", commId, "found_record_source", sourceHit, "attempt_source", z.instanceName)
		}
	default:
		return nil
	}
	return nil
}

func GetZeekFlushReason(connState string) flowutils.FlushReason {
	switch connState {
	case "S0", "S1", "OTH":
		return flowutils.PassiveTimeout
	case "SF", "REJ", "S2", "S3", "RSTO", "RSTR", "RSTOS0", "RSTRH", "SH", "SHR":
		return flowutils.Finished
	default:
		return flowutils.Unknown
	}
}
