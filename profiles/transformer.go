// Package profiles provides handlers for supported sources of network security monitoring events.
package profiles

import (
	"context"
	"time"

	"github.com/dgraph-io/dgo/v240"
	xidcache "github.com/ppochop/flow2granef/xid-cache"
	"github.com/prometheus/client_golang/prometheus"
)

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
	ProcessingTimeFlow     prometheus.Observer
	ProcessingTimeHttp     prometheus.Observer
	ProcessingTimeDns      prometheus.Observer
}

type TransformerFactory func(*xidcache.IdCache, *dgo.Dgraph, TransformerStats) Transformer
type TransformerDuplCheckFactory func(*xidcache.DuplCache, string) Transformer

type Transformer interface {
	Handle(ctx context.Context, flow []byte) error
}

func TimeTrack(start time.Time, obs prometheus.Observer) {
	elapsed := time.Since(start).Milliseconds()
	obs.Observe(float64(elapsed))
}
