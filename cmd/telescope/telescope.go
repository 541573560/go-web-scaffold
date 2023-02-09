package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"telescope/cache"
	"telescope/controller"
	"telescope/database"
	"telescope/version"
	"time"

	"github.com/BurntSushi/toml"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

var (
	configPath = flag.String("config", "config/telescope.toml", "path to config file")
)

func init() {
	version.SetSubName("API")
}

func main() {
	flag.Parse()

	var (
		err    error
		config Config
		logger *zap.Logger
	)

	defer func() {
		if err != nil {
			if logger != nil {
				logger.Error("telescope exits with error", zap.Error(err))
			} else {
				fmt.Println(err)
			}

			os.Exit(1)
		}
	}()

	_, err = toml.DecodeFile(*configPath, &config)
	if err != nil {
		err = fmt.Errorf("loading config file: toml.DecodeFile: %w", err)
		return
	}

	if config.Log.ProductionMode {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		err = fmt.Errorf("initializing zap logger: %w", err)
		return
	}
	defer logger.Sync() // nolint: errcheck
	zap.ReplaceGlobals(logger)

	logger.Info("starting...", zap.String("version", version.FullNameWithBuildDate))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("connecting to database...",
		zap.String("host", config.Database.Host),
		zap.String("db", config.Database.DatabaseName))
	db, err := database.NewDB(config.Database)
	if err != nil {
		err = fmt.Errorf("database.NewDatabase: %w", err)
		return
	}
	defer db.Close() // nolint: errcheck
	logger.Info("database connected")

	logger.Info("connecting to Redis...")
	redis, err := cache.NewRedisClient(ctx, config.Redis)
	if err != nil {
		err = fmt.Errorf("cache.NewRedisClient: %w", err)
		return
	}
	logger.Info("Redis connected")

	server := controller.NewServer(controller.ServerOpt{
		Port:          config.API.Port,
		Logger:        logger,
		Database:      db,
		Redis:         redis,
		AuditResponse: config.API.AuditResponse,
	})

	logger.Info("public API service is starting", zap.Int("port", config.API.Port))

	err = server.ListenAndServe()
	if err != nil {
		return
	}
}
