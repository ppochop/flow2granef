package ipfix

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v240"
	"github.com/ppochop/flow2granef/flowutils"
	"github.com/ppochop/flow2granef/profiles"
	dgraphhelpers "github.com/ppochop/flow2granef/profiles/dgraph_helpers"
	xidcache "github.com/ppochop/flow2granef/xid-cache"
	"github.com/satta/gommunityid"
)

type IpfixprobeTransformer struct {
	commIdGen gommunityid.CommunityID
	cache     *xidcache.IdCache
	dgoClient *dgo.Dgraph
	stats     profiles.TransformerStats
}

type IpfixprobeTransformerDuplCheck struct {
	instanceName string
	commIdGen    gommunityid.CommunityID
	cache        *xidcache.DuplCache
}

func init() {
	profiles.RegisterTransformer("ipfixprobe", InitIpfixprobeTransformer)
	profiles.RegisterDuplCheckTransformer("ipfixprobe", InitIpfixprobeTransformerDuplCheck)
	profiles.RegisterPreHandler("ipfixprobe", IpfixPreHandle)
}

func InitIpfixprobeTransformer(cache *xidcache.IdCache, dgoClient *dgo.Dgraph, stats profiles.TransformerStats) profiles.Transformer {
	commId, _ := gommunityid.GetCommunityIDByVersion(1, 0)
	return &IpfixprobeTransformer{
		commIdGen: commId,
		cache:     cache,
		dgoClient: dgoClient,
		stats:     stats,
	}
}

func InitIpfixprobeTransformerDuplCheck(cache *xidcache.DuplCache, name string) profiles.Transformer {
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
	commId := s.commIdGen.CalcBase64(ft)
	flow.CommId = commId
	xid := strconv.FormatUint(event.FlowId, 10)

	var hit xidcache.CacheHitResult
	switch flow.FlushReason {
	case flowutils.ActiveTimeout:
		xid, hit = s.cache.AddOrGet(commId, false, xid, flow.FirstTs, flow.LastTs)
	default:
		var foundXid string
		foundXid, hit = s.cache.Get(commId, flow.FirstTs)
		if hit != xidcache.Miss {
			xid = foundXid
		}
	}
	defer profiles.TimeTrack(time.Now(), s.stats.ProcessingTimeFlow)
	err = dgraphhelpers.HandleFlow(ctx, s.dgoClient, flow, xid, hit, &s.stats)
	switch {
	case event.IsDnsAnswer():
		defer profiles.TimeTrack(time.Now(), s.stats.ProcessingTimeDns)
		dns := event.GetGranefDNSRec()
		if dns == nil {
			return nil
		}
		err = dgraphhelpers.HandleDns(ctx, s.dgoClient, dns, xid, &s.stats)
	case event.HTTPHost != nil:
		defer profiles.TimeTrack(time.Now(), s.stats.ProcessingTimeHttp)
		http := event.GetGranefHTTPRec()
		if http == nil {
			return nil
		}
		http.ClientIp = flow.OrigIp
		http.ServerIp = flow.RespIp
		err = dgraphhelpers.HandleHttp(ctx, s.dgoClient, http, xid, &s.stats)
	}
	if err == nil {
		s.stats.EventsTransformed.Inc()
	}
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
