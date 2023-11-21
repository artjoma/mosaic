package main

import (
	"github.com/caarlos0/env/v10"
	"log/slog"
	"mosaic/api"
	"mosaic/engine"
	"mosaic/mosaicdb"
	"mosaic/storage"
	"mosaic/types"
	"mosaic/utils"
)

func main() {
	slog.Info("MOSAIC")

	appCfg := &types.AppConfig{}
	utils.PanicIfErr(env.Parse(appCfg))
	slog.Info("Config ok.")
	setup(appCfg)
}

func setup(appCfg *types.AppConfig) {
	database := mosaicdb.NewDatabase(appCfg.DbPath)
	storage := storage.NewStorage()
	engine, err := engine.NewEngine(appCfg.FilesTempFolder, database, storage, int(appCfg.FileUploadWorkersCount))
	if err != nil {
		panic(err)
	}
	httpApi := api.NewHttpApi(appCfg.ApiHttpAddress, appCfg.ApiHttpPort, engine)
	// block main thread
	httpApi.SetupHttpServer()
	// TODO gracefully shutdown using OS signals
}
