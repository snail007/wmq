package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"sync"

	"time"

	"github.com/Jeffail/gabs"
	"github.com/snail007/pool"
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

var (
	messageStartLocks   = make(map[string]*sync.Mutex)
	getLocker           = new(sync.Mutex)
	startMessageSignals = make(chan message, 1)
	consumerSignals     = make(map[string]chan string)
)

func initMessages() {
	for _, msg := range messages {
		messageDeclare(msg)
	}
	go func() {
		for {
			m := <-startMessageSignals
			err := messageDeclare(m)
			if err == nil {
				log.Debugf("start message success , exchange:%s", m.Name)
				for _, c := range m.Consumers {
					restartConsumer(m, c)
				}
			}
			log.Debugf("message restarted [%s]", m.Name)
			time.Sleep(time.Second * 5)
		}
	}()
}
func pushMessageStarSingal(msg message) bool {
	select {
	case startMessageSignals <- msg:
		return true
	default:
		return false
	}
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
func getMessage(name string) (msg *message, err error) {
	for _, m := range messages {
		if m.Name == name {
			msg = &m
			return
		}
	}
	return nil, errors.New("message not found")
}

func messageDeclare(msg message) (err error) {
	_, err = exchangeDeclare(msg.Name, msg.Mode, msg.Durable, publishPoolName, 1)
	if err == nil {
		log.Debugf("exchangeDeclare success :%s", msg.Name)
		for _, c := range msg.Consumers {
			var queueName = getQueueName(msg, c)
			_, _, err = queueDeclare(queueName, msg.Durable, consumePoolName, 1)
			if err == nil {
				log.Debugf("queueDeclare success , exchange:%s->queue:%s", msg.Name, queueName)
				err = queueBindToExchange(queueName, msg.Name, c.RouteKey)
				if err != nil {
					log.Errorf("queueBindToExchange :%s", err)
					break
				} else {
					log.Debugf("queueBindToExchange success , queue:%s->exchange:%s", queueName, msg.Name)
					go restartConsumer(msg, c)
				}
			} else {
				log.Errorf("queueDeclare , exchange:%s->queue:%s", msg.Name, queueName)
				break
			}
		}
	} else {
		log.Errorf("exchangeDeclare [ %s ]: %s ", msg.Name, err)
	}
	return
}

/**
删除一个消费者,需要进行下面的清理工作
1、停止消费协程
2、删除consumerSignals中的控制信号
3、删除messages里面的消费者
4、删除消费者对应的queue
**/
func deleteConsumer(msg message, c0 consumer) (err error) {
	//1、停止消费协程
	stopConsumer(msg, c0)
	time.Sleep(time.Millisecond * 100)
	//2、删除consumerSignals中的控制信号
	delete(consumerSignals, getQueueName(msg, c0))
	//3、删除messages里面的消费者
STOP:
	for k1, m := range messages {
		if m.Name == msg.Name {
			for k2, c := range m.Consumers {
				if c.ID == c0.ID {
					messages[k1].Consumers = append(messages[k1].Consumers[0:k2], messages[k1].Consumers[k2+1:]...)
					log.Debugf("consumer [ %s ] was deleted from messages", getQueueName(msg, c))
					//4、删除消费者对应的queue
					err = deleteQueue(getQueueName(m, c), 2)
					break STOP
				}
			}
		}
	}

	log.Infof("consumer [ %s ] was deleted.", getQueueName(msg, c0))
	return
}

//增加一个消费者，需要做如下工作
//1、添加消费者到messages中
//2、同步messages到文件
//3、申明对应的queue
//4、绑定queue到exchange
//5、调用restartConsumer启动

func addConsumer(messageName string, c consumer) (err error) {
	for k, m := range messages {
		if m.Name == messageName {
			messages[k].Consumers = append(messages[k].Consumers, c)
			//2、同步到文件
			//3、申明对应的queue
			_, _, err = queueDeclare(getQueueName(m, c), m.Durable, consumePoolName, 2)

			if err == nil {
				//4、绑定queue到exchange
				err = queueBindToExchange(getQueueName(m, c), m.Name, c.RouteKey)
				if err == nil {
					//5、启动
					restartConsumer(m, c)
					log.Debugf("Consumer added , %s->%s", m.Name, c.ID)
					return
				}
			}
			break
		}
	}
	log.Errorf("Consumer added fail , %s->%s", messageName, c.ID)
	err = errors.New("message " + messageName + " not found")
	return
}
func stopConsumer(msg message, c consumer) {
	queueName := getQueueName(msg, c)
	if singal, ok := consumerSignals[queueName]; ok {
		select {
		case singal <- "exit":
		default:
		}
	}

}
func restartConsumer(msg message, c consumer) {
	queueName := getQueueName(msg, c)
	var singal chan string
	var ok bool
	if singal, ok = consumerSignals[queueName]; !ok {
		singal = make(chan string, 1)
		consumerSignals[queueName] = singal
	}
	if ok {
		stopConsumer(msg, c)
		time.Sleep(time.Millisecond * 100)
	}
	var p pool.Pool
	var channel *amqp.Channel
	var err error
	go func() {
	EXIT:
		for {
			channel, p, err = getMqChannel(consumePoolName, 0)
			if err == nil {
				msgChan, e := channel.Consume(queueName, "", false, false, false, false, nil)
				if e == nil {
				RESTART:
					for {
						log.Infof("Consumer [ %s ] Waiting for message...", queueName)
						select {
						case msg := <-msgChan:
							if msg.ConsumerTag == "" {
								log.Warnf("Consumer [ %s ] Unavaliable Message recieved : %s", queueName, msg)
								p.Put(channel)
								break RESTART
							}
							log.Infof("Consumer [ %s ] Message recieved : %s", queueName, msg.Body)
							var processedOk = false
							log.Infof("process msg")
							processedOk = true
							if processedOk {
								channel.Ack(msg.DeliveryTag, false)
								log.Infof("process msg success")
							} else {
								channel.Nack(msg.DeliveryTag, false, true)
								log.Infof("process msg fail")
							}
							time.Sleep(time.Second * 10)
						case cmd := <-singal:
							log.Debugf("Consumer [ %s ] Control Command recieved : %s", queueName, cmd)
							if cmd == "exit" {
								p.Put(channel)
								break EXIT
							}
						}
					}
				} else {
					log.Errorf("Consumser Consume : %s", e)
					time.Sleep(time.Second * 3)
				}
			} else {
				log.Errorf("Consumer worker : %s", err)
				time.Sleep(time.Second * 3)
			}
		}
		log.Warnf("Consumer [ %s ] exited", queueName)
	}()
}
func getQueueName(msg message, c consumer) string {
	return msg.Name + "-" + c.ID
}
func publish(body, exchangeName, routeKey, token string, failRetryCount int) (err error) {
	var channelPool pool.Pool
	tryCount := 0
	var msg *message
	msg, err = getMessage(exchangeName)
	if err != nil {
		return
	}
	if msg.IsNeedToken && token != msg.Token {
		err = errors.New("token error")
		return
	}
	var channel *amqp.Channel
	for {
		if failRetryCount >= 0 && tryCount > failRetryCount {
			return
		}
		channel, channelPool, err = getMqChannel(publishPoolName, 0)
		if err == nil {
			err = channel.Publish(exchangeName, routeKey, false, false, amqp.Publishing{
				Body: []byte(body),
			})
			channelPool.Put(channel)
			if err == nil {
				log.Debugf("publish message success , exchange:%s", exchangeName)
				return
			}
			pushMessageStarSingal(*msg)
			time.Sleep(time.Second * 3)
		}
		if failRetryCount > 0 {
			tryCount++
		}
	}
}
