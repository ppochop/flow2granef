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

/*
var flowSource string
var workersNum uint
var in string
var passiveTimeout time.Duration
var duplCheck bool
*/
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-c
		defer cancel()
	}()

	duplCache := xidcache.NewDuplCache(30 * time.Minute)

	cache := xidcache.New(passiveTimeout.Round(time.Second))

	dgoClient := newClient(mC.DgraphAddress)

	if mC.ResetDgraph {
		reset(ctx, dgoClient)
		setup(ctx, dgoClient)
	}

	inputters := map[string]input.Input{}
	transformers := map[string]profiles.Transformer{}
	var wg sync.WaitGroup

	inputsStats := CreateInputStats(mC.Sources)
	transformersStats := CreateTransformerStats(mC.Sources)
	for key, source := range mC.Sources {
		inputConfig := source.InputConfig
		inputStats := inputsStats[key]
		inputter, err := input.GetInput(source.InputName, inputConfig, inputStats)
		if err != nil {
			slog.Error("Error when getting the inputter.", "error", err)
			os.Exit(1)
		}
		inputters[key] = inputter

		preHandler, err := profiles.GetPreHandler(source.TransformerName)
		if err != nil {
			slog.Error("Error when getting the prehandler.", "error", err)
			os.Exit(1)
		}

		channels := [](chan []byte){}

		for i := 0; i < int(source.WorkersNum); i++ {
			var transformer profiles.Transformer
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
			flowsChannel := make(chan []byte, 256)
			channels = append(channels, flowsChannel)
			transformers[fmt.Sprintf("%s:#%d", key, i)] = transformer

			wg.Add(1)
			slog.Info(fmt.Sprintf("Spawning worker %s:#%d", key, i))

			go func(key string, index int) {
				defer func() {
					slog.Info(fmt.Sprintf("Transform worker %s:#%d finished", key, index))
					wg.Done()
				}()
				spawnTransformer(ctx, transformer, flowsChannel)
			}(key, i)
		}

		wg.Add(1)
		go func(key string) {
			defer func() {
				slog.Info(fmt.Sprintf("Input worker %s finished", key))
				wg.Done()
			}()
			spawnInputter(ctx, inputter, preHandler, channels, uint32(source.WorkersNum))
		}(key)

	}

	wg.Add(1)
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
