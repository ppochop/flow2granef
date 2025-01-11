package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ppochop/flow2granef/input"
	"github.com/ppochop/flow2granef/profiles"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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

func sendToTransform(ctx context.Context, input []byte, preHandler profiles.PreHandler, channels []chan []byte, workers_num uint32) {
	worker_preid, err := preHandler(input)
	if err != nil {
		slog.Error("error during attempting to get ip pair id", "error", err)
		return
	}
	worker_id := worker_preid % workers_num
	select {
	case <-ctx.Done():
	case channels[worker_id] <- input:
	}
}

func spawnInputter(ctx context.Context, inputter input.Input, preHandler profiles.PreHandler, channels [](chan []byte), workers_num uint32) {
	gRoutines := make(chan struct{}, workers_num*4)
	for {
		select {
		case <-ctx.Done():
			return
		case gRoutines <- struct{}{}:
			input, err := inputter.NextEntry()
			if err != nil {
				slog.Error("error during parsing input record", "error", err)
				return
			}
			if input == nil {
				continue
			}
			go func() {
				sendToTransform(ctx, input, preHandler, channels, workers_num)
				<-gRoutines
			}()
		}
	}
}

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
