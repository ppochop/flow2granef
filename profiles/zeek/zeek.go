package zeek

/*
Flow records made from Zeek's conn log will use Zeek's uid as xid in dgraph.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/dgraph-io/dgo/v240"
	dgraphhelpers "github.com/ppochop/flow2granef/dgraph-helpers"
	"github.com/ppochop/flow2granef/flowutils"
	ipproto "github.com/ppochop/flow2granef/ip-proto"
	"github.com/ppochop/flow2granef/profiles"
	xidcache "github.com/ppochop/flow2granef/xid-cache"
	"github.com/satta/gommunityid"
)

type ZeekTransformer struct {
	commIdGen gommunityid.CommunityID
	cache     *xidcache.IdCache
	dgoClient *dgo.Dgraph
	stats     profiles.TransformerStats
}

type ZeekTransformerDuplCheck struct {
	instanceName string
	commIdGen    gommunityid.CommunityID
	cache        *xidcache.DuplCache
}

func init() {
	profiles.RegisterPreHandler("zeek", PreHandle)
	profiles.RegisterTransformer("zeek", InitZeekTransformer)
	profiles.RegisterDuplCheckTransformer("zeek", InitZeekTransformerDuplCheck)
}

func InitZeekTransformer(cache *xidcache.IdCache, dgoClient *dgo.Dgraph, stats profiles.TransformerStats) profiles.Transformer {
	commId, _ := gommunityid.GetCommunityIDByVersion(1, 0)
	return &ZeekTransformer{
		commIdGen: commId,
		cache:     cache,
		dgoClient: dgoClient,
		stats:     stats,
	}
}

func InitZeekTransformerDuplCheck(cache *xidcache.DuplCache, name string) profiles.Transformer {
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

func (z *ZeekTransformer) handleFlow(ctx context.Context, event *ZeekConn) error {
	flow := event.GetGranefFlowRec("zeek")
	ft := flow.GetFlowTuple()

	commId := z.commIdGen.CalcBase64(ft)
	flow.CommId = commId
	xid := event.Uid
	var hit xidcache.CacheHitResult

	switch flow.FlushReason {
	case flowutils.ActiveTimeout:
		xid, hit = z.cache.AddOrGet(commId, false, xid, flow.FirstTs, flow.LastTs)
	default:
		var foundXid string
		foundXid, hit = z.cache.Get(commId, flow.FirstTs)
		if hit != xidcache.Miss {
			xid = foundXid
		}
	}

	dgraphhelpers.HandleFlow(ctx, z.dgoClient, flow, xid, hit, &z.stats)
	return nil
}

func (z *ZeekTransformer) handleDns(ctx context.Context, eventDns *ZeekDns, eventConnL *ZeekConnLimited) error {
	dns := eventDns.GetGranefDNSRec()
	if len(dns.Answer) == 0 {
		return nil
	}
	flow := eventConnL.GetGranefFlowRec("zeek")
	ft := flow.GetFlowTuple()

	commId := z.commIdGen.CalcBase64(ft)
	flow.CommId = commId
	xid := eventConnL.Uid
	xid, _ = z.cache.AddOrGet(commId, true, xid, flow.FirstTs, flow.LastTs)
	return dgraphhelpers.HandleDnsWithFlowPlaceholder(ctx, z.dgoClient, dns, xid, &z.stats)
}

func (z *ZeekTransformer) handleHttp(ctx context.Context, eventHttp *ZeekHttp, eventConnL *ZeekConnLimited) error {
	http := eventHttp.GetGranefHTTPRec()
	flow := eventConnL.GetGranefFlowRec("zeek")
	flow.Protocol = ipproto.ProtocolFromName("tcp") // Zeek HTTP records are missing the protocol field
	ft := flow.GetFlowTuple()
	http.ClientIp = flow.OrigIp
	http.ServerIp = flow.RespIp

	commId := z.commIdGen.CalcBase64(ft)
	flow.CommId = commId
	xid := eventConnL.Uid
	xid, _ = z.cache.AddOrGet(commId, true, xid, flow.FirstTs, flow.LastTs)
	return dgraphhelpers.HandleHttpWithFlowPlaceholder(ctx, z.dgoClient, http, xid, &z.stats)
}

func (z *ZeekTransformer) Handle(ctx context.Context, data []byte) error {
	zeekBase := ZeekBase{}
	err := json.Unmarshal(data, &zeekBase)
	if err != nil {
		return err
	}
	z.stats.EventsProcessed.Inc()
	zType := zeekBase.decideType()
	switch zType {
	case ZeekLogConn:
		defer profiles.TimeTrack(time.Now(), z.stats.ProcessingTimeFlow)
		event := ZeekConn{}
		err = json.Unmarshal(data, &event)
		if err != nil {
			return err
		}
		err = z.handleFlow(ctx, &event)
	case ZeekLogDns:
		defer profiles.TimeTrack(time.Now(), z.stats.ProcessingTimeDns)
		eventDns := ZeekDns{}
		eventConnL := ZeekConnLimited{}
		err = json.Unmarshal(data, &eventDns)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &eventConnL)
		if err != nil {
			return err
		}
		err = z.handleDns(ctx, &eventDns, &eventConnL)

	case ZeekLogHttp:
		defer profiles.TimeTrack(time.Now(), z.stats.ProcessingTimeHttp)
		eventHttp := ZeekHttp{}
		eventConnL := ZeekConnLimited{}
		err = json.Unmarshal(data, &eventHttp)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &eventConnL)
		if err != nil {
			return err
		}
		err = z.handleHttp(ctx, &eventHttp, &eventConnL)

	default:
		return fmt.Errorf("unsupported zeek event type for uid %s", zeekBase.Uid)
	}
	if err == nil {
		z.stats.EventsTransformed.Inc()
	}
	return err
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
