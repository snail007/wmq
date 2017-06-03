package main

import (
	"time"

	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
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
	log                            = logrus.New()
	messages                       = []message{}
)

func main() {
	//init service
	log.Info("WMQ Service Started")
	initConsumerManager()

	initMessages()

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

	uri = fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		cfg.GetString("rabbitmq.username"),
		cfg.GetString("rabbitmq.password"),
		cfg.GetString("rabbitmq.host"),
		cfg.GetInt("rabbitmq.port"),
		cfg.GetString("rabbitmq.vhost"))
	messageDataFilePath = cfg.GetString("consume.DataFile")
	messages, err = loadMessagesFromFile(messageDataFilePath)
	if err != nil {
		log.Fatalf("load message data form file fail [%s],%s", messageDataFilePath, err)
	}
	if err = initPool(); err != nil {
		log.Fatalf("init connection to rabbitmq fail : %s", err)

	}
	if err = initChannelPool(); err != nil {
		log.Fatalf("init Channel Pool fail : %s", err)
	}

}
