package dgraphhelpers

import (
	"fmt"

	"github.com/dgraph-io/dgo/v240/protos/api"
	"github.com/ppochop/flow2granef/flowutils"
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
		CommitNow: true,
		SetNquads: []byte(mutation),
	}
	return &api.Request{
		Query:     query,
		CommitNow: true,
		Mutations: []*api.Mutation{mut},
	}
}

func buildFlowRecTxn(f *flowutils.FlowRec, xid string, cacheHit bool) *api.Request {
	var query string
	var flowMutations string
	if cacheHit {
		query = fmt.Sprintf(`
			query {
				Flow as var(func: eq(FlowRec.id, "%s"))
			}
		`, xid)
		flowMutations = fmt.Sprintf(`
			uid(Flow) <dgraph.type> "FlowRec" .
			uid(Flow) <FlowRec.last_ts> "%s" .
			uid(Flow) <FlowRec.flush_reason> "%s" .
			uid(Flow) <FlowRec.flow_source> "%s" .
		`, f.LastTs, f.FlushReason, f.FlowSource)
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
			uid(Flow) <FlowRec.first_ts> "%s" .
			uid(Flow) <FlowRec.last_ts> "%s" .
			uid(Flow) <FlowRec.protocol> "%s" .
			uid(Flow) <FlowRec.app> "%s" .
			uid(Flow) <FlowRec.flush_reason> "%s" .
			uid(Flow) <FlowRec.flow_source> "%s" .
		`, xid, f.CommId, f.OrigPort, f.RespPort, f.FirstTs, f.LastTs, f.Protocol.GetName(), f.App, f.FlushReason, f.FlowSource)
	}
	fMu := &api.Mutation{CommitNow: true}
	req := &api.Request{CommitNow: true}
	req.Query = query
	fMu.SetNquads = []byte(flowMutations)
	req.Mutations = []*api.Mutation{fMu}
	return req
}
