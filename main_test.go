package main

import (
	"github.com/caarlos0/env/v10"
	"mosaic/types"
	"mosaic/utils"
	"testing"
)

func TestStartApp(t *testing.T) {
	appCfg := &types.AppConfig{}
	// DB in mem!
	utils.PanicIfErr(env.Parse(appCfg))

	setup(appCfg)
}
