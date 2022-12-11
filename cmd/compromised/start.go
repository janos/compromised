// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"golang.org/x/exp/slog"
	"resenje.org/recovery"
	"resenje.org/web/logging"
	"resenje.org/web/server"
	"resenje.org/x/application"

	"resenje.org/compromised"
	"resenje.org/compromised/cmd/compromised/config"
	"resenje.org/compromised/pkg/api"
	"resenje.org/compromised/pkg/metrics"
	filepasswords "resenje.org/compromised/pkg/passwords/file"
)

func startCmd(daemon bool) error {
	if options.PasswordsDB == "" {
		fmt.Fprintln(os.Stderr, `Passwords database is not configured.

Download Pwned passwords SHA1 ordered by hash from https://haveibeenpwned.com/Passwords and execute index-database command to generate a database.

Refer to https://resenje.org/compromised documentation.`)
		return errors.New("configuration error")
	}

	// Initialize the application with loaded options.
	app, err := application.NewApp(
		config.Name,
		application.AppOptions{
			LogDir:            options.LogDir,
			PidFileName:       options.PidFileName,
			DaemonLogFileName: options.DaemonLogFileName,
			DaemonLogFileMode: options.DaemonLogFileMode.FileMode(),
		})
	if err != nil {
		return err
	}

	// Functions that will be called in parallel on application shutdown.
	var shutdownFuncs []func() error

	// Setup logging.

	loggingMetrics := logging.NewMetrics(&logging.MetricsOptions{
		Namespace: metrics.Namespace,
	})

	loggerWriter := logging.ApplicationLogWriteCloser(options.LogDir, config.Name, os.Stderr)
	defer loggerWriter.Close()
	logger := slog.New(slog.HandlerOptions{
		ReplaceAttr: func(a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				loggingMetrics.Inc(a.Value.Any().(slog.Level))
			}
			return a
		},
	}.NewTextHandler(loggerWriter))

	slog.SetDefault(logger)

	accessLoggerWriter := logging.ApplicationLogWriteCloser(options.LogDir, "access", os.Stderr)
	defer accessLoggerWriter.Close()
	accessLogger := slog.New(slog.NewTextHandler(accessLoggerWriter))

	// Log application version on start
	app.Functions = append(app.Functions, func() (err error) {
		logger.Info("start", "version", compromised.Version())
		return nil
	})

	// Recovery service provides unified way of logging and notifying
	// panic events.
	recoveryService := &recovery.Service{
		Version: compromised.Version(),
	}

	// Initialize server.
	srv, err := server.New(server.Options{
		Name:                  config.Name,
		Version:               compromised.Version(),
		ListenInstrumentation: options.ListenInstrumentation,
		Logger:                logger,
		RecoveryService:       recoveryService,
	})
	if err != nil {
		return fmt.Errorf("server: %w", err)
	}
	srv.WithMetrics(loggingMetrics.Metrics()...)

	passwordsService, err := filepasswords.New(options.PasswordsDB)
	if err != nil {
		return fmt.Errorf("passwords service: %w", err)
	}
	srv.WithMetrics(passwordsService.Metrics()...)
	shutdownFuncs = append(shutdownFuncs, passwordsService.Close)

	srvOptions := server.HTTPOptions{
		Name:   config.Name,
		Listen: options.Listen,
	}

	apiHandler, err := api.New(api.Options{
		Version:          compromised.Version(),
		Headers:          options.Headers,
		RealIPHeaderName: options.RealIPHeaderName,
		Logger:           logger,
		AccessLogger:     accessLogger,
		RecoveryService:  recoveryService,
		PasswordsService: passwordsService,
	})
	if err != nil {
		return fmt.Errorf("api: %w", err)
	}
	srv.WithMetrics(apiHandler.Metrics()...)
	srvOptions.SetHandler(apiHandler)

	// Configure main HTTP web server.
	if err := srv.WithHTTP(srvOptions); err != nil {
		return fmt.Errorf("configure %s server: %w", srvOptions.Name, err)
	}

	// Start web server.
	app.Functions = append(app.Functions, srv.Serve)

	// Define shutdown function.
	app.ShutdownFunc = func() error {
		// Shutdown web server.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		srv.Shutdown(ctx)
		cancel()

		// Shutdown all services in parallel.
		var wg sync.WaitGroup
		for _, shutdown := range shutdownFuncs {
			wg.Add(1)
			go func(shutdown func() error) {
				defer wg.Done()
				if err := shutdown(); err != nil {
					logger.Error("shutdown", err)
				}
			}(shutdown)
		}
		done := make(chan struct{})
		go func() {
			defer close(done)
			wg.Wait()
		}()
		select {
		case <-time.After(10 * time.Second):
		case <-done:
		}
		return nil
	}

	// Put the process in the background only if the Pid is not 1
	// (for example in docker) and the command is `daemon`.
	if syscall.Getpid() != 1 && daemon {
		app.Daemonize()
	}

	// Finally start the server.
	// This is a blocking function.
	if err := app.Start(logger); err != nil {
		return err
	}

	return nil
}
