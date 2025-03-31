/*
Flow2granef ingests network flows and related network security monitoring events into a Dgraph (graph) database.

Usage:

	flow2granef --config=config.toml
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ppochop/flow2granef/input"
	"github.com/ppochop/flow2granef/profiles"

	_ "github.com/ppochop/flow2granef/profiles/ipfix"
	_ "github.com/ppochop/flow2granef/profiles/suricata"
	_ "github.com/ppochop/flow2granef/profiles/zeek"
	xidcache "github.com/ppochop/flow2granef/xid-cache"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "/home/ppochop/school/theses/msc/tools/flow2granef/config2.toml", "Path to the config.")
}

func main() {
	flag.Parse()

	mC, err := ReadConfigFile(configPath)
	if err != nil {
		slog.Error("Error while parsing config.", "error", err)
		os.Exit(1)
	}

	passiveTimeout, err := time.ParseDuration(mC.PassiveTimeout)
	if err != nil {
		slog.Error("Error while parsing passive timeout.", "error", err)
		os.Exit(1)
	}

	// setup Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-c
		defer cancel()
	}()

	// setup caches
	duplCache := xidcache.NewDuplCache(30 * time.Minute)
	cache := xidcache.New(passiveTimeout.Round(time.Second))

	// connect to dgraph
	dgoClient := newClient(mC.DgraphAddress)

	if mC.ResetDgraph {
		reset(ctx, dgoClient)
		setup(ctx, dgoClient)
	}

	inputters := map[string]input.Input{}
	transformers := map[string]profiles.Transformer{}
	var wg sync.WaitGroup

	// prepare prometheus metric update calls
	inputsStats := CreateInputStats(mC.Sources)
	transformersStats := CreateTransformerStats(mC.Sources)

	// configure an input-reading worker for each source in the config
	for key, source := range mC.Sources {

		inputConfig := source.InputConfig
		inputStats := inputsStats[key]
		inputter, err := input.GetInput(source.InputName, inputConfig, inputStats)
		if err != nil {
			slog.Error("Error when getting the inputter.", "error", err)
			os.Exit(1)
		}
		inputters[key] = inputter

		/*
			preHandler, err := profiles.GetPreHandler(source.TransformerName)
			if err != nil {
				slog.Error("Error when getting the prehandler.", "error", err)
				os.Exit(1)
			}

			channels := [](chan []byte){}
		*/
		flows := make(chan []byte, 64) // the channel between an inputter and related transformers
		// configure the required number of transforming workers
		for i := 0; i < int(source.WorkersNum); i++ {
			var transformer profiles.Transformer

			// select the right transformer for the configured mode of operation
			if mC.DuplCheck {
				transformer, err = profiles.GetTransformerDuplCheck(source.TransformerName, duplCache, key)
			} else {
				dgoClient = newClient(mC.DgraphAddress)
				stats := transformersStats[key][i]
				transformer, err = profiles.GetTransformer(source.TransformerName, cache, dgoClient, stats)
			}
			if err != nil {
				slog.Error("Failed to fetch the right transformer.", "error", err)
				os.Exit(1)
			}
			//flowsChannel := make(chan []byte, 256)
			//channels = append(channels, flowsChannel)
			transformers[fmt.Sprintf("%s:#%d", key, i)] = transformer

			wg.Add(1)
			slog.Info(fmt.Sprintf("Spawning transform worker %s:#%d", key, i))

			go func(key string, index int) {
				defer func() {
					slog.Info(fmt.Sprintf("Transform worker %s:#%d finished", key, index))
					wg.Done()
				}()
				spawnTransformer(ctx, transformer, flows)
			}(key, i)
		}

		wg.Add(1)
		slog.Info(fmt.Sprintf("Spawning input worker %s", key))

		go func(key string) {
			defer func() {
				slog.Info(fmt.Sprintf("Input worker %s finished", key))
				wg.Done()
			}()
			//spawnInputter(ctx, inputter, preHandler, channels, uint32(source.WorkersNum))
			spawnSimpleInputter(ctx, inputter, flows)
		}(key)

	}

	wg.Add(1)
	slog.Info("Spawning stats worker")

	go func() {
		defer func() {
			slog.Info("Stats worker finished.")
			wg.Done()
		}()
		spawnStatsWorker(ctx)
	}()

	wg.Wait()
	slog.Info("All workers finished")
}
