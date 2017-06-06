package main

import (
	"github.com/snail007/mini-logger"
	"github.com/snail007/mini-logger/writers/console"
	"github.com/snail007/mini-logger/writers/files"
)

var log logger.MiniLogger

//initLog
func initLog() {
	var level uint8
	switch cfg.GetString("log.console-level") {
	case "debug":
		level = logger.AllLevels
	case "info":
		level = logger.InfoLevel | logger.WarnLevel | logger.ErrorLevel | logger.FatalLevel
	case "warn":
		level = logger.WarnLevel | logger.ErrorLevel | logger.FatalLevel
	case "error":
		level = logger.ErrorLevel | logger.FatalLevel
	case "fatal":
		level = logger.FatalLevel
	default:
		level = 0
	}
	log = logger.New(false, nil)
	log.AddWriter(console.NewDefault(), level)
	cfgF := files.GetDefaultFileConfig()
	cfgF.LogPath = cfg.GetString("log.dir")
	log.AddWriter(files.New(cfgF), logger.AllLevels)
}
