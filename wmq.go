package main

import (
	"path/filepath"
	"strings"

	"time"

	"github.com/Sirupsen/logrus"
)

const (
	poolInitialCap        = 5
	poolMaxCap            = 300
	poolChannelInitialCap = 100
	poolChannelMaxCap     = 1000
	consumePoolName       = "consume"
	publishPoolName       = "publish"
)

var (
	uri                            = "amqp://gome:gome@10.125.207.4:5672/"
	mqHeartbeat                    = time.Second * 2
	mqConnectionAndDeadlineTimeout = time.Second * 6
	mqConnectionFailRetrySleep     = time.Second * 5

	log      = logrus.New()
	messages = []message{}
)

func main() {

	//log.Info("service started")
	//exchangeDeclare("test", "fanout", true, true, false, false, nil, true, "publish", 1)
	//exchangeDeclare("test", "fanout", false, true, "publish", 1)
	//queueDeclare("test", false, true, consumePoolName, 1)
	//queueBindToExchange("test", "test", "")
	err := publish("hello haha", "test", "test", "JQJsUOqYzYZZgn8gUvs7sIinrJ0tDD8J", 2)
	if err != nil {
		log.Error(err)
	} else {
		log.Info("send SUCCESS")
	}
	go func() {
		time.Sleep(time.Second * 10)
		addConsumer("test", consumer{
			ID:       "333",
			URL:      "URL",
			Timeout:  5200,
			RouteKey: "#",
		})
		time.Sleep(time.Second * 3)
		deleteConsumer(messages[0], messages[0].Consumers[2])
		publish("hello world", "test", "test", "JQJsUOqYzYZZgn8gUvs7sIinrJ0tDD8J", 2)
	}()
	select {}
}

func init() {

	//begin init logger
	initLog()
	//end init logger

	//begin init var
	p, _ := filepath.Abs("./")
	if strings.Contains(p, "/Users") {
		uri = "amqp://guest:guest@127.0.0.1:5672/"
	}
	content, err := fileGetContents("a.json")
	fatal("get config file fail", err)
	messages, err = parseMessages(content)
	fatal("parse config file fail", err)
	initMessages()
	//end init var

	// b, _ := json.Marshal(messages)
	// c, _ := gabs.ParseJSON(b)
	// log.Info(c.StringIndent("", "	"))

}
func fatal(flag string, err interface{}) {
	if err != nil {
		log.Fatalf(flag+":%s", err)
	}
}
