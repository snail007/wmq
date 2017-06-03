package main

import (
	"time"

	"os"

	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
)

//initLog
func initLog() {
	//log.Formatter = new(logrus.JSONFormatter)
	switch cfg.GetString("log.console-level") {
	case "info":
		log.Level = logrus.InfoLevel
	case "warn":
		log.Level = logrus.WarnLevel
	case "error":
		log.Level = logrus.ErrorLevel
	default:
		log.Level = logrus.DebugLevel
	}

	// callerLevels := logrus.AllLevels
	// stackLevels := []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.WarnLevel, logrus.ErrorLevel}
	// logrus.AddHook(logrus_stack.NewHook(callerLevels, stackLevels))
	var path = strings.TrimRight(cfg.GetString("log.dir"), "/\\")
	var levels = cfg.GetStringSlice("log.level")
	if len(levels) > 0 {
		if !pathExists(path) {
			os.Mkdir(path, 0700)
		}
		debugWriter, _ := rotatelogs.New(
			path+"/debug.%Y%m%d.log",
			rotatelogs.WithLinkName(path+"/debug.log"),
			rotatelogs.WithMaxAge(time.Hour*24*7),
			rotatelogs.WithRotationTime(time.Hour*24),
		)
		infoWriter, _ := rotatelogs.New(
			path+"/info.%Y%m%d.log",
			rotatelogs.WithLinkName(path+"/info.log"),
			rotatelogs.WithMaxAge(time.Hour*24*7),
			rotatelogs.WithRotationTime(time.Hour*24),
		)
		errorWriter, _ := rotatelogs.New(
			path+"log/error.%Y%m%d.log",
			rotatelogs.WithLinkName(path+"/error.log"),
			rotatelogs.WithMaxAge(time.Hour*24*7),
			rotatelogs.WithRotationTime(time.Hour),
		)
		if ok, _ := inArray("debug", levels); ok {
			log.Hooks.Add(lfshook.NewHook(lfshook.WriterMap{
				logrus.DebugLevel: debugWriter,
			}))
		}
		if ok, _ := inArray("info", levels); ok {
			log.Hooks.Add(lfshook.NewHook(lfshook.WriterMap{
				logrus.InfoLevel: infoWriter,
			}))
		}
		if ok, _ := inArray("error", levels); ok {
			log.Hooks.Add(lfshook.NewHook(lfshook.WriterMap{
				logrus.ErrorLevel: errorWriter,
				logrus.WarnLevel:  errorWriter,
			}))
		}

	}

}
