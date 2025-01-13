package main

import (
	"fmt"

	"github.com/ppochop/flow2granef/input"
	"github.com/ppochop/flow2granef/profiles"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func createCounterVec(name string, help string, labels []string) *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
}

func createInputCounterVec(name string, help string) *prometheus.CounterVec {
	return createCounterVec(name, help, []string{"inputter"})
}

func createTransformCounterVec(name string, help string) *prometheus.CounterVec {
	return createCounterVec(name, help, []string{"transformer", "thread"})
}

func CreateInputStats(inputs map[string]SourceConfig) map[string]input.InputStats {
	ret := map[string]input.InputStats{}
	msgsConsumed := createInputCounterVec(
		"input_msg_consumed_total",
		"Number of consumed messages by the inputter",
	)
	for inputter := range inputs {
		ret[inputter] = input.InputStats{
			MsgsConsumed: msgsConsumed.WithLabelValues(inputter),
		}
	}
	return ret
}

func CreateTransformerStats(inputs map[string]SourceConfig) map[string][]profiles.TransformerStats {
	ret := map[string][]profiles.TransformerStats{}
	eventsProcessed := createTransformCounterVec(
		"transform_events_processed",
		"Number of events processed by the transformer",
	)
	eventsTransformed := createTransformCounterVec(
		"transform_events_transformed",
		"Number of events transformed (successful transactions) by the transformer",
	)
	flowsAdded := createTransformCounterVec(
		"transform_flows_added",
		"Number of added flows by the transformer",
	)
	dnsAdded := createTransformCounterVec(
		"transform_dns_added",
		"Number of added dns by the transformer",
	)
	httpAdded := createTransformCounterVec(
		"transform_http_added",
		"Number of added http by the transformer",
	)
	softfailedTxnFlows := createTransformCounterVec(
		"transform_softfailedtxn_flows",
		"Number of softfailed (to be retried) transactions of flows by the transformer",
	)
	softfailedTxnDns := createTransformCounterVec(
		"transform_softfailedtxn_dns",
		"Number of softfailed (to be retried) transactions of dns by the transformer",
	)
	softfailedTxnHttp := createTransformCounterVec(
		"transform_softfailedtxn_http",
		"Number of softfailed (to be retried) transactions of http by the transformer",
	)
	softfailedTxnHosts := createTransformCounterVec(
		"transform_softfailedtxn_hosts",
		"Number of softfailed (to be retried) transactions of hosts by the transformer",
	)
	softfailedTxnHostname := createTransformCounterVec(
		"transform_softfailedtxn_hostname",
		"Number of softfailed (inconsequential) transactions of hostname by the transformer",
	)
	softfailedTxnUserAgent := createTransformCounterVec(
		"transform_softfailedtxn_useragent",
		"Number of softfailed (inconsequential) transactions of user_agent by the transformer",
	)
	hardfailedTxnFlows := createTransformCounterVec(
		"transform_hardfailedtxn_flows",
		"Number of hardfailed (given up) transactions of flows by the transformer",
	)
	hardfailedTxnDns := createTransformCounterVec(
		"transform_hardfailedtxn_dns",
		"Number of hardfailed (given up) transactions of dns by the transformer",
	)
	hardfailedTxnHttp := createTransformCounterVec(
		"transform_hardfailedtxn_http",
		"Number of hardfailed (given up) transactions of http by the transformer",
	)
	repeatedTxnHosts := createTransformCounterVec(
		"transform_repeatedtxn_hosts",
		"Number of hardfailed (given up) transactions of hosts by the transformer",
	)

	for key, source := range inputs {
		for i := 0; i < int(source.WorkersNum); i++ {
			thread_id := fmt.Sprintf("#%d", i)
			ret[key] = append(ret[key], profiles.TransformerStats{
				EventsProcessed:        eventsProcessed.WithLabelValues(key, thread_id),
				EventsTransformed:      eventsTransformed.WithLabelValues(key, thread_id),
				FlowsAdded:             flowsAdded.WithLabelValues(key, thread_id),
				DnsAdded:               dnsAdded.WithLabelValues(key, thread_id),
				HttpAdded:              httpAdded.WithLabelValues(key, thread_id),
				SoftfailedTxnFlows:     softfailedTxnFlows.WithLabelValues(key, thread_id),
				SoftfailedTxnDns:       softfailedTxnDns.WithLabelValues(key, thread_id),
				SoftfailedTxnHttp:      softfailedTxnHttp.WithLabelValues(key, thread_id),
				SoftfailedTxnHosts:     softfailedTxnHosts.WithLabelValues(key, thread_id),
				SoftfailedTxnHostname:  softfailedTxnHostname.WithLabelValues(key, thread_id),
				SoftfailedTxnUserAgent: softfailedTxnUserAgent.WithLabelValues(key, thread_id),
				HardfailedTxnFlows:     hardfailedTxnFlows.WithLabelValues(key, thread_id),
				HardfailedTxnDns:       hardfailedTxnDns.WithLabelValues(key, thread_id),
				HardfailedTxnHttp:      hardfailedTxnHttp.WithLabelValues(key, thread_id),
				RepeatedTxnHosts:       repeatedTxnHosts.WithLabelValues(key, thread_id),
			})
		}
	}
	return ret
}
