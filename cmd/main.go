package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/routerarchitects/mango-go-foundation-service/internal/config"
	"github.com/routerarchitects/mango-go-foundation-service/internal/db"
	apphttp "github.com/routerarchitects/mango-go-foundation-service/internal/http"
	"github.com/routerarchitects/mango-go-foundation-service/internal/http/handlers"
	"github.com/routerarchitects/mango-go-foundation-service/internal/services"
	"github.com/routerarchitects/ow-common-mods/fiber/middleware/auth"
	"github.com/routerarchitects/ow-common-mods/servicediscovery"
	"github.com/routerarchitects/ow-common-mods/servicerpc"
	"github.com/routerarchitects/ow-common-mods/servicerpc/owsec"
	"github.com/routerarchitects/ra-common-mods/logger"
)

func main() {
	// 1. Intercept OS interrupt and termination signals
	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	// 2. Load configurations from environment variables
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to parse environment configurations: %v", err))
	}

	// 3. Initialize structured slog logger
	rootLog, loggerShutdown, err := logger.Init(cfg.Logger.Config)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize structured logger: %v", err))
	}
	defer loggerShutdown()

	if rootLog == nil {
		panic("logger initialization returned nil logger")
	}
	rootLog.InfoContext(ctx, "structured logger successfully initialized")

	// 4. Establish database connection pool
	database, err := db.Connect(ctx, cfg.Database, logger.Subsystem("db"))
	if err != nil {
		rootLog.Error("database connection failure", "error", err)
		panic(err)
	}
	defer database.Close()

	// 5. Run automated SQL migrations from schema directory
	if err := database.RunMigrations(ctx, "db/schema"); err != nil {
		rootLog.Error("database schema migration failure", "error", err)
		panic(err)
	}

	// 6. Initialize Service Discovery using common mods (conditional)
	var discovery *servicediscovery.Discovery
	if cfg.Discovery.Enabled {
		discovery, err = servicediscovery.New(
			cfg.Discovery.Config,
			cfg.Kafka.Config,
			logger.Subsystem("service-discovery"),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to create service discovery instance: %v", err))
		}
	} else {
		rootLog.Info("service discovery is disabled via configuration")
	}

	// 7. Initialize RPC client factory (conditional)
	var tokenValidator *owsec.SecurityClient
	if cfg.RPC.Enabled && cfg.Discovery.Enabled {
		rpcFactory, err := servicerpc.NewServiceRpc(
			discovery,
			servicerpc.ServiceRpcConfig{
				TLSRootCA:    cfg.Server.TLS_ROOTCA,
				InternalName: cfg.Discovery.PublicEndpoint,
			},
			logger.Subsystem("service-rpc"),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to create service RPC factory: %v", err))
		}

		// 8. Retrieve Security Client validator
		tokenValidator, err = rpcFactory.SecurityClient()
		if err != nil {
			panic(fmt.Sprintf("failed to create security auth client: %v", err))
		}
	} else {
		rootLog.Info("service RPC client factory and token validation are disabled via configuration")
	}

	// 9. Instantiate business services and handlers
	itemSvc := services.NewItemService(database)
	itemHandler := handlers.NewItemHandler(itemSvc)

	// 10. Assemble Fiber HTTP apps module
	publicAuthConfig := auth.PublicAuthConfig{}
	privateAuthConfig := auth.InternalAPIKeyConfig{
		ExpectedAPIKey: cfg.Discovery.InstanceKey,
	}

	module, err := apphttp.NewModule(apphttp.Dependencies{
		ServerLogger:      logger.Subsystem("server"),
		ServerConfig:      cfg.Server,
		SubsystemConfig:   cfg.Subsystem.Config,
		ItemHandler:       itemHandler,
		PublicAuthConfig:  publicAuthConfig,
		PrivateAuthConfig: privateAuthConfig,
		TokenValidator:    tokenValidator,
		AuthEnabled:       cfg.Auth.Enabled,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create HTTP module: %v", err))
	}

	// 11. Start service discovery heartbeat loop (conditional)
	if cfg.Discovery.Enabled && discovery != nil {
		if err := discovery.Start(ctx); err != nil {
			panic(fmt.Sprintf("failed to start service discovery publisher: %v", err))
		}
	}

	// 12. Bind HTTP ports and start Fiber apps
	serverErrCh, err := module.Start(ctx)
	if err != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if cfg.Discovery.Enabled && discovery != nil {
			if stopErr := discovery.Stop(shutdownCtx); stopErr != nil {
				rootLog.Error("failed to stop service discovery after server boot failure", "error", stopErr)
			}
		}
		panic(fmt.Sprintf("failed to start HTTPS listeners: %v", err))
	}

	// 13. Wait for OS signals or server failures
	select {
	case <-ctx.Done():
		rootLog.Info("OS shutdown signal intercepted, commencing graceful shutdown")
	case err := <-serverErrCh:
		if err != nil {
			rootLog.Error("HTTPS listener crashed unexpectedly", "error", err)
		} else {
			rootLog.Info("HTTPS listener exited normally")
		}
	}

	// 14. Graceful Shutdown sequence
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := module.Shutdown(); err != nil {
		rootLog.Error("forced HTTP shutdown occurred", "error", err)
	}

	if cfg.Discovery.Enabled && discovery != nil {
		if err := discovery.Stop(shutdownCtx); err != nil {
			rootLog.Error("failed to gracefully stop service discovery publisher", "error", err)
		}
	}

	rootLog.Info("graceful service shutdown completed, exiting")
}
