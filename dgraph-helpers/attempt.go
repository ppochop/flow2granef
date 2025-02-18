package dgraphhelpers

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/netip"
	"sync"
	"time"

	"github.com/dgraph-io/dgo/v240"
	"github.com/dgraph-io/dgo/v240/protos/api"
	"github.com/ppochop/flow2granef/flowutils"
	"github.com/ppochop/flow2granef/profiles"
	xidcache "github.com/ppochop/flow2granef/xid-cache"
	"github.com/prometheus/client_golang/prometheus"
)

func AttemptTxn(ctx context.Context, dC *dgo.Dgraph, req *api.Request, incCtr bool, softfailCtr prometheus.Counter, attempts int) error {
	var err error
	for i := 1; i <= attempts; i++ {
		if i != 1 {
			time.Sleep(time.Millisecond * time.Duration(i*10*(1+rand.Intn(10))))
		}
		txn := dC.NewTxn()
		defer txn.Discard(ctx)
		_, err = txn.Do(ctx, req)
		if err == nil {
			return nil
		}
		if incCtr {
			softfailCtr.Inc()
		}
	}
	return err
}

func AttemptHostsTxn(ctx context.Context, dC *dgo.Dgraph, ip1 *netip.Addr, ip2 *netip.Addr, softfailCtr prometheus.Counter) error {
	// Upsert both hosts in one transaction
	req0 := buildIpsTxn(ip1, ip2)
	err := AttemptTxn(ctx, dC, req0, true, softfailCtr, 1)
	if err != nil {
		// At least one upsert failed which means that at least one of the two hosts already exists.
		// In case only one of them exists, the other one's creation was aborted with the transaction
		// We have no way to know what the case is so the sane thing is to retry the upsert for each
		// host individually
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			req1 := buildIpTxn(ip1)
			AttemptTxn(ctx, dC, req1, false, softfailCtr, 3)
		}()
		go func() {
			defer wg.Done()
			req2 := buildIpTxn(ip2)
			AttemptTxn(ctx, dC, req2, false, softfailCtr, 3)
		}()
		wg.Wait()
		return fmt.Errorf("hosts upsert retried")
	}
	return nil
}

func AttemptFlowRecTxn(ctx context.Context, dC *dgo.Dgraph, f *flowutils.FlowRec, xid string, cacheHit xidcache.CacheHitResult, softfailCtr prometheus.Counter) error {
	req := buildFlowRecTxn(f, xid, cacheHit)
	return AttemptTxn(ctx, dC, req, true, softfailCtr, 10)
}

func HandleFlow(ctx context.Context, dC *dgo.Dgraph, f *flowutils.FlowRec, xid string, cacheHit xidcache.CacheHitResult, stats *profiles.TransformerStats) error {
	if cacheHit != xidcache.Hit { // Hosts may not exist
		// We don't care about failures here as they mean the host already exists
		err := AttemptHostsTxn(ctx, dC, f.OrigIp, f.RespIp, stats.SoftfailedTxnHosts)
		if err != nil {
			stats.RepeatedTxnHosts.Inc()
		}
	}

	err := AttemptFlowRecTxn(ctx, dC, f, xid, cacheHit, stats.SoftfailedTxnFlows)
	if err != nil {
		stats.HardfailedTxnFlows.Inc()
		slog.Error("Flow attempt failed", "err", err, "cache_hit", cacheHit, "xid", xid, "flow", f)
		return err
	}
	stats.FlowsAdded.Inc()
	return nil
}

func HandleDns(ctx context.Context, dC *dgo.Dgraph, d *flowutils.DNSRec, flowXid string, stats *profiles.TransformerStats) error {
	var wg sync.WaitGroup
	wg.Add(1 + len(d.Answer))

	go func() {
		defer wg.Done()
		reqHostname := buildHostnameTxn(d.Query)
		AttemptTxn(ctx, dC, reqHostname, true, stats.SoftfailedTxnHostname, 1)
	}()

	for _, ip := range d.Answer {
		go func() {
			defer wg.Done()
			reqHost := buildIpTxn(ip)
			AttemptTxn(ctx, dC, reqHost, true, stats.SoftfailedTxnHosts, 1)
		}()
	}
	dnsXid := fmt.Sprintf("%s%d", flowXid, *d.TransId)
	reqDns := BuildDnsTxn(d, flowXid, dnsXid)
	wg.Wait()
	err := AttemptTxn(ctx, dC, reqDns, true, stats.SoftfailedTxnDns, 10)
	if err != nil {
		stats.HardfailedTxnDns.Inc()
		slog.Error("Dns attempt failed", "err", err, "dns", d)
		return err
	}
	stats.DnsAdded.Inc()
	return nil
}

func HandleDnsWithFlowPlaceholder(ctx context.Context, dC *dgo.Dgraph, d *flowutils.DNSRec, flowXid string, stats *profiles.TransformerStats) error {
	reqFlowPlaceholder := buildFlowRecPlaceholderTxn(flowXid)
	AttemptTxn(ctx, dC, reqFlowPlaceholder, true, stats.SoftfailedTxnFlows, 1)

	return HandleDns(ctx, dC, d, flowXid, stats)
}

func HandleHttp(ctx context.Context, dC *dgo.Dgraph, h *flowutils.HTTPRec, flowXid string, stats *profiles.TransformerStats) error {
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		reqHostname := buildHostnameTxn(h.Hostname)
		AttemptTxn(ctx, dC, reqHostname, true, stats.SoftfailedTxnHostname, 1)
	}()

	go func() {
		defer wg.Done()
		reqUA := buildUserAgentTxn(h.UserAgent)
		AttemptTxn(ctx, dC, reqUA, true, stats.SoftfailedTxnUserAgent, 1)
	}()

	go func() {
		defer wg.Done()
		reqHostsEdges := buildHttpHostsEdges(h)
		AttemptTxn(ctx, dC, reqHostsEdges, false, stats.SoftfailedTxnHosts, 1)
	}()

	wg.Wait()
	url, path := handleUrl(h.Url)
	reqHTTP := buildHttpTxn(h, flowXid, url, path)
	err := AttemptTxn(ctx, dC, reqHTTP, true, stats.SoftfailedTxnHttp, 10)
	if err != nil {
		stats.HardfailedTxnHttp.Inc()
		slog.Error("HTTP attempt failed", "err", err, "http_ua", *h.UserAgent, "hostname", *h.Hostname, "url", *h.Url)
		return err
	}
	stats.HttpAdded.Inc()
	return nil
}

func HandleHttpWithFlowPlaceholder(ctx context.Context, dC *dgo.Dgraph, h *flowutils.HTTPRec, flowXid string, stats *profiles.TransformerStats) error {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		reqFlowPlaceholder := buildFlowRecPlaceholderTxn(flowXid)
		AttemptTxn(ctx, dC, reqFlowPlaceholder, true, stats.SoftfailedTxnFlows, 1)
	}()

	go func() {
		defer wg.Done()
		AttemptHostsTxn(ctx, dC, h.ClientIp, h.ServerIp, stats.SoftfailedTxnHosts)
	}()
	wg.Wait()
	return HandleHttp(ctx, dC, h, flowXid, stats)
}
