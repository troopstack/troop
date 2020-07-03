package rmq

import (
	"fmt"
	"log"
	"time"

	"github.com/troopstack/troop/src/modules/scout/utils"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

var (
	conn          *amqp.Connection
	channel       *amqp.Channel
	connNotify    chan *amqp.Error
	channelNotify chan *amqp.Error
	quit          chan bool
	ExchangeName  = "scout"
	routing       = []string{"scout", "scout.server"}
	//accepted          = false
	retryConnInterval = 1 * time.Minute
	hostname          string
	queueName         string
)

func SetupRMQ(initConn bool) {
	hostname, err := utils.Hostname()
	if err != nil {
		hostname = fmt.Sprintf("error:%s", err.Error())
		panic(err)
	}
	queueName = "scout.server." + hostname
	//routing = []string{"scout", "scout.server"}
	//if accepted {
	//	routing = append(routing, "scout."+hostname)
	//	routing = append(routing, queueName)
	//}
	if channel == nil {
		mq, err := utils.MQ()
		if err != nil {
			failOnError(err, "")
			panic(err)
		}
		rmqAddr := "amqp://" + mq.User + ":" + mq.Password + "@" + mq.Host + ":" + mq.Port + "/" + mq.VHost

		// 连接RabbitMQ
		conn, err = amqp.Dial(rmqAddr) // 建立连接
		if err != nil {
			retryConnNum := 0
			log.Println("Failed to connect to RabbitMQ", err)
			log.Printf("wait retry connect to RabbitMQ...")
			for {
				//if (initConn) {
				//	select {
				//	case <-utils.HandshakeChan:
				//		break
				//	}
				//}

				if retryConnNum > 3 {
					time.Sleep(retryConnInterval)
				}

				// 连接RabbitMQ
				conn, err = amqp.Dial(rmqAddr) // 建立连接
				if err == nil {
					log.Printf("connect to RabbitMQ successfully")
					break
				} else {
					log.Println("Failed to connect to RabbitMQ", err)
					log.Printf("wait retry connect to RabbitMQ...")
					retryConnNum++
				}
			}
		}
		// 打开通道
		channel, err = conn.Channel() // 创建channel
		failOnError(err, "Failed to open a channel")
		if err != nil {
			panic(err)
		}

		err = channel.ExchangeDeclare(
			ExchangeName,       // name
			amqp.ExchangeTopic, // type
			true,               // durable
			false,              // auto-deleted
			false,              // internal
			false,              // no-wait
			nil,                // arguments
		)
		failOnError(err, "Failed to declare an exchange")
		if err != nil {
			panic(err)
		}

		q, err := channel.QueueDeclare(
			queueName, // name
			false,     // durable
			true,      // delete when usused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		failOnError(err, "Failed to declare a queue")
		if err != nil {
			panic(err)
		}

		for r := range routing {
			QueueBind(routing[r])
		}

		msgs, err := channel.Consume(
			q.Name, // queue
			"",     // consumer
			true,   // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		failOnError(err, "Failed to register a consumer")
		if err != nil {
			panic(err)
		}

		go Handle(msgs)

		//if accepted {
		connNotify = conn.NotifyClose(make(chan *amqp.Error))
		channelNotify = channel.NotifyClose(make(chan *amqp.Error))
		//}
	}
	if initConn {
		go ReConnect()
	}
	return
}

func Handle(delivery <-chan amqp.Delivery) {
	for d := range delivery {
		go MessageProcess(d.Body)
	}
}

func QueueBind(routing string) {
	defer utils.CoverErrorMessage()
	err := channel.QueueBind(
		queueName,    // queue name
		routing,      // routing key
		ExchangeName, // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")
}

//func closeRMQ() {
//	defer utils.CoverErrorMessage()
//	conn.Close()
//}

func ReSetupRMQ() {
	//closeRMQ()
	//channel = nil
	//accepted = true
	//SetupRMQ(false)
	routing = append(routing, "scout."+hostname)
	routing = append(routing, queueName)
	//if !conn.IsClosed() {
	QueueBind("scout." + hostname)
	QueueBind(queueName)
	//}
}

func ReConnect() {
	for {
		select {
		case err := <-connNotify:
			if err != nil {
				failOnError(err, "rabbitmq consumer - connection NotifyClose")
			}
		case err := <-channelNotify:
			if err != nil {
				failOnError(err, "rabbitmq consumer - channel NotifyClose")
			}
		case <-quit:
			return
		}

		// backstop
		if !conn.IsClosed() {
			// 关闭 SubMsg message delivery
			if err := channel.Cancel("", true); err != nil {
				failOnError(err, "rabbitmq consumer - channel cancel failed")
			}

			if err := conn.Close(); err != nil {
				failOnError(err, "rabbitmq consumer - conn close failed")
			}
		}

		// IMPORTANT: 必须清空 Notify，否则死连接不会释放
		for err := range channelNotify {
			println(err)
		}
		for err := range connNotify {
			println(err)
		}

	quit:
		for {
			select {
			case <-quit:
				return
			default:
				log.Print("rabbitmq consumer - reconnect")
				channel = nil
				conn.Close()
				SetupRMQ(false)
				break quit
			}
		}
	}
}
