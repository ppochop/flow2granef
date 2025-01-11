package dgraphhelpers

import (
	"context"
	"log/slog"
	"net/netip"
	"time"

	"github.com/dgraph-io/dgo/v240"
	"github.com/dgraph-io/dgo/v240/protos/api"
	"github.com/ppochop/flow2granef/flowutils"
	"github.com/ppochop/flow2granef/profiles"
	"github.com/prometheus/client_golang/prometheus"
)

func AttemptTxn(ctx context.Context, dC *dgo.Dgraph, req *api.Request, softfailCtr prometheus.Counter) error {
	attempts := 0
	var err error
	for i := attempts; i < 10; i++ {
		txn := dC.NewTxn()
		defer txn.Discard(ctx)
		_, err = txn.Do(ctx, req)
		if err == nil {
			return nil
		}
		softfailCtr.Inc()
		time.Sleep(time.Millisecond * 10 * time.Duration(i))
	}
	return err
}

func AttemptHostTxn(ctx context.Context, dC *dgo.Dgraph, ip1 *netip.Addr, ip2 *netip.Addr, softfailCtr prometheus.Counter) error {
	req := buildIpTxn(ip1, ip2)
	return AttemptTxn(ctx, dC, req, softfailCtr)
}

func AttemptFlowRecTxn(ctx context.Context, dC *dgo.Dgraph, f *flowutils.FlowRec, xid string, cacheHit bool, softfailCtr prometheus.Counter) error {
	req := buildFlowRecTxn(f, xid, cacheHit)
	return AttemptTxn(ctx, dC, req, softfailCtr)
}

func HandleFlow(ctx context.Context, dC *dgo.Dgraph, f *flowutils.FlowRec, xid string, cacheHit bool, stats *profiles.TransformerStats) {
	err := AttemptHostTxn(ctx, dC, f.OrigIp, f.RespIp, stats.SoftfailedTxnHosts)
	if err != nil {
		stats.HardfailedTxnHosts.Inc()
		slog.Error("Failed to add Hosts", "err", err, "orig_ip", f.OrigIp, "recv_ip", f.RespIp)
	}

	err = AttemptFlowRecTxn(ctx, dC, f, xid, cacheHit, stats.SoftfailedTxnFlows)
	if err != nil {
		stats.HardfailedTxnFlows.Inc()
		slog.Error("Failed to add flow", "err", err, "flow", f)
	}
	stats.EventsTransformed.Inc()
}
