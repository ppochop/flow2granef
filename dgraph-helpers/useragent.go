package dgraphhelpers

import (
	"fmt"

	"github.com/dgraph-io/dgo/v240/protos/api"
)

func buildUserAgentTxn(useragent *string) *api.Request {
	query := fmt.Sprintf(`
		query {
			UA as var(func: eq(UserAgent.user_agent, "%s"))
		}
	`, *useragent)
	useragentMutation := fmt.Sprintf(`
		uid(UA) <dgraph.type> "UserAgent" .
		uid(UA) <UserAgent.user_agent> "%s" .
	`, *useragent)
	uM := &api.Mutation{
		CommitNow: true,
		SetNquads: []byte(useragentMutation),
		Cond:      `@if(eq(len(UA), 0))`,
	}
	req := &api.Request{CommitNow: true}
	req.Query = query
	req.Mutations = []*api.Mutation{uM}
	return req
}
