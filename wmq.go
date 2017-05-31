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
	poolChannelInitialCap = 10
	poolChannelMaxCap     = 1000
)

var (
	uri                            = "amqp://gome:gome@10.125.207.4:5672/"
	mqHeartbeat                    = time.Second * 2
	mqConnectionAndDeadlineTimeout = time.Second * 4
	mqConnectionFailRetrySleep     = time.Second * 3

	log      = logrus.New()
	messages = []message{}
)

func main() {

	//log.Info("service started")
	// for i := 0; i < 10; i++ {
	exchangeDeclare("test", "fanout", true)
	exchangeDeclare("test", "fanout", false)
	// }
	queueDeclare("test", true)
	queueDeclare("test", false)
	queueBindToExchange("test", "test", "")
	// err := publish("hello haha", "test", "test", "JQJsUOqYzYZZgn8gUvs7sIinrJ0tDD8J", 2)
	// if err != nil {
	// 	log.Error(err)
	// } else {
	// 	log.Info("send SUCCESS")
	// }

	go func() {
		// for {
		// 	time.Sleep(time.Second * 2)
		// 	log.Infof("pool len : %d , channel pool len : %d", pools.Len(), channelPools.Len())
		// }
	}()
	go func() {
		//time.Sleep(time.Second * 10)
		// saveConsumer("test", consumer{
		// 	ID:       "333",
		// 	URL:      "URL",
		// 	Timeout:  5200,
		// 	RouteKey: "#",
		// })
		//time.Sleep(time.Second * 3)
		//deleteConsumer(messages[0], messages[0].Consumers[2])
		// saveConsumer("test", consumer{
		// 	ID:       "333",
		// 	URL:      "URL",
		// 	Timeout:  5200,
		// 	RouteKey: "test",
		// })
		//log.Debug("waiting...")
		//time.Sleep(time.Second * 3)
		//publish("hello world", "test", "test", "JQJsUOqYzYZZgn8gUvs7sIinrJ0tDD8J")
	}()
	select {}
}

func init() {
	var err error
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

	if err = initPool(); err != nil {
		log.Fatalf("init connection to rabbitmq fail : %s", err)

	}
	if err = initChannelPool(); err != nil {
		log.Fatalf("init Channel Pool fail : %s", err)
	}

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
