package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"time"

	"runtime"

	"strings"

	"io/ioutil"

	"sync"

	"github.com/Jeffail/gabs"
	"github.com/streadway/amqp"
	"gopkg.in/resty.v0"
)

type message struct {
	Consumers   []consumer
	Durable     bool
	IsNeedToken bool
	Mode        string
	Name        string
	Token       string
	Comment     string
}
type consumer struct {
	ID        string
	URL       string
	RouteKey  string
	Timeout   float64
	Code      int
	CheckCode bool
	Comment   string
}

var (
	msgLock = &sync.Mutex{}
)

func parseMessages(str string) (messages []message, err error) {
	defer func(err *error) {
		e := recover()
		if e != nil {
			*err = fmt.Errorf("%s", e)
		}
	}(&err)
	jsonData, _ := gabs.ParseJSON([]byte(str))
	msgs, _ := jsonData.Children()
	for _, m := range msgs {
		msg := message{}
		json.Unmarshal(m.Bytes(), &msg)
		messages = append(messages, msg)
	}
	return
}
func messageIsExists(name string) bool {
	for _, m := range messages {
		if m.Name == name {
			return true
		}
	}
	return false
}
func consumerIsExists(exchangeName, consumerName string) (exists bool, exchangeIndex, consumerIndex int) {
	for k1, m := range messages {
		if m.Name == exchangeName {
			for k2, c := range m.Consumers {
				if c.ID == consumerName {
					return true, k1, k2
				}
			}
		}
	}
	return false, 0, 0
}
func getMessage(name string) (msg *message, key int, err error) {
	for k, m := range messages {
		if m.Name == name {
			msg = &m
			key = k
			return
		}
	}
	err = errors.New("message not found")
	return
}
func getConsumer(msgMame, consumerID string) (consumer *consumer, msgIndex, consumerIndex int, err error) {
	_, i, e := getMessage(msgMame)
	if e != nil {
		err = e
		return
	}
	msgIndex = i
	for k, c := range messages[i].Consumers {
		if c.ID == consumerID {
			consumer = &c
			consumerIndex = k
			return
		}
	}
	err = errors.New("consumer not found")
	return
}
func config() (jsonData string, err error) {
	msgLock.Lock()
	defer msgLock.Unlock()
	j, err := json.Marshal(messages)
	jsonData = string(j)
	return
}
func statusMessage(messageName string) (jsonData string, err error) {
	m, _, err := getMessage(messageName)
	if err != nil {
		return
	}
	var ja []string

	for _, c := range m.Consumers {
		j, _ := statusConsumer(m.Name, c.ID)
		ja = append(ja, j.String())
	}
	jsonData = "[" + strings.Join(ja, ",") + "]"
	return
}
func statusConsumer(messageName, consumerID string) (json *gabs.Container, err error) {
	msgLock.Lock()
	defer msgLock.Unlock()
	c, i, _, e := getConsumer(messageName, consumerID)
	if e != nil {
		return nil, e
	}
	m := messages[i]
	var q amqp.Queue
	//var ch *amqp.Channel
	//ch, _ = getMqChannel()
	//q, e = ch.QueueDeclare(getConsumerKey(m, *c), m.Durable, !m.Durable, false, false, nil)
	//defer channelPools.Put(ch)
	q, _, e = queueDeclare(getConsumerKey(m, *c), m.Durable)
	if e != nil {
		return nil, e
	}
	count := q.Messages
	var jsonObj = gabs.New()
	jsonObj.Set(count, "Count")
	jsonObj.Set(consumerID, "ID")
	jsonObj.Set(messageName, "MsgName")
	jsonObj.Set("0", "LastTime")
	lasttime, e := statusConsumerWorker(*c, m)
	if e == nil {
		jsonObj.Set(lasttime, "LastTime")
	}
	//jsonObj.StringIndent("", "  ")
	//json = jsonObj.String()
	json = jsonObj
	return
}
func addMessage(m message) (err error) {
	msgLock.Lock()
	defer msgLock.Unlock()
	messages = append(messages, m)
	log.Infof("message [ %s ] was added", m.Name)
	initMessages()
	return nil
}
func updateMessage(m message) (err error) {
	msgLock.Lock()
	defer msgLock.Unlock()
	_, i, e := getMessage(m.Name)
	if e != nil {
		err = e
		return
	}
	m.Consumers = messages[i].Consumers
	messages[i] = m
	for _, c := range m.Consumers {
		_, e := stopConsumerWorker(c, m)
		if e != nil {
			err = e
			log.Infof("message [ %s ] was updated fail , %s", m.Name, e)
			return
		}
	}
	log.Infof("message [ %s ] was updated", m.Name)
	initMessages()
	return
}
func deleteMessage(m message) (err error) {
	msgLock.Lock()
	defer msgLock.Unlock()
	msg, i, e := getMessage(m.Name)
	if e != nil {
		err = e
		return
	}
	//stop all consumer worker
	for _, c := range msg.Consumers {
		_, e := stopConsumerWorker(c, m)
		if e != nil {
			err = e
			log.Infof("message [ %s ] was delete fail [stopConsumerWorker] , %s", m.Name, e)
			return
		}
		er := deleteQueue(getConsumerKey(*msg, c))
		if er != nil {
			err = er
			log.Infof("message [ %s ] was delete fail [deleteQueue] , %s", m.Name, e)
			return
		}
	}
	//update messsages data
	messages = append(messages[:i], messages[i+1:]...)
	log.Infof("message [ %s ] was deleted", m.Name)
	initMessages()
	return
}
func addConsumer(msg message, c0 consumer) (err error) {
	msgLock.Lock()
	defer msgLock.Unlock()
	//add it to messages
	var i int
	_, i, err = getMessage(msg.Name)
	if err != nil {
		return
	}
	//update messages data
	messages[i].Consumers = append(messages[i].Consumers, c0)
	//update worker
	initMessages()
	log.Infof("consumer [ %s ] was added", getConsumerKey(msg, c0))
	return
}

