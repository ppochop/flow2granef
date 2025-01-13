package profiles

import (
	"context"
	"time"

	"github.com/dgraph-io/dgo/v240"
	"github.com/prometheus/client_golang/prometheus"
)

type Cache interface {
	Set(commId string, xid string, lastTs time.Time)
	Get(commId string, firstTs time.Time) (string, bool)
	AddOrGet(commId string, xid string, firstTs time.Time, lastTs time.Time) (string, bool)
}

type CacheDuplCheck interface {
	DuplHandle(string, time.Time, time.Time, string) (string, bool)
}

type TransformerStats struct {
	EventsProcessed        prometheus.Counter
	EventsTransformed      prometheus.Counter
	FlowsAdded             prometheus.Counter
	DnsAdded               prometheus.Counter
	HttpAdded              prometheus.Counter
	SoftfailedTxnFlows     prometheus.Counter
	SoftfailedTxnDns       prometheus.Counter
	SoftfailedTxnHttp      prometheus.Counter
	SoftfailedTxnHosts     prometheus.Counter
	SoftfailedTxnHostname  prometheus.Counter
	SoftfailedTxnUserAgent prometheus.Counter
	HardfailedTxnFlows     prometheus.Counter
	HardfailedTxnDns       prometheus.Counter
	HardfailedTxnHttp      prometheus.Counter
	RepeatedTxnHosts       prometheus.Counter
}

type TransformerFactory func(Cache, *dgo.Dgraph, TransformerStats) Transformer
type TransformerDuplCheckFactory func(CacheDuplCheck, string) Transformer

type Transformer interface {
	Handle(ctx context.Context, flow []byte) error
}
