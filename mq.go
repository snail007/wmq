package main

import (
	"net"
	"sync"

	logger "github.com/snail007/mini-logger"
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
	log.With(logger.Fields{"func": "getMqConnection", "call": "pools.Get"}).Errorf("fail,%s", err)
	return
}
func getMqChannel() (channel *amqp.Channel, err error) {
	var c interface{}
	c, err = channelPools.Get()
	if err == nil {
		channel = c.(*amqp.Channel)
		return
	}
	log.With(logger.Fields{"func": "getMqChannel", "call": "channelPools.Get"}).Errorf("fail,%s", err)
	return
}
func getQueueName(queueName string) string {
	return cfg.GetString("rabbitmq.prefix") + queueName
}
func getExchangeName(exchangeName string) string {
	return cfg.GetString("rabbitmq.prefix") + exchangeName
}
func queueDeclare(name string, durable bool) (queue amqp.Queue, channel *amqp.Channel, err error) {
	name = getQueueName(name)
	ctx := ctxFunc("queueDeclare").With(logger.Fields{"queue": name})

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
			ctx.Debug("declare success")
			return
		}
		channel, err = getMqChannel()
		ctx1 := ctx.With(logger.Fields{"subCall": "channel.getMqChannel"})
		if err == nil {
			_, err = channel.QueueDelete(name, false, false, false)
			channelPools.Put(channel)
			ctx2 := ctx.With(logger.Fields{"subSubCall": "channel.QueueDelete"})
			if err != nil {
				ctx2.Errorf("fail,%s", err)
			} else {
				ctx2.Debug("delete success")
			}
			retryCount++
			goto RETRY
		} else {
			channelPools.Put(channel)
			ctx1.Errorf("fail,%s", err)
		}
	} else {
		channelPools.Put(channel)
		ctx.With(logger.Fields{"call": "channel.getMqChannel"}).Errorf("declare fail,%s", err)
	}
	return
}

func exchangeDeclare(name, kind string, durable bool) (channel *amqp.Channel, err error) {
	name = getExchangeName(name)
	ctx := ctxFunc("exchangeDeclare").With(logger.Fields{"exchange": name})
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
	ctx1 := ctx.With(logger.Fields{"call": "getMqChannel"})
	if err == nil {
		err = channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, nil)
		ctx2 := ctx1.With(logger.Fields{"subCall": "channel.ExchangeDeclare"})
		channelPools.Put(channel)
		if err == nil {
			ctx2.Debug("declare success")
			return
		}
		channel, err = getMqChannel()
		ctx2 = ctx1.With(logger.Fields{"subCall": "getMqChannel"})
		if err == nil {
			err = channel.ExchangeDelete(name, false, false)
			ctx3 := ctx1.With(logger.Fields{"subSubCall": "channel.ExchangeDelete"})
			channelPools.Put(channel)
			if err != nil {
				ctx3.Errorf("fail,%s", err)
			} else {
				ctx3.Debug("success")
			}
			retryCount++
			goto RETRY
		} else {
			channelPools.Put(channel)
			ctx2.Errorf("fail,%s", err)
		}
	} else {
		channelPools.Put(channel)
		ctx1.Errorf("fail,%s", err)
	}
	return
}

func queueBindToExchange(queuename, exchangeName, routeKey string) (err error) {
	queuename = getQueueName(queuename)
	exchangeName = getExchangeName(exchangeName)
	ctx := ctxFunc("queueBindToExchange").With(logger.Fields{"queue": queuename, "exchange": exchangeName})
	var channel *amqp.Channel
	channel, err = getMqChannel()
	defer func() { channelPools.Put(channel) }()
	if err == nil {
		err = channel.QueueBind(queuename, routeKey, exchangeName, false, nil)
		if err == nil {
			ctx.Debugf("success")
			return
		}
	}
	ctx.Errorf("fail,%s", err)
	return
}

func deleteQueue(queueName string) (err error) {
	queueName = getQueueName(queueName)
	ctx := ctxFunc("deleteQueue").With(logger.Fields{"queue": queueName})
	var channel *amqp.Channel
	defer func() { channelPools.Put(channel) }()
	channel, err = getMqChannel()
	if err == nil {
		_, err = channel.QueueDelete(queueName, false, false, false)
		ctx1 := ctx.With(logger.Fields{"call": "channel.QueueDelete"})
		if err != nil {
			ctx1.Errorf("fail,%s", err)
		} else {
			ctx1.Debug("success")
		}
	} else {
		ctx.Errorf("fail,%s", err)
	}
	return
}
func deleteExchange(exchangeName string) (err error) {
	exchangeName = getExchangeName(exchangeName)
	ctx := ctxFunc("deleteExchange").With(logger.Fields{"exchange": exchangeName})
	var channel *amqp.Channel
	channel, err = getMqChannel()
	defer func() { channelPools.Put(channel) }()
	if err == nil {
		err = channel.ExchangeDelete(exchangeName, false, false)
		ctx1 := ctx.With(logger.Fields{"call": "channel.ExchangeDelete"})
		if err != nil {
			ctx1.Errorf("fail,%s", err)
		} else {
			ctx1.Debug("success")
			return
		}
	} else {
		ctx.Errorf("fail,%s", err)
	}
	return
}

func initPool() (err error) {
	ctx := ctxFunc("initPool")
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
					// if err := c.SetDeadline(time.Now().Add(mqConnectionAndDeadlineTimeout)); err != nil {
					// 	return nil, err
					// }
					return c, nil
				},
			})
			if err == nil {
				ctx.Debugf("Connect to RabbitMQ SUCCESS")
				return
			}
			ctx.Debugf("Connect to RabbitMQ FAIL,ERR:%s", err)
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
	ctx := ctxFunc("initChannelPool")
	poolcfg := poolConfig{
		InitialCap: poolChannelInitialCap,
		MaxCap:     poolChannelMaxCap,
		Release: func(conn interface{}) {
			if conn != nil && conn.(*amqp.Channel) != nil {
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
					ctx.Debugf("Channel Create  SUCCESS")
					return
				}
			}
			ctx.Debugf("Channel Create FAIL")
			return
		},
		IsActive: func(conn interface{}) (ok bool) {
			if conn == nil {
				return false
			}
			ch, ok := conn.(*amqp.Channel)
			if !ok || ch == nil {
				return false
			}
			err := ch.Tx()
			if err != nil {
				//log.Infof("channel is not active %s", err)
				return false
			}
			ch.TxCommit()
			//log.Debugf("channel is active")
			return true
		},
	}
	channelPools, err = newNetPool(poolcfg)
	return
}
