package main

import (
	"path/filepath"
	"strings"

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
	uri                            = "amqp://gome:gome@10.125.207.4:5672/"
	mqHeartbeat                    = time.Second * 2
	mqConnectionAndDeadlineTimeout = time.Second * 4
	mqConnectionFailRetrySleep     = time.Second * 3
	messageDataFilePath            = "message.json"
	log                            = logrus.New()
	messages                       = []message{}
)

func main() {
	//init service
	log.Info("WMQ Service Started")
	initConsumerManager()
	initMessages()
	//init api service
	go serveAPI(3302, "abc")
	//init publish service
	go servePublish(3303)
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

	p, _ := filepath.Abs("./")
	if strings.Contains(p, "/Users") {
		uri = "amqp://guest:guest@127.0.0.1:5672/"
	}

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
