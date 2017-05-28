package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/Sirupsen/logrus"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/snail007/pool"
	"github.com/streadway/amqp"
)

type mqConnection struct {
	conn          *amqp.Connection
	isManualClose bool
}
type message struct {
	Consumers   []consumer
	Durable     bool
	IsNeedToken bool
	Mode        string
	Name        string
	Token       string
}
type consumer struct {
	ID       string
	URL      string
	RouteKey string
	Timeout  float64
}

const (
	poolInitialCap  = 5
	poolMaxCap      = 30
	consumePoolName = "consume"
)

var (
	pools                      = make(map[string]pool.Pool)
	uri                        = "amqp://gome:gome@10.125.207.4:5672/"
	mqConnectionFailRetrySleep = 5 * time.Second
	log                        = logrus.New()
	messages                   = []message{}
)

func init() {
	var err error
	var content string
	content, err = fileGetContents("ac.json")
	fatal("get config file fail", err)
	messages, err = parseMessages(content)
	fatal("parse config file fail", err)
	for _, msg := range messages {
		log.Println(msg.Name)
	}
	p, _ := filepath.Abs("./")
	if strings.Contains(p, "/Users") {
		uri = "amqp://guest:guest@127.0.0.1:5672/"
	}
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

}
func main() {
	log.Info("service started")

	//queueDeclare("test", false, true, false, false, nil, true, consumePoolName)

	select {}

}
func fileGetContents(file string) (content string, err error) {
	defer func(err *error) {
		e := recover()
		if e != nil {
			*err = fmt.Errorf("%s", e)
		}
	}(&err)
	bytes, err := ioutil.ReadFile(file)
	content = string(bytes)
	return
}
func fatal(flag string, err interface{}) {
	if err != nil {
		log.Fatalf(flag+":%s", err)
	}
}
func value(v interface{}, defaultValue interface{}) interface{} {
	if v != nil {
		return v
	}
	return defaultValue
}
func parseMessages(str string) (messages []message, err error) {
	defer func(err *error) {
		e := recover()
		if e != nil {
			*err = fmt.Errorf("%s", e)
		}
	}(&err)
	json, _ := gabs.ParseJSON([]byte(str))
	msgs, _ := json.Children()
	for _, m := range msgs {
		consumers := []consumer{}
		cs, _ := m.S("consumers").Children()
		for _, c := range cs {
			cc := consumer{
				ID:       c.S("id").Data().(string),
				URL:      c.S("url").Data().(string),
				RouteKey: c.S("routeKey").Data().(string),
				Timeout:  c.S("timeout").Data().(float64),
			}
			consumers = append(consumers, cc)
		}
		msg := message{
			Name:        m.S("name").Data().(string),
			Durable:     m.S("durable").Data().(bool),
			IsNeedToken: m.S("isNeedToken").Data().(bool),
			Mode:        m.S("mode").Data().(string),
			Token:       m.S("token").Data().(string),
			Consumers:   consumers,
		}
		messages = append(messages, msg)
	}
	return
}
func connectToRabbitMQ(uri string) *amqp.Connection {
	for {
		conn, err := amqp.Dial(uri)
		if err == nil {
			return conn
		}
		log.Errorf("Trying to reconnect to RabbitMQ at %s\n", uri)
		time.Sleep(mqConnectionFailRetrySleep)
	}
}
func getMqConnection(poolName string) (mqConnection, pool.Pool, error) {
	for {
		if _, ok := pools[poolName]; ok {
			c, e := pools[poolName].Get()
			if e != nil {
				log.Errorf("pools[poolName].Get() fail in exists pool:%s", e)
				continue
			}
			var conn = c.(mqConnection)
			return conn, pools[poolName], e
		}
		factory := func() (interface{}, error) {
			c := connectToRabbitMQ(uri)
			errchn := make(chan *amqp.Error)
			c.NotifyClose(errchn)
			mqConn := mqConnection{
				conn:          c,
				isManualClose: false,
			}
			go func(c mqConnection) {
				var rabbitErr *amqp.Error
				for {
					rabbitErr = <-errchn
					if mqConn.isManualClose {
						break
					}
					if rabbitErr != nil {
						log.Errorf("Connecting to %s:%s", uri, rabbitErr)
						mqConn.conn = connectToRabbitMQ(uri)
						errchn = make(chan *amqp.Error)
						mqConn.conn.NotifyClose(errchn)
					}
				}
			}(mqConn)
			return mqConn, nil
		}

		close := func(v interface{}) error {
			mqConnection := v.(mqConnection)
			mqConnection.isManualClose = true
			return mqConnection.conn.Close()
		}
		poolConfig := &pool.PoolConfig{
			InitialCap:  5,
			MaxCap:      30,
			Factory:     factory,
			AutoClose:   false,
			Close:       close,
			IdleTimeout: 15 * time.Second,
		}
		p, e := pool.NewChannelPool(poolConfig)
		if e != nil {
			log.Errorf("fail pool.NewChannelPool(poolConfig):%s", e)
			continue
		}
		pools[poolName] = p
		c, e := pools[poolName].Get()
		if e != nil {
			log.Errorf("fail pools[poolName].Get():%s", e)
			continue
		}
		return c.(mqConnection), pools[poolName], e
	}
}
func getMqChannel(poolName string) *amqp.Channel {
	for {
		mqConn, poolConsme, err := getMqConnection(poolName)
		poolConsme.Put(mqConn)
		if err != nil {
			log.Errorf("getConnection("+poolName+"):%s", err)
			continue
		}
		channel, err := mqConn.conn.Channel()
		if err != nil {
			log.Errorf("mqConn.conn.Channel():%s", err)
			continue
		}
		return channel
	}
}

func queueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table, autoClose bool, poolName string) (queue amqp.Queue, channel *amqp.Channel) {
	var err error
	for {
		channel = getMqChannel(consumePoolName)
		queue, err = channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
		if err != nil {
			channel = getMqChannel(consumePoolName)
			_, err = channel.QueueDelete(name, false, false, false)
			if err != nil {
				log.Infof("hannel.QueueDelete("+name+", false, false, false):%s", err)
				continue
			}
		}
		if autoClose {
			channel.Close()
		}
		return
	}
}
func exchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table, autoClose bool, poolName string) (channel *amqp.Channel) {
	var err error
	for {
		channel = getMqChannel(consumePoolName)
		err = channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args)
		if err != nil {
			channel = getMqChannel(consumePoolName)
			err = channel.ExchangeDelete(name, false, false)
			if err != nil {
				log.Infof("hannel.ExchangeDelete("+name+", false, false, false):%s", err)
				continue
			}
		}
		if autoClose {
			channel.Close()
		}
		return
	}
}
