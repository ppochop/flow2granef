package dgraphhelpers

import (
	"fmt"

	"github.com/dgraph-io/dgo/v240/protos/api"
	"github.com/ppochop/flow2granef/flowutils"
)

func BuildDnsTxn(d *flowutils.DNSRec, flowXid string, dnsXid string) *api.Request {
	query := fmt.Sprintf(`
		query {
			Host as var(func: eq(Host.ip, "%s"))
			Hostname as var(func: eq(Hostname.name, "%s"))
			Flow as var(func: eq(FlowRec.id, "%s"))
			Dns as var(func: eq(DNS.xid, "%s"))
		}
	`, d.Answer.StringExpanded(), *d.Query, flowXid, dnsXid)
	dnsMutations := fmt.Sprintf(`
		uid(Dns) <dgraph.type> "DNS" .
		uid(Dns) <DNS.xid> "%s" .
		uid(Dns) <DNS.trans_id> "%d" .
		uid(Dns) <DNS.query> uid(Hostname) .
		uid(Dns) <DNS.answer> uid(Host) .
	`, dnsXid, *d.TransId)
	hostMutations := `uid(Host) <Host.hostname> uid(Hostname) .`
	flowMutations := `uid(Flow) <FlowRec.produced> uid(Dns) .`
	dMu := &api.Mutation{CommitNow: true}
	hMu := &api.Mutation{CommitNow: true}
	fMu := &api.Mutation{CommitNow: true}
	req := &api.Request{CommitNow: true}
	req.Query = query
	dMu.SetNquads = []byte(dnsMutations)
	hMu.SetNquads = []byte(hostMutations)
	fMu.SetNquads = []byte(flowMutations)
	req.Mutations = []*api.Mutation{dMu, hMu, fMu}
	return req
}
