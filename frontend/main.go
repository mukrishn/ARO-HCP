package main

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
)

const ProgramName = "ARO HCP Frontend"

func main() {
	version := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				version = setting.Value
				break
			}
		}
	}
	logger := DefaultLogger()

	logger.Info(fmt.Sprintf("%s (%s) started", ProgramName, version))

	// Fetch the region from the env variable
	region := os.Getenv("REGION")
	if region == "" {
		logger.Error("REGION env variable is not set.")
	}
	logger.Info(fmt.Sprintf("Application running in region: %s", region))

	ctx := context.Background()
	stop := make(chan struct{})

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	listener, err := net.Listen("tcp4", ":8443")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Init prometheus emitter
	prometheusEmitter := NewPrometheusEmitter()

	// Configure database configuration and client
	dbConfig := NewDatabaseConfig()

	dbClient, err := NewDatabaseClient(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Creating the database client failed: %v", err))
	}

	frontend := NewFrontend(logger, listener, prometheusEmitter, dbClient, region)

	// Verify the Async DB is available and accessible
	logger.Info("Testing DB Access")
	result, err := frontend.dbClient.DBConnectionTest(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("Database test failed to fetch properties: %v", err))
	} else {
		logger.Info(fmt.Sprintf("Database check completed - %s", result))
	}

	go frontend.Run(ctx, stop)

	sig := <-signalChannel
	logger.Info(fmt.Sprintf("caught %s signal", sig))
	close(stop)
	frontend.Join()

	logger.Info(fmt.Sprintf("%s (%s) stopped", ProgramName, version))
}
