package main

import (
	"github.com/snail007/mini-logger"
)

var log logger.MiniLogger

//initLog
func initLog() {
	var level byte
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
	log.AddWriter(&logger.ConsoleWriter{
		Format: "[{level}] [{date} {time}.{mili}] {text} {fields}",
	}, level)
}
