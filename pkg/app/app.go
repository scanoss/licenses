// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2018-2022 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package app

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golobby/config/v3"
	"github.com/golobby/config/v3/pkg/feeder"
	"github.com/scanoss/go-grpc-helper/pkg/files"
	gs "github.com/scanoss/go-grpc-helper/pkg/grpc/server"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"scanoss.com/licenses/pkg/cache"
	myconfig "scanoss.com/licenses/pkg/config"
	scanossdb "scanoss.com/licenses/pkg/db"
	models "scanoss.com/licenses/pkg/model"
	"scanoss.com/licenses/pkg/protocol/grpc"
	"scanoss.com/licenses/pkg/protocol/rest"
	"scanoss.com/licenses/pkg/server"
)

//go:generate bash ../../get_version.sh
//go:embed version.txt
var version string

// getConfig checks command line args for option to feed into the config parser.
func getConfig() (*myconfig.ServerConfig, error) {
	var jsonConfig, envConfig, loggingConfig string
	flag.StringVar(&jsonConfig, "json-config", "", "Application JSON config")
	flag.StringVar(&envConfig, "env-config", "", "Application dot-ENV config")
	flag.StringVar(&loggingConfig, "logging-config", "", "Logging config file")
	debug := flag.Bool("debug", false, "Enable debug")
	ver := flag.Bool("version", false, "Display current version")
	flag.Parse()
	if *ver {
		fmt.Printf("Version: %v", version)
		os.Exit(1)
	}
	var feeders []config.Feeder
	if len(jsonConfig) > 0 {
		feeders = append(feeders, feeder.Json{Path: jsonConfig})
	}
	if len(envConfig) > 0 {
		feeders = append(feeders, feeder.DotEnv{Path: envConfig})
	}
	if *debug {
		err := os.Setenv("APP_DEBUG", "1")
		if err != nil {
			fmt.Printf("Warning: Failed to set env APP_DEBUG to 1: %v", err)
			return nil, err
		}
	}
	myConfig, err := myconfig.NewServerConfig(feeders)
	if len(loggingConfig) > 0 {
		myConfig.Logging.ConfigFile = loggingConfig // Override any logging config file with this one.
	}
	return myConfig, err
}

// Boostrap runs the gRPC License Server.
func Boostrap() error {
	// Load command line options and config
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	// Check mode to determine which logger to load
	err = zlog.SetupAppLogger(cfg.App.Mode, cfg.Logging.ConfigFile, cfg.App.Debug)
	if err != nil {
		return err
	}
	defer zlog.SyncZap()
	// Check if TLS/SSL should be enabled
	startTLS, err := files.CheckTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile)
	if err != nil {
		return err
	}
	// Check if IP filtering should be enabled
	allowedIPs, deniedIPs, err := files.LoadFiltering(cfg.Filtering.AllowListFile, cfg.Filtering.DenyListFile)
	if err != nil {
		return err
	}
	zlog.S.Infof("Starting SCANOSS License Service: %v", strings.TrimSpace(version))
	// Open the serving knowledge base. The file-based drivers take a path; postgres
	// takes an explicit DSN, or one assembled from the host settings.
	db, err := scanossdb.Open(scanossdb.Config{
		Driver: cfg.Database.Driver,
		Path:   cfg.Database.Path,
		DSN:    postgresDSN(cfg),
	})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			zlog.S.Warnf("Problem closing the database: %v", closeErr)
		}
	}()
	// Setup dynamic logging (if necessary)
	zlog.SetupAppDynamicLogging(cfg.Logging.DynamicPort, cfg.Logging.DynamicLogging)

	ctx := context.Background()
	// Initialize SPDX license cache
	refreshSPDXCacheTime := time.Duration(cfg.Cache.SPDXRefreshHours) * time.Hour
	spdxCache := cache.NewSPDXLicenseCache(models.NewLicenseRegistryModel(db), zlog.S, refreshSPDXCacheTime)
	if err = spdxCache.Start(ctx); err != nil {
		return fmt.Errorf("failed to initialize SPDX license cache: %v", err)
	}
	defer spdxCache.Stop()

	v2API := server.NewLicenseServer(cfg, db, spdxCache)
	// Start the REST grpc-gateway if requested
	var srv *http.Server
	if len(cfg.App.RESTPort) > 0 {
		if srv, err = rest.RunServer(cfg, ctx, cfg.App.GRPCPort, cfg.App.RESTPort, allowedIPs, deniedIPs, startTLS); err != nil {
			fmt.Printf("Failed to start REST server: %v", err)
			return err
		}
	}
	// Start the gRPC service
	server, err := grpc.RunServer(cfg, v2API, cfg.App.GRPCPort, allowedIPs, deniedIPs, startTLS, version)
	if err != nil {
		return err
	}
	// graceful shutdown
	return gs.WaitServerComplete(srv, server)
}

// postgresDSN returns the connection string for the postgres driver, assembling one
// from the individual host settings when no explicit DSN is configured. It keeps the
// format go-grpc-helper's OpenDBConnection used to build, so existing postgres
// deployments read the same configuration as before. Empty for the file-based
// drivers, which use Path instead.
func postgresDSN(cfg *myconfig.ServerConfig) string {
	if cfg.Database.Driver != scanossdb.DriverPostgres {
		return ""
	}
	if len(cfg.Database.Dsn) > 0 {
		return cfg.Database.Dsn
	}
	return fmt.Sprintf("%s://%s:%s@%s/%s?sslmode=%s", scanossdb.DriverPostgres, cfg.Database.User,
		cfg.Database.Passwd, cfg.Database.Host, cfg.Database.Schema, cfg.Database.SslMode)
}