func updateConsumer(msg message, c0 consumer) (err error) {
	msgLock.Lock()
	defer msgLock.Unlock()
	_, i, k, e := getConsumer(msg.Name, c0.ID)
	if e != nil {
		return e
	}
	//update messages data
	messages[i].Consumers[k] = c0

	//update consumer worker
	_, err = updateConsumerWorker(c0, msg)
	return
}

func deleteConsumer(msg message, c0 consumer) (err error) {
	msgLock.Lock()
	defer msgLock.Unlock()
	_, i0, i, _ := getConsumer(msg.Name, c0.ID)
	//delete messages consumer
	messages[i0].Consumers = append(messages[i0].Consumers[:i], messages[i0].Consumers[i+1:]...)
	//stop consumer
	_, err = stopConsumerWorker(c0, msg)
	//update messages worker
	initMessages()
	log.Infof("consumer [ %s ] was deleted", getConsumerKey(msg, c0))
	return
}

func publish(body, exchangeName, routeKey, token string) (err error) {
	var msg *message
	msg, _, err = getMessage(exchangeName)
	if err != nil {
		return
	}
	if msg.IsNeedToken && token != msg.Token {
		err = errors.New("token error")
		return
	}
	var channel *amqp.Channel
	channel, err = getMqChannel()
	if err == nil {
		err = channel.Publish(exchangeName, routeKey, false, false, amqp.Publishing{
			Body: []byte(body),
		})
		channelPools.Put(channel)
		if err == nil {
			log.Debugf("publish message success , exchange:%s", exchangeName)
			return
		}
		log.Warnf("publish message fail , exchange:%s,%s", exchangeName, err)
	} else {
		log.Warnf("publish message ->getMqChannel , exchange:%s,%s", exchangeName, err)
	}
	return
}

//WrapedConsumer xx
type wrapedConsumer struct {
	consumer  consumer
	message   message
	action    string
	lasttime  int64
	readChan  <-chan string
	writeChan chan<- string
}

type manageConsumer struct {
	wrapedConsumer
	consumerReadChan  chan string
	consumerWriteChan chan string
	key               string
}

var consumerManageReadChan, consumerManageWriteChan = make(chan wrapedConsumer, 1), make(chan string, 1)

