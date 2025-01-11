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
	softfailedTxnFlows := createTransformCounterVec(
		"transform_softfailedtxn_flows",
		"Number of softfailed (to be retried) transactions of flows by the transformer",
	)
	softfailedTxnHosts := createTransformCounterVec(
		"transform_softfailedtxn_hosts",
		"Number of softfailed (to be retried) transactions of hosts by the transformer",
	)
	hardfailedTxnFlows := createTransformCounterVec(
		"transform_hardfailedtxn_flows",
		"Number of hardfailed (given up) transactions of flows by the transformer",
	)
	hardfailedTxnHosts := createTransformCounterVec(
		"transform_hardfailedtxn_hosts",
		"Number of hardfailed (given up) transactions of hosts by the transformer",
	)

	for key, source := range inputs {
		for i := 0; i < int(source.WorkersNum); i++ {
			thread_id := fmt.Sprintf("#%d", i)
			ret[key] = append(ret[key], profiles.TransformerStats{
				EventsProcessed:    eventsProcessed.WithLabelValues(key, thread_id),
				EventsTransformed:  eventsTransformed.WithLabelValues(key, thread_id),
				SoftfailedTxnFlows: softfailedTxnFlows.WithLabelValues(key, thread_id),
				SoftfailedTxnHosts: softfailedTxnHosts.WithLabelValues(key, thread_id),
				HardfailedTxnFlows: hardfailedTxnFlows.WithLabelValues(key, thread_id),
				HardfailedTxnHosts: hardfailedTxnHosts.WithLabelValues(key, thread_id),
			})
		}
	}
	return ret
}
