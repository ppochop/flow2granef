package profiles

import (
	"context"
	"time"

	"github.com/dgraph-io/dgo/v240"
	"github.com/prometheus/client_golang/prometheus"
)

type Cache interface {
	Set(string, string, time.Time)
	Get(string) (string, bool)
	AddOrGet(commId string, xid string, lastTs time.Time) (string, bool)
}

type CacheDuplCheck interface {
	DuplHandle(string, time.Time, time.Time, string) (string, bool)
}

type TransformerStats struct {
	EventsProcessed    prometheus.Counter
	EventsTransformed  prometheus.Counter
	SoftfailedTxnFlows prometheus.Counter
	SoftfailedTxnHosts prometheus.Counter
	HardfailedTxnFlows prometheus.Counter
	HardfailedTxnHosts prometheus.Counter
}

type TransformerFactory func(Cache, *dgo.Dgraph, TransformerStats) Transformer
type TransformerDuplCheckFactory func(CacheDuplCheck, string) Transformer

type Transformer interface {
	Handle(ctx context.Context, flow []byte) error
}
