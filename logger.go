package main

import (
	"time"

	"github.com/Gurpartap/logrus-stack"
	"github.com/Sirupsen/logrus"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
)

//initLog
func initLog() {
	//log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.DebugLevel
	infoWriter, _ := rotatelogs.New(
		"log/info.%Y%m%d%H%M.log",
		rotatelogs.WithLinkName("log/info.log"),
		rotatelogs.WithMaxAge(time.Hour*24*7),
		rotatelogs.WithRotationTime(time.Hour),
	)
	errorWriter, _ := rotatelogs.New(
		"log/error.%Y%m%d%H%M.log",
		rotatelogs.WithLinkName("log/error.log"),
		rotatelogs.WithMaxAge(time.Hour*24*7),
		rotatelogs.WithRotationTime(time.Hour),
	)
	log.Hooks.Add(lfshook.NewHook(lfshook.WriterMap{
		logrus.InfoLevel:  infoWriter,
		logrus.ErrorLevel: errorWriter,
	}))
	logrus.AddHook(logrus_stack.StandardHook())
}
