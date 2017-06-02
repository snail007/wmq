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
	messageDataFilePath            = "message.json"
	log                            = logrus.New()
	messages                       = []message{}
)

func main() {

	go func() {
		//process("aaa", consumer{URL: "http://gitcode.com/wmq.php", Timeout: 3000})
		// for {
		// time.Sleep(time.Second * 5)
		// messages[0].Consumers[0].URL = "init url"
		// updateConsumer(messages[0].Consumers[0], messages[0])
		// time.Sleep(time.Second * 5)
		// messages[0].Consumers[0].URL = "updated url"
		// updateConsumer(messages[0].Consumers[0], messages[0])
		go func() {
			// for {
			// 	status, err := statusConsumerWorker(messages[0].Consumers[0], messages[0])
			// 	if err == nil {
			// 		i, _ := strconv.ParseInt(status, 10, 64)
			// 		log.Infof("%d,%s", i, time.Unix(i, 0).Format("2006-01-02 15:04:05"))
			// 	}
			// 	time.Sleep(time.Second * 4)
			// }
		}()
		//   "Durable": false,
		//     "IsNeedToken": true,
		//     "Mode": "topic",
		//     "Name": "test",
		//     "Token": "JQJsUOqYzYZZgn8gUvs7sIinrJ0tDD8J"
		// time.Sleep(time.Second * 5)
		// addMessage(message{
		// 	Name:        "vaddtest",
		// 	Durable:     false,
		// 	IsNeedToken: true,
		// 	Mode:        "fanout",
		// 	Token:       "fadafasdfs",
		// })
		// status, _ := statusConsumer(messages[0].Consumers[0], messages[0])
		// i, _ := strconv.ParseInt(status, 10, 64)
		// log.Infof("%d,%s", i, time.Unix(i, 0).Format("2006-01-02 15:04:05"))

		// time.Sleep(time.Second * 30)
		// if writeConfigToFile(messages, configFilePath) == nil {
		// 	log.Info("write success")
		// }
		//reload()
		// deleteConsumer(messages[0], messages[0].Consumers[0])
		// time.Sleep(time.Second * 5)
		// addConsumer(messages[1], consumer{
		// 	ID:       "1212121",
		// 	RouteKey: "",
		// })
		// publish("hello world", "addtest", "", "fadafasdfs")

		// time.Sleep(time.Second * 5)
		// log.Info(config())
		// err := publish("hello world 00000000000", "vaddtest", "", "fadafasdfs")
		// if err != nil {
		// 	log.Errorf("%s", err)
		// }
		// time.Sleep(time.Second * 5)
		// updateMessage(message{
		// 	Name:        "vaddtest",
		// 	Durable:     false,
		// 	IsNeedToken: true,
		// 	Mode:        "topic",
		// 	Token:       "fadafasdfs",
		// })
		// time.Sleep(time.Second * 5)
		// deleteMessage(message{
		// 	Name:        "vaddtest",
		// 	Durable:     false,
		// 	IsNeedToken: true,
		// 	Mode:        "topic",
		// 	Token:       "fadafasdfs",
		// })
		// time.Sleep(time.Second * 30)
		// messages[0].Consumers[1].URL = "333 URL"
		//deleteConsumer(messages[0], messages[0].Consumers[1])
		// updateConsumer(messages[0], messages[0].Consumers[1])
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
		// for {
		// 	time.Sleep(time.Second * 3)
		// 	err := publish("hello world", "test", "test", "JQJsUOqYzYZZgn8gUvs7sIinrJ0tDD8J")
		// 	if err != nil {
		// 		log.Errorf("publish %s ", err)
		// 	}
		// 	a, _ := statusMessage("test")
		// 	log.Infof("%s  %s", a, err)
		// }
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

	//init service
	log.Info("WMQ Service Started")
	initConsumerManager()
	initMessages()
	//init api service
	go serveAPI(3302, "abc")
	//init publish service
	go servePublish(3303)
}
