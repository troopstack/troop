package rmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/general/utils"

	"github.com/streadway/amqp"
)

type Service struct {
	AmqpUrl       string // Amqp地址
	ConnectionNum int    // 连接数
	ChannelNum    int    // 每个连接的channel数量

	connections    map[int]connection
	channels       map[int]channel
	idleChannels   []int
	busyChannels   map[int]int
	m              *sync.Mutex
	wg             *sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	connectIdChan  chan int
	lockConnectIds map[int]bool
	temChannel     chan int
}

type connection struct {
	conn       *amqp.Connection
	connNotify chan *amqp.Error
	quit       chan bool
}

type channel struct {
	ch            *amqp.Channel
	connectId     int
	notifyClose   chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
}

const (
	retryCount        = 5
	waitConfirmTime   = 5 * time.Second
	retryConnInterval = 1 * time.Minute
)

var AmqpServer Service

func InitAmqp() {
	mq, err := utils.MQ()
	if err != nil {
		failOnError(err, "")
	}
	AmqpServer.AmqpUrl = "amqp://" + mq.User + ":" + mq.Password + "@" + mq.Host + ":" + mq.Port + "/" + mq.VHost

	if AmqpServer.ConnectionNum == 0 {
		AmqpServer.ConnectionNum = mq.MaxConnectionNum
	}
	if AmqpServer.ChannelNum == 0 {
		AmqpServer.ChannelNum = mq.MaxChannelNum
	}
	AmqpServer.m = new(sync.Mutex)
	AmqpServer.wg = new(sync.WaitGroup)
	AmqpServer.ctx, AmqpServer.cancel = context.WithTimeout(context.Background(), waitConfirmTime)
	AmqpServer.lockConnectIds = make(map[int]bool)
	AmqpServer.connectIdChan = make(chan int)
	AmqpServer.busyChannels = make(map[int]int)
	AmqpServer.temChannel = make(chan int, 100)

	AmqpServer.connectPool()
	AmqpServer.channelPool()
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

func AvailableConnNum() (num int) {
	num = 0
	for index := range AmqpServer.connections {
		if !AmqpServer.connections[index].conn.IsClosed() {
			num++
		}
	}
	return
}

func (S *Service) connectPool() {
	S.connections = make(map[int]connection)
	for i := 0; i < S.ConnectionNum; i++ {
		connection, err := S.connect()
		if err != nil {
			for {
				connection, err = S.connect()
				if err == nil {
					log.Printf("retry connect to RabbitMQ successfully")
					break
				} else {
					log.Println("Failed to connect to RabbitMQ", err)
					log.Printf("wait retry connect to RabbitMQ...")
					time.Sleep(retryConnInterval)
				}
			}
		}
		S.connections[i] = connection
		go S.ReConnect(i)
	}
}

func (S *Service) channelPool() {
	S.channels = make(map[int]channel)
	for index, _ := range S.connections {
		for j := 0; j < S.ChannelNum; j++ {
			key := index*S.ChannelNum + j
			S.channels[key] = S.createChannel(index)
			S.idleChannels = append(S.idleChannels, key)
		}
	}
}

func (S *Service) connect() (connection, error) {
	connAMQP, err := amqp.Dial(S.AmqpUrl)
	conn := connection{
		conn: connAMQP,
	}

	if err == nil {
		conn.connNotify = connAMQP.NotifyClose(make(chan *amqp.Error))
	}
	return conn, err
}

func (S *Service) ReConnect(i int) {
	connNum := i + 1
	for {
		select {
		case err := <-S.connections[i].connNotify:
			if err != nil {
				failOnError(err, "rabbitmq - connection NotifyClose")
			}
		case <-S.connections[i].quit:
			return
		}

		// backstop
		if !S.connections[i].conn.IsClosed() {
			if err := S.connections[i].conn.Close(); err != nil {
				failOnError(err, "rabbitmq - channel cancel failed")
			}
		}

		// IMPORTANT: 必须清空 Notify，否则死连接不会释放
		for err := range S.connections[i].connNotify {
			println(err)
		}

	quit:
		for {
			select {
			case <-S.connections[i].quit:
				return
			default:
				log.Printf("[RabbitMQ Connect Number %d] reconnect", connNum)
				S.connections[i].conn.Close()
				connection, err := S.connect()
				if err != nil {
					for {
						connection, err = S.connect()
						if err == nil {
							log.Printf("[RabbitMQ Connect Number %d] retry connect to RabbitMQ successfully", connNum)
							break
						} else {
							log.Printf("[RabbitMQ Connect Number %d] Failed to connect to RabbitMQ %s", connNum, err)
							log.Printf("[RabbitMQ Connect Number %d] wait retry connect to RabbitMQ...", connNum)
							time.Sleep(retryConnInterval)
						}
					}
				}
				S.connections[i] = connection
				for c := range S.channels {
					if S.channels[c].connectId == i {
						S.channels[c] = S.createChannel(i)
					}
				}
				break quit
			}
		}
	}
}

func (S *Service) recreateChannel(connectId int, err error) (ch *amqp.Channel) {
	if strings.Index(err.Error(), "channel/connection is not open") >= 0 || strings.Index(err.Error(), "CHANNEL_ERROR - expected 'channel.open'") >= 0 {
		if S.connections[connectId].conn.IsClosed() {
			S.lockWriteConnect(connectId)
		}
		ch, err = S.connections[connectId].conn.Channel()
		failOnError(err, "Failed to open a channel")
	} else {
		failOnError(err, "Failed to open a channel")
	}
	return
}

func (S *Service) lockWriteConnect(connectId int) {

	S.m.Lock()
	if !S.lockConnectIds[connectId] {
		S.lockConnectIds[connectId] = true
		S.m.Unlock()

		go func(connectId int) {
			S.wg.Add(1)
			defer S.wg.Done()

			S.connections[connectId], _ = S.connect()
			S.connectIdChan <- connectId

		}(connectId)
	} else {
		S.m.Unlock()
	}

	for {
		select {
		case cid := <-S.connectIdChan:

			delete(S.lockConnectIds, cid)

			if len(S.lockConnectIds) == 0 {
				S.wg.Wait()
				return
			} else {
				continue
			}
		case <-time.After(waitConfirmTime):
			S.lockConnectIds = make(map[int]bool)
			S.wg.Wait()
			return
		}
	}
}

func (S *Service) createChannel(connectId int) channel {
	var notifyClose = make(chan *amqp.Error, AmqpServer.ConnectionNum*AmqpServer.ChannelNum)
	var notifyConfirm = make(chan amqp.Confirmation, AmqpServer.ConnectionNum*AmqpServer.ChannelNum)

	cha := channel{
		connectId:     connectId,
		notifyClose:   notifyClose,
		notifyConfirm: notifyConfirm,
	}
	if S.connections[connectId].conn.IsClosed() {
		S.lockWriteConnect(connectId)
	}
	ch, err := S.connections[connectId].conn.Channel()
	if err != nil {
		ch = S.recreateChannel(connectId, err)
	}
	ch.Confirm(false)
	ch.NotifyClose(cha.notifyClose)
	ch.NotifyPublish(cha.notifyConfirm)

	S.NotifyReturn(func(message amqp.Return) {
		taskMsg := model.ScoutMessage{}
		err := json.Unmarshal(message.Body, &taskMsg)
		//log.Printf(" [%s] unreachable", message.RoutingKey)
		if err == nil {
			if taskMsg.Type == "task" || taskMsg.Type == "ping" || taskMsg.Type == "plugin" || taskMsg.Type == "fileManage" || taskMsg.Type == "bala_task" {
				utils.TaskReturn(taskMsg, message.RoutingKey)
			}
		}
	}, ch)

	cha.ch = ch
	return cha
}

func (S *Service) NotifyReturn(notifier func(message amqp.Return), ch *amqp.Channel) {
	go func() {
		for res := range ch.NotifyReturn(make(chan amqp.Return)) {
			notifier(res)
		}
	}()
}

func (S *Service) getChannel() (*amqp.Channel, int) {
	S.m.Lock()
	defer S.m.Unlock()
	idleLength := len(S.idleChannels)
	if idleLength > 0 {
		rand.Seed(time.Now().Unix())
		index := rand.Intn(idleLength)
		channelId := S.idleChannels[index]
		S.idleChannels = append(S.idleChannels[:index], S.idleChannels[index+1:]...)
		S.busyChannels[channelId] = channelId

		ch := S.channels[channelId].ch
		return ch, channelId
	} else {
		//return S.createChannel(0,S.connections[0]),-1
		return nil, -1
	}
}

func (S *Service) declareExchange(ch *amqp.Channel, exchangeName string, channelId int) *amqp.Channel {
	done := make(chan *amqp.Channel, 1)

	go func() {
		// 用于检查交换机是否存在,已经存在不需要重复声明
		err := ch.ExchangeDeclarePassive(
			exchangeName,
			amqp.ExchangeTopic,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			err := ch.ExchangeDeclare(
				exchangeName,       // name
				amqp.ExchangeTopic, // type
				true,               // durable
				false,              // auto-deleted
				false,              // internal
				false,              // no-wait
				nil,                // arguments
			)
			if err != nil {
				ch = S.reDeclareExchange(channelId, exchangeName, err)
			}
		}
		done <- ch
	}()

	select {
	case doCh := <-done:
		return doCh
	case <-time.After(time.Duration(1 * time.Second)):
		return ch
	}
}

func (S *Service) reDeclareExchange(channelId int, exchangeName string, err error) (ch *amqp.Channel) {

	var connectionId int
	if strings.Index(err.Error(), "channel/connection is not open") >= 0 {

		if channelId == -1 {
			rand.Seed(time.Now().Unix())
			index := rand.Intn(S.ConnectionNum)
			connectionId = index
		} else {
			connectionId = int(channelId / S.ChannelNum)
		}
		cha := S.createChannel(connectionId)

		S.lockWriteChannel(channelId, cha)
		err := cha.ch.ExchangeDeclare(
			exchangeName,       // name
			amqp.ExchangeTopic, // type
			true,               // durable
			false,              // auto-deleted
			false,              // internal
			false,              // no-wait
			nil,                // arguments
		)
		if err != nil {
			failOnError(err, "Failed to declare an exchange")
		}
		return cha.ch
	} else {
		failOnError(err, "Failed to declare an exchange")
		return nil
	}
}

func (S *Service) lockWriteChannel(channelId int, cha channel) {
	S.m.Lock()
	defer S.m.Unlock()
	S.channels[channelId] = cha
}

func (S *Service) dataForm(notice interface{}) string {
	body, err := json.Marshal(notice)
	if err != nil {
		log.Panic(err)
	}
	return string(body)
}

func (S *Service) publish(channelId int, ch *amqp.Channel, exchangeName string, routeKey string, data string, priority uint8) (err error) {
	err = ch.Publish(
		exchangeName,
		routeKey,
		true,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(data),
			Priority:     priority,
		})

	if err != nil {
		if strings.Index(err.Error(), "channel/connection is not open") >= 0 {
			err = S.rePublish(channelId, exchangeName, err, routeKey, data, priority)
		}
	}

	return
}

