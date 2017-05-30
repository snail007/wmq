package main

import (
	"net"
	"time"

	"github.com/snail007/pool"
	"github.com/streadway/amqp"
)

type mqConnection struct {
	conn          *amqp.Connection
	isManualClose bool
}

var (
	pools        = make(map[string]pool.Pool)
	channelPools = make(map[string]pool.Pool)
)

func connectToRabbitMQ(failRetryCount int) (conn *amqp.Connection, err error) {
	tryCount := 0
	for {
		conn, err = amqp.DialConfig(uri, amqp.Config{
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
			log.Debugf("Connect to RabbitMQ at %s success", uri)
			return
		}
		if failRetryCount >= 0 {
			tryCount++
		}
		if failRetryCount >= 0 && tryCount > failRetryCount {
			return
		}
		log.Errorf("Trying to reconnect to RabbitMQ at %s:%s", uri, err)
		time.Sleep(mqConnectionFailRetrySleep)
	}
}
func getMqConnection(poolName string, failRetryCount int) (conn mqConnection, pool0 pool.Pool, err error) {
	var exists bool
	tryCount := 0
	var c interface{}
	for {
		if failRetryCount >= 0 && tryCount > failRetryCount {
			return
		}
		if pool0, exists = pools[poolName]; exists {
			c, err = pools[poolName].Get()
			if err == nil {
				conn = c.(mqConnection)
				return
			}
			log.Errorf("pools[poolName].Get() fail in exists pool:%s", err)
		}

		factory := func() (interface{}, error) {
			c, err := connectToRabbitMQ(failRetryCount)
			if err == nil {
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
							log.Errorf("connection closed unexpectedly , retrying connecting to %s, error:%s", uri, rabbitErr)
							mqConn.conn, err = connectToRabbitMQ(-1)
							errchn = make(chan *amqp.Error)
							mqConn.conn.NotifyClose(errchn)
						}
					}
				}(mqConn)
				return mqConn, err
			}
			return nil, err
		}

		close := func(v interface{}) error {
			mqConnection := v.(mqConnection)
			mqConnection.isManualClose = true
			return mqConnection.conn.Close()
		}
		poolConfig := &pool.PoolConfig{
			InitialCap:  poolInitialCap,
			MaxCap:      poolMaxCap,
			Factory:     factory,
			AutoClose:   false,
			Close:       close,
			IdleTimeout: 15 * time.Second,
		}

		pool0, err = pool.NewChannelPool(poolConfig)
		if err == nil {
			log.Debug("init new pool : " + poolName)
			pools[poolName] = pool0
			c, err = pools[poolName].Get()
			if err == nil {
				conn = c.(mqConnection)
				return
			}
			log.Errorf("fail pools[poolName].Get():%s", err)
		} else {
			log.Errorf("fail pool.NewChannelPool(poolConfig):%s", err)
		}
		if failRetryCount >= 0 {
			tryCount++
		}
	}
}
func getMqChannel(poolName string, failRetryCount int) (channel *amqp.Channel, pool0 pool.Pool, err error) {
	var exists bool
	var connPool pool.Pool
	tryCount := 0
	var c interface{}
	for {
		if failRetryCount >= 0 && tryCount > failRetryCount {
			return
		}
		if pool0, exists = channelPools[poolName]; exists {
			c, err = channelPools[poolName].Get()
			if err == nil {
				channel = c.(*amqp.Channel)
				return
			}
			log.Errorf("channelPools[poolName].Get() fail in exists channelPools:%s", err)
		}

		factory := func() (channel interface{}, err error) {
			c, connPool, err = getMqConnection(poolName, 0)
			if err == nil {
				errchn := make(chan *amqp.Error)
				channel, err = c.(mqConnection).conn.Channel()
				connPool.Put(c)
				if err == nil {
					channel.(*amqp.Channel).NotifyClose(errchn)
					go func() {
						var rabbitErr *amqp.Error
						for {
							rabbitErr = <-errchn
							if rabbitErr != nil {
								log.Errorf("channel closed unexpectedly , retrying connecting , error:%s", rabbitErr)
								c, connPool, _ = getMqConnection(poolName, -1)
								errchn = make(chan *amqp.Error)
								channel, err = c.(mqConnection).conn.Channel()
								connPool.Put(c)
								if err == nil {
									channel.(*amqp.Channel).NotifyClose(errchn)
									log.Infof("channel reconnected success ")
									return
								}
							}
						}
					}()
				}
			}
			return
		}

		close := func(v interface{}) error {
			channel := v.(*amqp.Channel)
			return channel.Close()
		}
		poolConfig := &pool.PoolConfig{
			InitialCap:  poolChannelInitialCap,
			MaxCap:      poolChannelMaxCap,
			Factory:     factory,
			AutoClose:   false,
			Close:       close,
			IdleTimeout: 15 * time.Second,
		}

		pool0, err = pool.NewChannelPool(poolConfig)
		if err == nil {
			log.Debug("init new channel pool : " + poolName)
			channelPools[poolName] = pool0
			c, err = channelPools[poolName].Get()
			if err == nil {
				channel = c.(*amqp.Channel)
				return
			}
			log.Errorf("fail channelPools[poolName].Get():%s", err)
		} else {
			log.Errorf("fail channelPools pool.NewChannelPool(poolConfig):%s", err)
		}
		if failRetryCount >= 0 {
			tryCount++
		}
	}
}