func updateConsumerWorker(c consumer, m message) (answer string, err error) {
	return notifyConsumerManager("update", c, m)
}
func stopConsumerWorker(c consumer, m message) (answer string, err error) {
	return notifyConsumerManager("delete", c, m)
}
func statusConsumerWorker(c consumer, m message) (answer string, err error) {
	return notifyConsumerManager("status", c, m)
}
func notifyConsumerManager(action string, c consumer, m message) (answer string, err error) {
	if action != "update" && action != "delete" && action != "status" {
		err = fmt.Errorf("action must be [update|delete]")
		return
	}
	log.Debugf("sending %s to %s , waiting answer ...", action, m.Name+"_"+c.ID)
	consumerManageReadChan <- wrapedConsumer{
		action:   action,
		consumer: c,
		message:  m,
	}
	log.Debugf("send %s to %s okay , response to manage ...", action, m.Name+"_"+c.ID)
	answer = <-consumerManageWriteChan
	log.Debugf("manage answer : %s", answer)
	return
}
func getConsumerKey(m message, c consumer) string {
	return m.Name + c.ID
}
func initMessages() {
	for _, m := range messages {
		_, err := exchangeDeclare(m.Name, m.Mode, m.Durable)
		if err == nil {
			for _, c := range m.Consumers {
				_, _, err := queueDeclare(getConsumerKey(m, c), m.Durable)
				if err == nil {
					err := queueBindToExchange(getConsumerKey(m, c), m.Name, c.RouteKey)
					if err == nil {
						answer, err := updateConsumerWorker(c, m)
						if err == nil {
							log.Debugf("updateConsumer answer %s ", answer)
						} else {
							log.Warnf("updateConsumer %s ", err)
						}
					} else {
						log.Warnf("queueBindToExchange %s ", err)
					}
				} else {
					log.Warnf("queueDeclare %s ", err)
				}
			}
		} else {
			log.Warnf("exchangeDeclare %s ", err)
		}
	}
}
func restart() (err error) {
	msgLock.Lock()
	defer msgLock.Unlock()
	err = stopAllConsumer()
	if err != nil {
		return
	}
	initMessages()
	return
}
func reload() {
	msgLock.Lock()
	defer msgLock.Unlock()
	initMessages()
}
func stopAllConsumer() (err error) {
	for _, m := range messages {
		for _, c := range m.Consumers {
			_, err = stopConsumerWorker(c, m)
			if err != nil {
				return
			}
		}
	}
	return
}
func initConsumerManager() {
	wrapedConsumers := make(map[string]*manageConsumer)
	go func() {
		log.Debugf("Consumer Manager started")
		//Consumer Manager worker loop,waiting for control command and exec switch
		for {
			//waiting for control data
			wrapedConsumer := <-consumerManageReadChan

			log.Debugf("revecived consumer[%s]", wrapedConsumer.action)
			//preprocess for control data
			c := &manageConsumer{
				wrapedConsumer:    wrapedConsumer,
				consumerReadChan:  make(chan string, 1),
				consumerWriteChan: make(chan string, 1),
				key:               getConsumerKey(wrapedConsumer.message, wrapedConsumer.consumer),
			}
			//decide what to do
			switch wrapedConsumer.action {
			case "update": //update or insert consumer
				t := "update"
				if _, ok := wrapedConsumers[c.key]; !ok {
					t = "insert"
				} else {
					c.lasttime = wrapedConsumers[c.key].lasttime
				}
				log.Debugf("%s consumer[%s]", t, c.key)
				wrapedConsumers[c.key] = c
				if t == "insert" {
					//start consumer go
					go func() {
						defer func() {
							delete(wrapedConsumers, c.key)
							log.Warnf("consumer[%s] goroutine exited", c.key)
						}()
						//0.consumer worker start
						for {
						RETRY:
							//1.check if exists in wrapedConsumers data
							if _, ok := wrapedConsumers[c.key]; !ok {
								log.Warnf("consumer[ %s ] not found , now exit", c.key)
								runtime.Goexit()
							}
							//update consumer active time
							wrapedConsumers[c.key].lasttime = time.Now().Unix()
							//2.try get connection
							conn, err := pools.Get()
							if err != nil {
								pools.Put(conn)
								log.Warnf("get conn fail for consumer %s , sleep 3 seconds ... %s", c.key, err)
								time.Sleep(time.Second * 3)
								continue
							}
							//3.try get channel on connecton
							channel, err := conn.(*amqp.Connection).Channel()
							if err != nil {
								pools.Put(conn)
								log.Warnf("get channel fail for consumer %s , sleep 3 seconds ... %s", c.key, err)
								time.Sleep(time.Second * 3)
								continue
							}
							//4.try  declare exchange  on channel
							_, err = exchangeDeclare(wrapedConsumers[c.key].message.Name,
								wrapedConsumers[c.key].message.Mode,
								wrapedConsumers[c.key].message.Durable)
							if err != nil {
								pools.Put(conn)
								log.Warnf("exchangeDeclare() fail for consumer %s , sleep 3 seconds ... %s", c.key, err)
								time.Sleep(time.Second * 3)
								continue
							}
							//5.try  declare queue  on channel
							_, _, err = queueDeclare(getConsumerKey(wrapedConsumers[c.key].message, wrapedConsumers[c.key].consumer),
								wrapedConsumers[c.key].message.Durable)
							if err != nil {
								pools.Put(conn)
								log.Warnf("queueDeclare() fail for consumer %s , sleep 3 seconds ... %s", c.key, err)
								time.Sleep(time.Second * 3)
								continue
							}
							//6.try  bind queue to exchange
							err = queueBindToExchange(getConsumerKey(wrapedConsumers[c.key].message, wrapedConsumers[c.key].consumer),
								wrapedConsumers[c.key].message.Name,
								wrapedConsumers[c.key].consumer.RouteKey)
							if err != nil {
								pools.Put(conn)
								log.Warnf("queueBindToExchange() fail for consumer %s , sleep 3 seconds ... %s", c.key, err)
								time.Sleep(time.Second * 3)
								continue
							}
							//7.try  set qos on channel
							err = channel.Qos(1, 0, false)
							if err != nil {
								pools.Put(conn)
								log.Warnf("set channel Qos fail for consumer , sleep 3 seconds ... %s", err)
								time.Sleep(time.Second * 3)
								continue
							}
							//8.try consume queue
							deliveryChn, err := channel.Consume(getQueueName(c.key), "", false, false, false, false, nil)
							if err != nil {
								pools.Put(conn)
								log.Warnf("get deliveryChn fail for consumer , sleep 3 seconds ...%s", err)
								time.Sleep(time.Second * 3)
								continue
							}
							//worker loop,use chan waiting for control command or delivery
							for {
								if _, ok := wrapedConsumers[c.key]; !ok {
									log.Warnf("consumer[ %s ] not found , now exit", c.key)
									runtime.Goexit()
								}
								//update consumer active time
								wrapedConsumers[c.key].lasttime = time.Now().Unix()
								log.Infof("Consumer [ %s ] waiting for message ...", c.key)
								select {
								case cmd := <-wrapedConsumers[c.key].consumerReadChan:
									if cmd == "exit" {
										pools.Put(conn)
										wrapedConsumers[c.key].consumerWriteChan <- "exit_ok"
										runtime.Goexit()
									}
								case delivery, ok := <-deliveryChn:
									if !ok {
										pools.Put(conn)
										log.Warnf("get delivery fail for consumer")
										goto RETRY
									}
									log.Debugf("delivery revecived [ %s ] : %s", c.key, string(delivery.Body)[0:20]+"...")
									if process(string(delivery.Body), c.consumer) == nil {
										//process success
										err = delivery.Ack(false)
									} else {
										//process fail
										err = delivery.Nack(false, true)
									}
									if err != nil {
										log.Warnf("consumer %s ack/nack fail , %s", c.key, err)
									}
									time.Sleep(time.Second * 5)
								}
							}
						}
					}()
				}
				consumerManageWriteChan <- "completed"
			case "delete": //stop consumer
				if _, ok := wrapedConsumers[c.key]; ok {
					log.Debugf("sending stop singal to consumer[%s] goroutine ...", c.key)
					wrapedConsumers[c.key].consumerReadChan <- "exit"
					log.Debugf("send  stop singal to consumer[%s] goroutine success", c.key)
				}
				consumerManageWriteChan <- "completed"
			case "status": //get consumer last active time
				v, ok := wrapedConsumers[c.key]
				var t int64
				if ok {
					t = v.lasttime
				}
				consumerManageWriteChan <- fmt.Sprintf("%d", t)
			default:
				consumerManageWriteChan <- "unkown command"
			}
		}
	}()
}
func writeMessagesToFile(messages0 []message, configFilePath0 string) (err error) {
	msgLock.Lock()
	defer msgLock.Unlock()
	b, err := json.Marshal(messages0)
	c, err := gabs.ParseJSON(b)
	content := c.StringIndent("", "	")
	err = ioutil.WriteFile(configFilePath0, []byte(content), 0600)
	return
}
func loadMessagesFromFile(configFilePath0 string) (messages0 []message, err error) {
	msgLock.Lock()
	defer msgLock.Unlock()
	if _, err := os.Stat(configFilePath0); os.IsNotExist(err) {
		err = ioutil.WriteFile(configFilePath0, []byte("[]"), 0600)
	} else {
		var content string
		content, err = fileGetContents(configFilePath0)
		if err == nil {
			messages0, err = parseMessages(content)
			if err == nil {
				messages = messages0
			}
		}
	}
	return
}

