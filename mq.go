package main

import (
	"net"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

var (
	pools, channelPools ConnPool
)

type mqconn struct {
	conn    *amqp.Connection
	ctlchan chan string
}

var lock = &sync.Mutex{}

func getMqConnection() (conn *amqp.Connection, err error) {
	var c interface{}
	c, err = pools.Get()
	if err == nil {
		conn = c.(*amqp.Connection)
		return
	}
	log.Errorf("fail pools.Get():%s", err)
	return
}
func getMqChannel() (channel *amqp.Channel, err error) {
	var c interface{}
	c, err = channelPools.Get()
	if err == nil {
		channel = c.(*amqp.Channel)
		return
	}
	log.Errorf("channelPools.Get() fail in exists channelPools:%s", err)
	return
}

func queueDeclare(name string, durable bool) (queue amqp.Queue, channel *amqp.Channel, err error) {
	autoDelete, exclusive, noWait := true, false, false
	if durable {
		autoDelete = false
	}
	maxRetryCount := 1
	retryCount := 0
RETRY:
	if retryCount > maxRetryCount {
		return
	}
	channel, err = getMqChannel()
	if err == nil {
		queue, err = channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, nil)
		channelPools.Put(channel)
		if err == nil {
			log.Debug("[" + name + "] queueDeclare success")
			return
		}
		channel, err = getMqChannel()
		if err == nil {
			_, err = channel.QueueDelete(name, false, false, false)
			channelPools.Put(channel)
			if err != nil {
				log.Errorf("channel.QueueDelete("+name+", false, false, false):%s", err)
			} else {
				log.Debug("[" + name + "] QueueDelete success")
			}
			retryCount++
			goto RETRY
		} else {
			channelPools.Put(channel)
			log.Errorf("getMqChannel [for QueueDelete]:%s", err)
		}
	} else {
		channelPools.Put(channel)
		log.Errorf("getMqChannel [for QueueDeclare]:%s", err)
	}
	return
}

func exchangeDeclare(name, kind string, durable bool) (channel *amqp.Channel, err error) {
	autoDelete, internal, noWait := true, false, false
	if durable {
		autoDelete = false
	}
	maxRetryCount := 1
	retryCount := 0
RETRY:
	if retryCount > maxRetryCount {
		return
	}
	channel, err = getMqChannel()
	if err == nil {
		err = channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, nil)
		channelPools.Put(channel)
		if err == nil {
			log.Debug("[" + name + "] ExchangeDeclare success")
			return
		}
		channel, err = getMqChannel()
		if err == nil {
			err = channel.ExchangeDelete(name, false, false)
			channelPools.Put(channel)
			if err != nil {
				log.Errorf("channel.ExchangeDelete("+name+", false, false):%s", err)
			} else {
				log.Debug("[" + name + "] ExchangeDelete success")
			}
			retryCount++
			goto RETRY
		} else {
			channelPools.Put(channel)
			log.Errorf("getMqChannel [for ExchangeDelete]:%s", err)
		}
	} else {
		channelPools.Put(channel)
		log.Errorf("getMqChannel [for ExchangeDeclare]:%s", err)
	}
	return
}

func queueBindToExchange(queuename, exchangeName, routeKey string) (err error) {
	var channel *amqp.Channel
	channel, err = getMqChannel()
	defer func() { channelPools.Put(channel) }()
	if err == nil {
		err = channel.QueueBind(queuename, routeKey, exchangeName, false, nil)
		if err == nil {
			log.Debugf("queueBindToExchange [ %s->%s ] success", queuename, exchangeName)
			return
		}
	}
	log.Errorf("queueBindToExchange [ %s->%s ] fail", queuename, exchangeName)
	return
}

func deleteQueue(queueName string) (err error) {
	var channel *amqp.Channel
	defer func() { channelPools.Put(channel) }()
	channel, err = getMqChannel()
	if err == nil {
		_, err = channel.QueueDelete(queueName, false, false, false)
		if err != nil {
			log.Errorf("deleteQueue->channel.QueueDelete("+queueName+", false, false, false):%s", err)
		} else {
			log.Debug("[" + queueName + "] deleteQueue->QueueDelete success")
		}
	} else {
		log.Errorf("deleteQueue->getMqChannel:%s", err)
	}
	return
}
func deleteExchange(exchangeName string) (err error) {
	var channel *amqp.Channel
	channel, err = getMqChannel()
	defer func() { channelPools.Put(channel) }()
	if err == nil {
		err = channel.ExchangeDelete(exchangeName, false, false)
		if err != nil {
			log.Errorf("channel.ExchangeDelete("+exchangeName+", false, false):%s", err)
		} else {
			log.Debug("[" + exchangeName + "] ExchangeDelete success")
			return
		}
	} else {
		log.Errorf("deleteQueue->getMqChannel:%s", err)
	}
	return
}

func initPool() (err error) {
	poolcfg := poolConfig{
		InitialCap: poolInitialCap,
		MaxCap:     poolMaxCap,
		Release: func(conn interface{}) {
			if conn != nil {
				conn.(*amqp.Connection).Close()
				conn = nil
			}
		},
		Factory: func() (retConn interface{}, err error) {
			retConn, err = amqp.DialConfig(uri, amqp.Config{
				Heartbeat: mqHeartbeat,
				Dial: func(network, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(network, addr, mqConnectionAndDeadlineTimeout)
					if err != nil {
						return nil, err
					}
					if err := c.SetDeadline(time.Now().Add(mqConnectionAndDeadlineTimeout)); err != nil {
						return nil, err
					}
					return c, nil
				},
			})
			if err == nil {
				log.Debugf("Connect to RabbitMQ SUCCESS")
				return
			}
			log.Debugf("Connect to RabbitMQ FAIL")
			return
		},
		IsActive: func(conn interface{}) (ok bool) {
			if conn == nil {
				return false
			}
			ch, err := conn.(*amqp.Connection).Channel()
			if err != nil {
				//log.Debugf("conn is not active")
				return false
			}
			ch.Close()
			//log.Debugf("conn is active")
			return true
		},
	}
	pools, err = newNetPool(poolcfg)
	return
}
func initChannelPool() (err error) {
	poolcfg := poolConfig{
		InitialCap: poolChannelInitialCap,
		MaxCap:     poolChannelMaxCap,
		Release: func(conn interface{}) {
			if conn != nil {
				//log.Errorf("channel was released,%s", conn)
				conn.(*amqp.Channel).Close()
				conn = nil
			}
		},
		Factory: func() (retConn interface{}, err error) {
			conn, err := pools.Get()
			defer pools.Put(conn)
			if err == nil {
				retConn, err = conn.(*amqp.Connection).Channel()
				if err == nil {
					log.Debugf("Channel Create  SUCCESS")
					return
				}
			}
			log.Debugf("Channel Create FAIL")
			return
		},
		IsActive: func(conn interface{}) (ok bool) {
			if conn == nil {
				return false
			}
			err := conn.(*amqp.Channel).Tx()
			if err != nil {
				//log.Infof("channel is not active %s", err)
				return false
			}
			conn.(*amqp.Channel).TxCommit()
			//log.Debugf("channel is active")
			return true
		},
	}
	channelPools, err = newNetPool(poolcfg)
	return
}
