package dgraphhelpers

import (
	"fmt"
	"net/netip"

	"github.com/dgraph-io/dgo/v240/protos/api"
)

func buildIpsTxn(srcIp *netip.Addr, destIp *netip.Addr) *api.Request {
	query := fmt.Sprintf(`
		query {
			Orig as var(func: eq(Host.ip, "%s"))
			Resp as var(func: eq(Host.ip, "%s"))
		}
	`, srcIp.StringExpanded(), destIp.StringExpanded())
	hostMutationSrc := fmt.Sprintf(`
		uid(Orig) <dgraph.type> "Host" .
		uid(Orig) <Host.ip> "%s" .
	`, srcIp.StringExpanded())
	hMS := &api.Mutation{
		CommitNow: true,
		SetNquads: []byte(hostMutationSrc),
		Cond:      `@if(eq(len(Orig), 0))`,
	}
	hostMutationDest := fmt.Sprintf(`
		uid(Resp) <dgraph.type> "Host" .
		uid(Resp) <Host.ip> "%s" .
	`, destIp.StringExpanded())
	hMD := &api.Mutation{
		CommitNow: true,
		SetNquads: []byte(hostMutationDest),
		Cond:      `@if(eq(len(Resp), 0))`,
	}
	req := &api.Request{CommitNow: true}
	req.Query = query
	req.Mutations = []*api.Mutation{hMS, hMD}
	return req
}

func buildIpTxn(ip *netip.Addr) *api.Request {
	query := fmt.Sprintf(`
		query {
			Host as var(func: eq(Host.ip, "%s"))
		}
	`, ip.StringExpanded())
	hostMutation := fmt.Sprintf(`
		uid(Host) <dgraph.type> "Host" .
		uid(Host) <Host.ip> "%s" .
	`, ip.StringExpanded())
	hM := &api.Mutation{
		CommitNow: true,
		SetNquads: []byte(hostMutation),
		Cond:      `@if(eq(len(Host), 0))`,
	}
	req := &api.Request{CommitNow: true}
	req.Query = query
	req.Mutations = []*api.Mutation{hM}
	return req
}