func queueDeclare(name string, durable bool, poolName string, failRetryCount int) (queue amqp.Queue, channel *amqp.Channel, err error) {
	tryCount := 0
	var channelPool pool.Pool
	autoDelete, exclusive, noWait := true, false, false
	if durable {
		autoDelete = false
	}
	for {
		if failRetryCount >= 0 && tryCount > failRetryCount {
			return
		}
		channel, channelPool, err = getMqChannel(poolName, 0)
		if err == nil {
			queue, err = channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, nil)
			channelPool.Put(channel)
			if err == nil {
				log.Debug("[" + name + "] queueDeclare success")
				return
			}
			log.Debugf("channel.queueDeclare:%s", err)
			channel, channelPool, err = getMqChannel(poolName, 0)
			if err == nil {
				_, err = channel.QueueDelete(name, false, false, false)
				channelPool.Put(channel)
				if err != nil {
					log.Errorf("channel.QueueDelete("+name+", false, false, false):%s", err)
				} else {
					log.Debug("[" + name + "] QueueDelete success")
				}
			} else {
				log.Errorf("getMqChannel [for QueueDelete]:%s", err)
			}
		} else {
			log.Errorf("getMqChannel [for QueueDeclare]:%s", err)
		}
		if failRetryCount > 0 {
			tryCount++
		}
	}
}

func exchangeDeclare(name, kind string, durable bool, poolName string, failRetryCount int) (channel *amqp.Channel, err error) {
	tryCount := 0
	var channelPool pool.Pool
	autoDelete, internal, noWait := true, false, false
	if durable {
		autoDelete = false
	}
	for {
		if failRetryCount >= 0 && tryCount > failRetryCount {
			return
		}
		channel, channelPool, err = getMqChannel(poolName, 0)
		if err == nil {
			err = channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, nil)
			channelPool.Put(channel)
			if err == nil {
				log.Debug("[" + name + "] ExchangeDeclare success")
				return
			}
			log.Debugf("channel.ExchangeDeclare:%s", err)
			channel, channelPool, err = getMqChannel(poolName, 0)
			if err == nil {
				err = channel.ExchangeDelete(name, false, false)
				channelPool.Put(channel)
				if err != nil {
					log.Errorf("channel.ExchangeDelete("+name+", false, false):%s", err)
				} else {
					log.Debug("[" + name + "] ExchangeDelete success")
				}
			} else {
				log.Errorf("getMqChannel [for ExchangeDelete]:%s", err)
			}
		} else {
			log.Errorf("getMqChannel [for ExchangeDeclare]:%s", err)
		}
		if failRetryCount > 0 {
			tryCount++
		}
	}
}

func queueBindToExchange(queuename, exchangeName, routeKey string) (err error) {
	var channel *amqp.Channel
	var channelPool pool.Pool
	channel, channelPool, err = getMqChannel(publishPoolName, 3)
	if err == nil {
		err = channel.QueueBind(queuename, routeKey, exchangeName, false, nil)
		channelPool.Put(channel)
	}
	return
}

func deleteQueue(queueName string, failRetryCount int) (err error) {
	tryCount := 0
	var channel *amqp.Channel
	var channelPool pool.Pool
	for {
		if failRetryCount >= 0 && tryCount > failRetryCount {
			return
		}
		channel, channelPool, err = getMqChannel(consumePoolName, 0)
		if err == nil {
			_, err = channel.QueueDelete(queueName, false, false, false)
			channelPool.Put(channel)
			if err != nil {
				log.Errorf("deleteQueue->channel.QueueDelete("+queueName+", false, false, false):%s", err)
			} else {
				log.Debug("[" + queueName + "] deleteQueue->QueueDelete success")
				break
			}
		} else {
			log.Errorf("deleteQueue->getMqChannel:%s", err)
		}
		if failRetryCount > 0 {
			tryCount++
		}
	}
	return
}
