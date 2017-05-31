package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"sync"

	"time"

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

var (
	messageStartLocks = make(map[string]*sync.Mutex)
	getLocker         = new(sync.Mutex)
	//startMessageSignals = make(chan message, 1)
	consumerSignals = make(map[string]chan string)
)

// func initMessages() {
// 	for _, msg := range messages {
// 		messageDeclare(msg)
// 	}
// 	go func() {
// 		for {
// 			m := <-startMessageSignals
// 			err := messageDeclare(m)
// 			if err == nil {
// 				log.Debugf("start message success , exchange:%s", m.Name)
// 				for _, c := range m.Consumers {
// 					restartConsumer(m, c)
// 				}
// 			}
// 			log.Debugf("message restarted [%s]", m.Name)
// 			time.Sleep(time.Second * 5)
// 		}
// 	}()
// }
// func pushMessageStarSingal(msg message) bool {
// 	select {
// 	case startMessageSignals <- msg:
// 		return true
// 	default:
// 		return false
// 	}
// }
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
func getMessage(name string) (msg *message, err error) {
	for _, m := range messages {
		if m.Name == name {
			msg = &m
			return
		}
	}
	return nil, errors.New("message not found")
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
					err = deleteQueue(getQueueName(m, c))
					break STOP
				}
			}
		}
	}

	log.Infof("consumer [ %s ] was deleted.", getQueueName(msg, c0))
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

func getQueueName(msg message, c consumer) string {
	return msg.Name + "-" + c.ID
}
func publish(body, exchangeName, routeKey, token string) (err error) {
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

func messagesMointor() {
	// controlChns:=
	go func() {
		for {
			for _, m := range messages {
				_, err := exchangeDeclare(m.Name, m.Mode, m.Durable)
				if err == nil {
					for _, c := range m.Consumers {
						_, _, err := queueDeclare(getQueueName(m, c), m.Durable)
						if err == nil {
							queueBindToExchange(getQueueName(m, c), m.Name, c.RouteKey)
						}
					}
				}

			}
			time.Sleep(time.Second * 5)
		}
	}()
}
