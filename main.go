package main

import (
	"context"
	"github.com/caarlos0/env/v10"
	"log/slog"
	"mosaic/api"
	"mosaic/engine"
	"mosaic/mosaicdb"
	"mosaic/storage"
	"mosaic/types"
	"mosaic/utils"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	slog.Info("MOSAIC")

	appCfg := &types.AppConfig{}
	utils.PanicIfErr(env.Parse(appCfg))
	slog.Info("Config ok.")
	sigs := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	cancelCtx := setup(appCfg)

	<-sigs
	cancelCtx()
}

func setup(appCfg *types.AppConfig) context.CancelFunc {
	database := mosaicdb.NewDatabase(appCfg.DbPath)
	storage := storage.NewStorage()
	appCtx, cancel := context.WithCancel(context.Background())
	engine, err := engine.NewEngine(appCtx, appCfg.FilesTempFolder, database, storage, int(appCfg.FileUploadWorkersCount))
	if err != nil {
		panic(err)
	}

	// setup http API
	httpApi := api.NewHttpApi(appCfg.ApiHttpAddress, appCfg.ApiHttpPort, engine)
	go func() {
		// block main thread
		httpApi.SetupHttpServer()
	}()

	return cancel

}
