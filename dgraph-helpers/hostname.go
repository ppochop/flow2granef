package dgraphhelpers

import (
	"fmt"

	"github.com/dgraph-io/dgo/v240/protos/api"
)

func buildHostnameTxn(hostname *string) *api.Request {
	query := fmt.Sprintf(`
		query {
			Hostname as var(func: eq(Hostname.name, "%s"))
		}
	`, *hostname)
	hostnameMutation := fmt.Sprintf(`
		uid(Hostname) <dgraph.type> "Hostname" .
		uid(Hostname) <Hostname.name> "%s" .
	`, *hostname)
	hM := &api.Mutation{
		CommitNow: true,
		SetNquads: []byte(hostnameMutation),
		Cond:      `@if(eq(len(Hostname), 0))`,
	}
	req := &api.Request{CommitNow: true}
	req.Query = query
	req.Mutations = []*api.Mutation{hM}
	return req
}
