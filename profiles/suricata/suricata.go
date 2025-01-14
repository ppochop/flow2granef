package suricata

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"

	"github.com/dgraph-io/dgo/v240"
	"github.com/ppochop/flow2granef/flowutils"
	"github.com/ppochop/flow2granef/profiles"
	dgraphhelpers "github.com/ppochop/flow2granef/profiles/dgraph_helpers"
	"github.com/satta/gommunityid"
)

type SuricataTransformer struct {
	commIdGen gommunityid.CommunityID
	reTime    *regexp.Regexp
	cache     profiles.Cache
	dgoClient *dgo.Dgraph
	stats     profiles.TransformerStats
}

type SuricataTransformerDuplCheck struct {
	instanceName string
	commIdGen    gommunityid.CommunityID
	cache        profiles.CacheDuplCheck
}

func init() {
	profiles.RegisterPreHandler("suricata", PreHandle)
	profiles.RegisterTransformer("suricata", InitSuricataTransformer)
	profiles.RegisterDuplCheckTransformer("suricata", InitSuricataTransformerDuplCheck)
}

func InitSuricataTransformer(cache profiles.Cache, dgoClient *dgo.Dgraph, stats profiles.TransformerStats) profiles.Transformer {
	commId, _ := gommunityid.GetCommunityIDByVersion(1, 0)
	return &SuricataTransformer{
		commIdGen: commId,
		cache:     cache,
		dgoClient: dgoClient,
		stats:     stats,
	}
}

func InitSuricataTransformerDuplCheck(cache profiles.CacheDuplCheck, name string) profiles.Transformer {
	commId, _ := gommunityid.GetCommunityIDByVersion(1, 0)
	return &SuricataTransformerDuplCheck{
		instanceName: name,
		commIdGen:    commId,
		cache:        cache,
	}
}

func PreHandle(data []byte) (uint32, error) {
	preflow := SuricataPreEve{}
	err := json.Unmarshal(data, &preflow)
	if err != nil {
		return 0, err
	}
	id := preflow.GetPreflowId()
	return id, nil
}

func (s *SuricataTransformer) handleFlow(ctx context.Context, event *SuricataEve) error {
	flow := event.GetGranefFlowRec("suricata")
	ft := flow.GetFlowTuple()

	commId := s.commIdGen.CalcBase64(ft)
	flow.CommId = commId
	xid := strconv.FormatUint(event.FlowId, 10)
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

func (s *SuricataTransformer) handleDns(ctx context.Context, event *SuricataEve) error {
	if len(event.Dns.Answers) == 0 {
		return nil
	}
	dns := event.Dns.GetGranefDNSRec()
	flow := event.GetGranefMiminalFlowRec("suricata")
	ft := flow.GetFlowTuple()

	commId := s.commIdGen.CalcBase64(ft)
	flow.CommId = commId
	xid := strconv.FormatUint(event.FlowId, 10)

	return dgraphhelpers.HandleDnsWithFlowPlaceholder(ctx, s.dgoClient, dns, xid, &s.stats)
}

func (s *SuricataTransformer) Handle(ctx context.Context, data []byte) error {
	event := SuricataEve{}
	err := json.Unmarshal(data, &event)
	if err != nil {
		return err
	}
	s.stats.EventsProcessed.Inc()
	switch event.DetermineEventType() {
	case SuricataEventFlow:
		err = s.handleFlow(ctx, &event)
	case SuricataEventDns:
		err = s.handleDns(ctx, &event)
	case SuricataEventHttp:
		return nil
	default:
		return fmt.Errorf("unsupported suricata event type: %s", event.EventType)
	}
	if err == nil {
		s.stats.EventsTransformed.Inc()
	}
	return err
}

func (s *SuricataTransformerDuplCheck) Handle(ctx context.Context, data []byte) error {
	eveLog := SuricataEve{}
	err := json.Unmarshal(data, &eveLog)
	if err != nil {
		return err
	}
	if eveLog.DetermineEventType() != SuricataEventFlow {
		return nil
	}
	flow := eveLog.GetGranefFlowRec("suricata")
	ft := flow.GetFlowTuple()
	commId := s.commIdGen.CalcBase64(ft)
	sourceHit, found := s.cache.DuplHandle(commId, flow.FirstTs, flow.LastTs, s.instanceName)
	if found && sourceHit != s.instanceName {
		slog.Warn("Duplicate flow record found.", "community-id", commId, "found_record_source", sourceHit, "attempt_source", s.instanceName)
	}
	return nil
}

func (s *SuricataTransformerDuplCheck) GetStats() map[string]uint {
	return map[string]uint{}
}
