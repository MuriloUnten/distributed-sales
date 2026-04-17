package main

import (
	"encoding/json"
	"os"

	"github.com/MuriloUnten/distributed-sales/common"
	"github.com/charmbracelet/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	logger = log.New(os.Stdout)
)

func main() {
	connection, err := amqp.Dial(common.Url)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	ch, err := connection.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		common.LogsExchangeName,
		"topic",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	queue, err := ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = ch.QueueBind(queue.Name, "#", common.LogsExchangeName, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	messages, err := ch.Consume(queue.Name, "", true, true, false, false, nil)

	var forever chan struct{}

	go listen(messages)

	log.Printf("Waiting for messages")
	<-forever
}

func listen(messages <-chan amqp.Delivery) {
	connection, err := amqp.Dial(common.Url)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	ch, err := connection.Channel()
	if err != nil {
		log.Fatal(err)
	}

	defer ch.Close()
	for msg := range messages {
		handleMessage(&msg)
		msg.Ack(false)
	}
}

func handleMessage(msg *amqp.Delivery) {
		var logMsg common.LogMessage
		err := json.Unmarshal(msg.Body, &logMsg)
		if err != nil {
			logger.Error("error parsing log message: ", err)
			return
		}

		switch msg.RoutingKey {
		case common.InfoKey:
			logger.Infof("[%s] (%s): %s\n", logMsg.Timestamp, logMsg.Sender, logMsg.Payload)
		case common.WarningKey:
			logger.Warnf("[%s] (%s): %s\n", logMsg.Timestamp, logMsg.Sender, logMsg.Payload)
		case common.ErrorKey:
			logger.Errorf("[%s] (%s): %s\n", logMsg.Timestamp, logMsg.Sender, logMsg.Payload)
		case common.DebugKey:
			logger.Debugf("[%s] (%s): %s\n", logMsg.Timestamp, logMsg.Sender, logMsg.Payload)

		default:
			logger.Error("received message with invalid type: ", msg.RoutingKey)
		}
}
