package main

import (
	"log"
	"time"

	"github.com/snail007/pool"
	"github.com/streadway/amqp"
)

type mqConnection struct {
	conn          *amqp.Connection
	isManualClose bool
}

const (
	poolInitialCap  = 5
	poolMaxCap      = 30
	consumePoolName = "consume"
)

var (
	pools = make(map[string]pool.Pool)
	uri   = "amqp://gome:gome@10.125.207.4:5672/"
)

func connectToRabbitMQ(uri string) *amqp.Connection {
	for {
		conn, err := amqp.Dial(uri)

		if err == nil {
			return conn
		}

		log.Println(err)
		log.Printf("Trying to reconnect to RabbitMQ at %s\n", uri)
		time.Sleep(500 * time.Millisecond)
	}
}
func getConnection(poolName string) (mqConnection, pool.Pool, error) {
	if _, ok := pools[poolName]; ok {
		c, e := pools[poolName].Get()
		failOnError(e, "fail pools[poolName].Get() in ok")
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
					log.Printf("Connecting to %s\n", uri)
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
	failOnError(e, "fail pool.NewChannelPool")
	pools[poolName] = p
	c, e := pools[poolName].Get()
	failOnError(e, "fail pools[poolName].Get()")
	return c.(mqConnection), pools[poolName], e
}
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
func main() {
	// var conn, err = amqp.Dial("amqp://gome:gome@10.125.207.4:5672/")
	// failOnError(err, "Failed to connect to RabbitMQ")
	// defer conn.Close()
	queueDeclare("test", false, true, false, false, nil)
	queueDeclare("test", true, false, false, false, nil)
	queueDeclare("test", true, false, false, false, nil)
	queueDeclare("test", true, false, false, false, nil)
	queueDeclare("test", true, false, false, false, nil)
	queueDeclare("test", true, false, false, false, nil)
	queueDeclare("test", true, false, false, false, nil)
	queueDeclare("test", true, false, false, false, nil)
	time.Sleep(time.Second * 1000)
	// err3 := ch.ExchangeDeclare("hello", "fanout", false, true, false, false, nil)
	// failOnError(err3, "Failed err3")
	// ch.Close()
	// ch, err1 = conn.Channel()
	// failOnError(err1, "Failed to open a channel")
	// ch.QueueBind("hello", "", "hello", false, nil)
	// ch.QueueDeclare()
	// var _, er = ch.QueueDeclare(
	// 	"hello", // name
	// 	false,   // durable
	// 	true,    // delete when unused
	// 	false,   // exclusive
	// 	false,   // no-wait
	// 	nil,     // arguments
	// )
	// failOnError(er, "Failed to declare a queue")
	// ch.Close()
	// ch, err1 = conn.Channel()
	// failOnError(err1, "Failed to open a channel")
	// var _, er3 = ch.Consume("hello", "", true, false, true, false, nil)
	// failOnError(er3, "Failed to er3")
	// time.Sleep(time.Second * 1000)
}

func queueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) amqp.Queue {
	var queue amqp.Queue
	var err error
	var mqConn mqConnection
	var channel *amqp.Channel
	var poolConsme pool.Pool
	for {
		mqConn, poolConsme, err = getConnection(consumePoolName)
		poolConsme.Put(mqConn)
		failOnError(err, "getConnection("+consumePoolName+")")
		if err != nil {
			continue
		}
		channel, err = mqConn.conn.Channel()
		defer channel.Close()
		failOnError(err, "mqConn.conn.Channel()")
		queue, err = channel.QueueDeclarePassive(name, durable, autoDelete, exclusive, noWait, args)
		failOnError(err, "channel.QueueDeclarePassive")
		mqConn, poolConsme, err := getConnection(consumePoolName)
		poolConsme.Put(mqConn)
		failOnError(err, "getConnection("+consumePoolName+") [for delete]")
		channel1, err1 := mqConn.conn.Channel()
		failOnError(err1, "mqConn.conn.Channel() [for delete]")
		defer channel1.Close()
		channel1.QueueDelete(name, false, false, false)
		mqConn, poolConsme, err = getConnection(consumePoolName)
		poolConsme.Put(mqConn)
		failOnError(err, "getConnection("+consumePoolName+") [for return]")
		queue, err = channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
		failOnError(err, "channel.QueueDeclare("+consumePoolName+") [for return]")
		if err == nil {
			return queue
		}
	}

}