func (S *Service) rePublish(channelId int, exchangeName string, errmsg error, routeKey string, data string, priority uint8) (err error) {

	ch := S.reDeclareExchange(channelId, exchangeName, errmsg)
	err = ch.Publish(
		exchangeName, // exchange
		routeKey,     //severityFrom(os.Args), // routing key
		true,         // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(data),
			Priority:     priority,
		})
	return
}

func (S *Service) backChannelId(channelId int, ch *amqp.Channel) {
	if channelId == -1 {
		return
	}
	S.m.Lock()
	defer S.m.Unlock()
	S.idleChannels = append(S.idleChannels, channelId)
	delete(S.busyChannels, channelId)
	return
}

func (S *Service) PutIntoQueue(exchangeName string, routeKey string, notice interface{}, priority uint8) (message interface{}, pubErr error) {
	defer func() {
		msg := recover()
		if msg != nil {
			fmt.Println("msg: ", msg)
			pubErrorMsg, _ := msg.(string)
			fmt.Println("pubErrorMsg : ", pubErrorMsg)
			pubErr = errors.New(pubErrorMsg)
			return
		}
	}()
	ch, channelId := S.getChannel()
	cha := S.channels[channelId]
	if ch == nil {
		rand.Seed(time.Now().Unix())
		index := rand.Intn(S.ConnectionNum)
		cha = S.createChannel(index)
		defer func() {
			cha.ch.Close()
		}()
		ch = cha.ch
	}
	ch = S.declareExchange(ch, exchangeName, channelId)
	if channelId != -1 {
		defer func() {
			cha.notifyClose = make(chan *amqp.Error, AmqpServer.ConnectionNum*AmqpServer.ChannelNum)
			cha.notifyConfirm = make(chan amqp.Confirmation, AmqpServer.ConnectionNum*AmqpServer.ChannelNum)
			S.backChannelId(channelId, ch)
		}()
	}
	data := S.dataForm(notice)
	var tryTime = 1
	for {
		pubErr = S.publish(channelId, ch, exchangeName, routeKey, data, priority)
		if pubErr != nil {
			if tryTime <= retryCount {
				log.Printf("%s: %s", "Failed to publish a message, try again.", pubErr)
				tryTime++
				continue
			} else {
				log.Printf("%s: %s data: %s", "Failed to publish a message", pubErr, data)
				return notice, pubErr
			}
		}
		select {
		case confirm := <-cha.notifyConfirm:
			if confirm.Ack {
				log.Printf(" [%s] Sent %d message %s", routeKey, confirm.DeliveryTag, data)
				return notice, nil
			}
		case <-time.After(waitConfirmTime):
			log.Printf(" [%s] message: %s data: %s", "Can not receive the confirm.", routeKey, data)
			confirmErr := errors.New("can not receive the confirm")
			return notice, confirmErr
		}
	}
}