func process(content string, c consumer) (err error) {
	//content = "{\"body\":\"sss\",\"header\":{\"ID\":\"test\"},\"ip\":\"127.0.0.1\",\"method\":\"get\"}"
	jsonParsed, err := gabs.ParseJSON([]byte(content))
	if err != nil {
		log.Warnf("message from rabbitmq not suppported and drop it, msg : %s", content)
		return nil
	}

	if !jsonParsed.Exists("body") || !jsonParsed.Exists("header") ||
		!jsonParsed.Exists("ip") || !jsonParsed.Exists("method") || !jsonParsed.Exists("args") {
		log.Warnf("message from rabbitmq not suppported and drop it, msg : %s", content)
		return nil
	}

	body := jsonParsed.S("body").Data()
	header, _ := jsonParsed.S("header").ChildrenMap()
	ip := jsonParsed.S("ip").Data().(string)
	args := jsonParsed.S("args").Data().(string)
	method := jsonParsed.S("method").Data().(string)
	url := c.URL
	if args != "" {
		if strings.Contains(c.URL, "?") {
			url = c.URL + "&" + args
		} else {
			url = c.URL + "?" + args
		}
	}
	//log.Warnf("body:%s", body)
	var headerMap = make(map[string]string)
	for k, child := range header {
		headerMap[k] = child.Data().(string)
		log.Warnf("%s:%s", k, child.Data())
	}
	//log.Warnf("method:%s", method)
	client := resty.New().
		SetTimeout(time.Millisecond*time.Duration(c.Timeout)).R().
		SetHeaders(headerMap).
		SetHeader("X-Forward-For", ip).
		SetHeader("User-Agent", "wmq v0.1 - https://github.com/snail007/wmq")
	var resp *resty.Response
	if method == "post" {
		resp, err = client.SetBody(body).Post(url)
	} else if method == "get" {
		resp, err = client.Get(url)
	} else {
		err = fmt.Errorf("method [ %s ] not supported", method)
	}
	if err != nil {
		log.Warnf("consume fail [ %s ] %s", c.URL, err)
		return
	}
	if c.CheckCode {
		code := resp.StatusCode()
		if code != c.Code {
			err = fmt.Errorf("consume fail [ %s ] , response http code is %d , 200 expected ", c.URL, code)
			log.Warnf("consume fail [ %s ] , response http code is %d , 200 expected ", c.URL, code)
		} else {
			log.Debugf("consume success , %s", c.URL)
		}
	} else {
		log.Debugf("consume success , %s", c.URL)
	}
	return
}
