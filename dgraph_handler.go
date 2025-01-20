package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/dgraph-io/dgo/v240"
	"github.com/dgraph-io/dgo/v240/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newClient(address string) *dgo.Dgraph {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)
}

func reset(ctx context.Context, c *dgo.Dgraph) {
	err := c.Alter(ctx, &api.Operation{DropOp: api.Operation_ALL})
	if err != nil {
		slog.Error("Failed to reset db", "err", err)
		panic(1)
	}
}

func setup(ctx context.Context, c *dgo.Dgraph) {
	err := c.Alter(ctx, &api.Operation{
		Schema: `
Host.ip: string @upsert @index(cidr, exact) .
Host.hostname: [uid] @reverse .
Host.user_agent: [uid] @reverse .


FlowRec.id: string @upsert @index(exact) .
FlowRec.community_id: string @index(hash) .
FlowRec.originated_by: uid @reverse .
FlowRec.received_by: uid @reverse .
FlowRec.orig_port: int .
FlowRec.recv_port: int .
FlowRec.from_orig_bytes: int .
FlowRec.from_recv_bytes: int .
FlowRec.from_orig_pkts: int .
FlowRec.from_recv_pkts: int .
FlowRec.first_ts: dateTime @index(hour) .
FlowRec.last_ts: dateTime @index(hour) .
FlowRec.protocol: string @index(hash) .
FlowRec.app: string @index(hash) .
FlowRec.flush_reason: string @index(hash) .
FlowRec.flow_source: string @index(hash) .
FlowRec.produced: [uid] @reverse .

DNS.xid: string @upsert @index(exact) .
DNS.trans_id: int @index(int) .
DNS.query: [uid] @count @reverse .
DNS.answer: [uid] @count @reverse .

HTTP.url: string @index(hash, trigram) .
HTTP.path: string @index(hash) .
HTTP.hostname: [uid] @reverse .
HTTP.user_agent: [uid] @reverse .

UserAgent.user_agent: string @upsert @index(exact) .

Hostname.name: string @upsert @index(exact) .

type Host {
    Host.ip
    Host.user_agent
    Host.hostname
    <~DNS.answer>
    <~FlowRec.originated_by>
    <~FlowRec.received_by>
}

type FlowRec {
    FlowRec.id
    FlowRec.community_id
    FlowRec.originated_by
    FlowRec.received_by
    FlowRec.produced
    FlowRec.orig_port
    FlowRec.recv_port
    FlowRec.from_orig_bytes
    FlowRec.from_recv_bytes
    FlowRec.from_orig_pkts
    FlowRec.from_recv_pkts
    FlowRec.first_ts
    FlowRec.last_ts
    FlowRec.protocol
    FlowRec.app
    FlowRec.flush_reason
    FlowRec.flow_source
}

type HTTP {
    <~FlowRec.produced>
    HTTP.user_agent
    HTTP.hostname
    HTTP.url
    HTTP.path
}

type DNS {
    DNS.xid
    <~FlowRec.produced>
    DNS.trans_id
    DNS.query
    DNS.answer
}

type UserAgent {
    <~Host.user_agent>
    <~HTTP.user_agent>
    UserAgent.user_agent
}

type Hostname {
    <~Host.hostname>
    <~HTTP.hostname>
    <~DNS.query>
    Hostname.name
}
		`,
	})
	if err != nil {
		slog.Error("Failed to update schema", "err", err)
		panic(1)
	}
}
