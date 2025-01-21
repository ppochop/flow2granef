package dgraphhelpers

import (
	"fmt"

	"github.com/dgraph-io/dgo/v240/protos/api"
	"github.com/ppochop/flow2granef/flowutils"
	xidcache "github.com/ppochop/flow2granef/xid-cache"
)

func buildFlowRecPlaceholderTxn(xid string) *api.Request {
	query := fmt.Sprintf(`
		query {
			Flow as var(func: eq(FlowRec.id, "%s"))
		}
	`, xid)
	mutation := fmt.Sprintf(`
		uid(Flow) <dgraph.type> "FlowRec" .
		uid(Flow) <FlowRec.id> "%s" .
	`, xid)
	mut := &api.Mutation{
		SetNquads: []byte(mutation),
		Cond:      `@if(eq(len(Flow), 0))`,
	}
	return &api.Request{
		Query:     query,
		CommitNow: true,
		Mutations: []*api.Mutation{mut},
	}
}

func buildFlowRecTxn(f *flowutils.FlowRec, xid string, cacheHit xidcache.CacheHitResult) *api.Request {
	var query string
	var flowMutations string
	fMu := &api.Mutation{}
	req := &api.Request{CommitNow: true}
	if cacheHit == xidcache.Hit {
		query = fmt.Sprintf(`
			query {
				Flow as var(func: eq(FlowRec.id, "%s")) {
					OB as FlowRec.from_orig_bytes
					RB as FlowRec.from_recv_bytes
					OP as FlowRec.from_orig_pkts
					RP as FlowRec.from_recv_pkts
					NOB as math(OB + %d)
					NRB as math(RB + %d)
					NOP as math(OP + %d)
					NRP as math(RP + %d)
				}
			}
		`, xid, f.OrigBytes, f.RespBytes, f.OrigPkts, f.RespPkts)
		flowMutations = fmt.Sprintf(`
			uid(Flow) <dgraph.type> "FlowRec" .
			uid(Flow) <FlowRec.last_ts> "%s" .
			uid(Flow) <FlowRec.flush_reason> "%s" .
			uid(Flow) <FlowRec.flow_source> "%s" .
			uid(Flow) <FlowRec.from_orig_bytes> val(NOB) .
			uid(Flow) <FlowRec.from_recv_bytes> val(NRB) .
			uid(Flow) <FlowRec.from_orig_pkts> val(NOP) .
			uid(Flow) <FlowRec.from_recv_pkts> val(NRP) .
		`, f.LastTs, f.FlushReason, f.FlowSource)
		fMu.SetNquads = []byte(flowMutations)
		req.Mutations = []*api.Mutation{fMu}
	} else {
		query = fmt.Sprintf(`
			query {
				Orig as var(func: eq(Host.ip, "%s"))
				Resp as var(func: eq(Host.ip, "%s"))
				Flow as var(func: eq(FlowRec.id, "%s"))
			}
		`, f.OrigIp.StringExpanded(), f.RespIp.StringExpanded(), xid)
		flowMutations = fmt.Sprintf(`
			uid(Flow) <dgraph.type> "FlowRec" .
			uid(Flow) <FlowRec.id> "%s" .
			uid(Flow) <FlowRec.community_id> "%s" .
			uid(Flow) <FlowRec.originated_by> uid(Orig) .
			uid(Flow) <FlowRec.received_by> uid(Resp) .
			uid(Flow) <FlowRec.orig_port> "%d" .
			uid(Flow) <FlowRec.recv_port> "%d" .
			uid(Flow) <FlowRec.from_orig_bytes> "%d" .
			uid(Flow) <FlowRec.from_recv_bytes> "%d" .
			uid(Flow) <FlowRec.from_orig_pkts> "%d" .
			uid(Flow) <FlowRec.from_recv_pkts> "%d" .
			uid(Flow) <FlowRec.first_ts> "%s" .
			uid(Flow) <FlowRec.last_ts> "%s" .
			uid(Flow) <FlowRec.protocol> "%s" .
			uid(Flow) <FlowRec.app> "%s" .
			uid(Flow) <FlowRec.flush_reason> "%s" .
			uid(Flow) <FlowRec.flow_source> "%s" .
		`, xid, f.CommId, f.OrigPort, f.RespPort, f.OrigBytes, f.RespBytes, f.OrigPkts, f.RespPkts, f.FirstTs, f.LastTs, f.Protocol.GetName(), f.App, f.FlushReason, f.FlowSource)
		origMutations := fmt.Sprintf(`
			uid(Orig) <dgraph.type> "Host" .
			uid(Orig) <Host.ip> "%s" .
		`, f.OrigIp.StringExpanded())
		respMutations := fmt.Sprintf(`
			uid(Resp) <dgraph.type> "Host" .
			uid(Resp) <Host.ip> "%s" .
		`, f.RespIp.StringExpanded())
		oMu := &api.Mutation{
			SetNquads: []byte(origMutations),
			Cond:      `@if(eq(len(Orig), 0))`,
		}
		rMu := &api.Mutation{
			SetNquads: []byte(respMutations),
			Cond:      `@if(eq(len(Resp), 0))`,
		}
		fMu.SetNquads = []byte(flowMutations)
		req.Mutations = []*api.Mutation{fMu, oMu, rMu}
	}

	req.Query = query
	return req
}
