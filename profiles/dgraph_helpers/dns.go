package dgraphhelpers

import (
	"fmt"
	"strings"

	"github.com/dgraph-io/dgo/v240/protos/api"
	"github.com/ppochop/flow2granef/flowutils"
)

func buildDnsAux(d *flowutils.DNSRec) (qAux string, dMAux string, hMAux string) {
	hostQueryAux := make([]string, len(d.Answer))
	dnsMutAux := make([]string, len(d.Answer))
	hostMutAux := make([]string, len(d.Answer))
	for i, ip := range d.Answer {
		hostQueryAux[i] = fmt.Sprintf("\t\t\tHost%d as var(func: eq(Host.ip, \"%s\"))", i, ip.StringExpanded())
		dnsMutAux[i] = fmt.Sprintf("\t\tuid(Dns) <DNS.answer> uid(Host%d) .", i)
		hostMutAux[i] = fmt.Sprintf("uid(Host%d) <Host.hostname> uid(Hostname) .", i)
	}
	ret0 := strings.Join(hostQueryAux, "\n")
	ret1 := strings.Join(dnsMutAux, "\n")
	ret2 := strings.Join(hostMutAux, "\n")
	return ret0, ret1, ret2
}

func BuildDnsTxn(d *flowutils.DNSRec, flowXid string, dnsXid string) *api.Request {
	qAux, dMAux, hMAux := buildDnsAux(d)
	query := fmt.Sprintf(`
		query {
			Hostname as var(func: eq(Hostname.name, "%s"))
			Flow as var(func: eq(FlowRec.id, "%s"))
			Dns as var(func: eq(DNS.xid, "%s"))
%s
		}
	`, *d.Query, flowXid, dnsXid, qAux)
	dnsMutations := fmt.Sprintf(`
		uid(Dns) <dgraph.type> "DNS" .
		uid(Dns) <DNS.xid> "%s" .
		uid(Dns) <DNS.trans_id> "%d" .
		uid(Dns) <DNS.qtype_name> "%s" .
		uid(Dns) <DNS.query> uid(Hostname) .
%s
	`, dnsXid, *d.TransId, *d.QType, dMAux)
	hostMutations := hMAux
	flowMutations := `uid(Flow) <FlowRec.produced> uid(Dns) .`
	dMu := &api.Mutation{
		SetNquads: []byte(dnsMutations),
	}
	hMu := &api.Mutation{
		SetNquads: []byte(hostMutations),
	}
	fMu := &api.Mutation{
		SetNquads: []byte(flowMutations),
	}
	req := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{dMu, hMu, fMu},
		CommitNow: true,
	}
	return req
}
