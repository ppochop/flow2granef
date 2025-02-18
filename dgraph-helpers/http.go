package dgraphhelpers

import (
	"fmt"
	"strings"

	"github.com/dgraph-io/dgo/v240/protos/api"
	"github.com/ppochop/flow2granef/flowutils"
)

var urlReplacer strings.Replacer

func init() {
	urlReplacer = *strings.NewReplacer("\n", "", "\t", "", "\r", "", `"`, `\"`, `\`, `\\`)
}

func handleUrl(url *string) (cleanUrl *string, path *string) {
	replaced := urlReplacer.Replace(*url)
	//ret := strings.SplitN(replaced, "?", 2)
	ret, _, _ := strings.Cut(replaced, "?")
	return &replaced, &ret
}

func buildHttpHostsEdges(h *flowutils.HTTPRec) *api.Request {
	query := fmt.Sprintf(`
		query {
			Hostname as var(func: eq(Hostname.name, "%s"))
			ClientHost as var(func: eq(Host.ip, "%s"))
			ServerHost as var(func: eq(Host.ip, "%s"))
			UA as var(func: eq(UserAgent.user_agent, "%s"))
		}
	`, *h.Hostname, h.ClientIp.StringExpanded(), h.ServerIp.StringExpanded(), *h.UserAgent)
	hostnameMutation := "uid(ServerHost) <Host.hostname> uid(Hostname) ."
	uaMutation := "uid(ClientHost) <Host.user_agent> uid(UA) ."
	hM := &api.Mutation{
		SetNquads: []byte(hostnameMutation),
	}
	uM := &api.Mutation{
		SetNquads: []byte(uaMutation),
	}
	cM := &api.Mutation{
		SetNquads: []byte(fmt.Sprintf(`
			uid(ClientHost) <dgraph.type> "Host" .
			uid(ClientHost) <Host.ip> "%s" .
		`, h.ClientIp.StringExpanded())),
		Cond: `@if(eq(len(ClientHost), 0))`,
	}
	sM := &api.Mutation{
		SetNquads: []byte(fmt.Sprintf(`
			uid(ServerHost) <dgraph.type> "Host" .
			uid(ServerHost) <Host.ip> "%s" .
		`, h.ServerIp.StringExpanded())),
		Cond: `@if(eq(len(ServerHost), 0))`,
	}
	req := &api.Request{
		CommitNow: true,
		Query:     query,
		Mutations: []*api.Mutation{hM, uM, cM, sM},
	}
	return req
}

func buildHttpTxn(h *flowutils.HTTPRec, flowXid string, url *string, path *string) *api.Request {
	query := fmt.Sprintf(`
		query {
			Hostname as var(func: eq(Hostname.name, "%s"))
			Flow as var(func: eq(FlowRec.id, "%s")) 
			UA as var(func: eq(UserAgent.user_agent, "%s"))
		}
	`, *h.Hostname, flowXid, *h.UserAgent)
	httpMutation := fmt.Sprintf(`
		<_:http> <dgraph.type> "HTTP" .
		<_:http> <HTTP.url> "%s" .
		<_:http> <HTTP.path> "%s" .
		<_:http> <HTTP.method> "%s" .
		<_:http> <HTTP.status_code> "%d" .
		<_:http> <HTTP.hostname> uid(Hostname) .
		<_:http> <HTTP.user_agent> uid(UA) .
		uid(Flow) <FlowRec.produced> <_:http> .
	`, *url, *path, *h.Method, h.StatusCode)

	httpM := &api.Mutation{
		SetNquads: []byte(httpMutation),
	}
	req := &api.Request{
		CommitNow: true,
		Query:     query,
		Mutations: []*api.Mutation{httpM},
	}
	return req
}
