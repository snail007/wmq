package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"time"

	"runtime"

	"strings"

	"github.com/Jeffail/gabs"
	"github.com/streadway/amqp"
)

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
	messages = append(messages, m)
	log.Infof("message [ %s ] was added", m.Name)
	initMessages()
	return nil
}
func updateMessage(m message) (err error) {
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
func initConsumerManager() {
	wrapedConsumers := make(map[string]*manageConsumer)
	go func() {
		log.Debugf("Consumer Manager started")
		for {
			wrapedConsumer := <-consumerManageReadChan
			log.Debugf("revecived consumer[%s]", wrapedConsumer.action)
			c := &manageConsumer{
				wrapedConsumer:    wrapedConsumer,
				consumerReadChan:  make(chan string, 1),
				consumerWriteChan: make(chan string, 1),
				key:               getConsumerKey(wrapedConsumer.message, wrapedConsumer.consumer),
			}
			switch wrapedConsumer.action {
			case "update":
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
						for {
						RETRY:
							if _, ok := wrapedConsumers[c.key]; !ok {
								log.Warn("consumer[ %s ] not found , now exit", c.key)
								runtime.Goexit()
							}
							//update consumer active time
							wrapedConsumers[c.key].lasttime = time.Now().Unix()
							conn, err := pools.Get()
							if err != nil {
								pools.Put(conn)
								log.Warn("get conn fail for consumer , sleep 3 seconds ... %s", err)
								time.Sleep(time.Second * 3)
								continue
							}
							channel, err := conn.(*amqp.Connection).Channel()
							if err != nil {
								pools.Put(conn)
								log.Warn("get channel fail for consumer , sleep 3 seconds ... %s", err)
								time.Sleep(time.Second * 3)
								continue
							}
							err = channel.Qos(1, 0, false)
							if err != nil {
								pools.Put(conn)
								log.Warn("set channel Qos fail for consumer , sleep 3 seconds ... %s", err)
								time.Sleep(time.Second * 3)
								continue
							}
							deliveryChn, err := channel.Consume(c.key, "", false, false, false, false, nil)
							if err != nil {
								pools.Put(conn)
								log.Warn("get deliveryChn fail for consumer , sleep 3 seconds ...%s", err)
								time.Sleep(time.Second * 3)
								continue
							}
							for {
								if _, ok := wrapedConsumers[c.key]; !ok {
									log.Warn("consumer[ %s ] not found , now exit", c.key)
									runtime.Goexit()
								}
								//update consumer active time
								wrapedConsumers[c.key].lasttime = time.Now().Unix()
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
										log.Warn("get delivery fail for consumer")
										goto RETRY
									}
									//process success
									err := delivery.Ack(false)
									//process fail
									//err := delivery.Nack(false, true)
									if err != nil {
										log.Warnf("consumer %s ack/nack fail , %s", c.key, err)
									}
									log.Debugf("delivery revecived [ %s ] : %s", c.key, string(delivery.Body))
									time.Sleep(time.Second * 10)
								}
							}
						}
					}()
				}
				consumerManageWriteChan <- "completed"
			case "delete":
				//stop consumer
				if _, ok := wrapedConsumers[c.key]; ok {
					log.Debugf("sending stop singal to consumer[%s] goroutine ...", c.key)
					wrapedConsumers[c.key].consumerReadChan <- "exit"
					log.Debugf("send  stop singal to consumer[%s] goroutine success", c.key)
				}
				consumerManageWriteChan <- "completed"
			case "status":
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
