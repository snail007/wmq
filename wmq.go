package main

import (
	"time"

	logger "github.com/snail007/mini-logger"

	"fmt"
	"os"
)

const (
	poolInitialCap        = 5
	poolMaxCap            = 300
	poolChannelInitialCap = 10
	poolChannelMaxCap     = 1000
)

var (
	uri                            = ""
	mqHeartbeat                    = time.Second * 2
	mqConnectionAndDeadlineTimeout = time.Second * 4
	mqConnectionFailRetrySleep     = time.Second * 3
	messageDataFilePath            = ""
	messages                       = []message{}
)

func panicHandler(output string) {
	fmt.Println("called" + output)
}
func main() {
	ctx := log.With(logger.Fields{"func": "main"})
	// defer func() {
	// 	logger.Flush()
	// }()
	// l1 := logger.New(false, nil)
	// l1.AddWriter(logger.NewDefaultConsoleWriter(), logger.AllLevels)
	// l1.Info("hello world4")
	// ctx := l1.With(logger.Fields{"user": "test"})
	// ctx.Info("hello world")
	// subctx := ctx.With(logger.Fields{"queue": "haha"})
	// subctx.Info("show")

	// l1.Info("aaa")
	// subctx.Info("bbb")
	// ctx.Info("ccc")
	// // time.Sleep(time.Second * 3)

	// return
	// l1.Info("hello world5")
	//init service
	ctx.Info("WMQ Service Started")
	initConsumerManager()

	if err := initMessages(); err != nil {
		ctx.With(logger.Fields{"call": "initMessages()"}).Fatalln("%s", err)
	}

	if !cfg.GetBool("api-disable") {
		//init api service
		go serveAPI(cfg.GetString("listen.api"), cfg.GetString("api.token"))
	}

	//init publish service
	go servePublish(cfg.GetString("listen.publish"))

	select {}
}

func init() {
	var err error

	err = initConfig()
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(0)
	}

	initLog()
	ctx := log.With(logger.Fields{"func": "init"})
	uri = fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		cfg.GetString("rabbitmq.username"),
		cfg.GetString("rabbitmq.password"),
		cfg.GetString("rabbitmq.host"),
		cfg.GetInt("rabbitmq.port"),
		cfg.GetString("rabbitmq.vhost"))
	messageDataFilePath = cfg.GetString("consume.DataFile")
	messages, err = loadMessagesFromFile(messageDataFilePath)
	if err != nil {
		ctx.Fatalf("load message data form file fail [%s],%s", messageDataFilePath, err)
	}
	if err = initPool(); err != nil {
		ctx.Fatalf("init connection to rabbitmq fail : %s", err)

	}
	if err = initChannelPool(); err != nil {
		ctx.Fatalf("init Channel Pool fail : %s", err)
	}

}
