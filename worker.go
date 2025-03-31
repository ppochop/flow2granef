package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ppochop/flow2granef/input"
	"github.com/ppochop/flow2granef/profiles"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// spawnTransformer is the working loop of a transform worker.
// It reads from the flows channel and attempts to handle each event.
func spawnTransformer(ctx context.Context, transformer profiles.Transformer, flows <-chan []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case f := <-flows:
			err := transformer.Handle(ctx, f)
			if err != nil {
				slog.Error("error during transforming the record", "error", err)
				continue
			}
		}
	}
}

// spawnSimpleInputter is the working loop of an input worker.
// It fetches encoded events from the inputter and feeds it to the flows channel.
func spawnSimpleInputter(ctx context.Context, inputter input.Input, flows chan []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			input, err := inputter.NextEntry()
			if err != nil {
				slog.Error("error during parsing input record", "error", err)
				return
			}
			if input == nil {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case flows <- input:
			}
		}
	}
}

// spawnStatsWorker is responsible for handling the Prometheus metrics endpoint.
func spawnStatsWorker(ctx context.Context) error {
	srv := http.Server{
		Addr: ":2112",
	}
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("metrics server crashed", "err", err)
		return err
	}
	return nil
}
