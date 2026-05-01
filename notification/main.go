package main

import (
	"crypto"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"

	"github.com/MuriloUnten/distributed-sales/common"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	logger *common.DistributedLogger = nil
	votes map[string]int = nil
	privateKey *rsa.PrivateKey
)

func main() {
	votes = make(map[string]int)

	key, err := common.LoadPrivateKeyFromFile("./keys/private/private_key.pem")
	privateKey = key
	if err != nil {
		log.Fatal("cannot continue due to failure loading private key: ", err)
	}

	listener, err := common.InitListener(common.PopularKey)
	if err != nil {
		log.Fatal("error starting listener: ", err)
	}
	defer listener.Deinit()
	err = listener.Ch.QueueBind(listener.Queue.Name, common.PublishedKey, common.ExchangeName, false, nil)
	if err != nil {
		log.Fatal("error starting listener: ", err)
	}

	sender, err := common.InitSender()
	if err != nil {
		log.Fatal("error starting sender: ", err)
	}
	defer sender.Deinit()

	logger, err := common.ConnectToLoggingService("notification")
	if err != nil {
		log.Fatal("failed to connect to logging service")
	}
	defer logger.Disconnect()

	messages, err := listener.Ch.Consume(listener.Queue.Name, "", false, true, false, false, nil)

	var forever chan struct{}

	go listen(messages, sender)

	log.Printf("Waiting for messages")
	<-forever
}

func listen(messages <-chan amqp.Delivery, sender *common.RabbitMQSender) {

	registeredPubKeys, err := common.LoadPublicKeysFromDirectory("./keys/public")
	if err != nil {
		log.Fatal("cannot continue due to failure loading public keys: ", err)
	}

	for msg := range messages {
		handleMessage(msg.Body, msg.RoutingKey, registeredPubKeys, sender)
		msg.Ack(false)
	}
}

func handleMessage(msg []byte, routingKey string, registeredPubKeys []*rsa.PublicKey, sender *common.RabbitMQSender) {
	signedMessage := new(common.SignedMessage)
	err := json.Unmarshal(msg, signedMessage)
	if err != nil {
		logger.Error("error decoding signed message: " + err.Error())
		return
	}

	sale := new(common.SalePayload)
	err = json.Unmarshal(signedMessage.Payload, sale)
	if err != nil {
		logger.Error("error decoding sale message: " + err.Error())
		return
	}

	validated := common.ValidateSignature(signedMessage.Signature, signedMessage.Payload, registeredPubKeys)
	if !validated {
		logger.Info("dropping message due to failed validation")
		return
	}

	fmt.Printf("%+v\n", sale)

	var payload common.NotificationMessage
	switch routingKey {
		case common.PublishedKey:
			payload = common.NotificationMessage{
				Event: common.CreatedEvent,
			}
		case common.PopularKey:
			payload = common.NotificationMessage{
				Event: common.PopularEvent,
			}
		default:
			logger.Warn("Received from unexpected routing key: " + routingKey)
			return
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error(err.Error())
	}
	signature, err := common.Sign(privateKey, crypto.SHA256, payloadBytes)
	signed := common.SignedMessage{
		Signature: signature,
		Payload: payloadBytes,
	}
	signedMsgBytes, err := json.Marshal(signed)
	if err != nil {
		logger.Error(err.Error())
	}

	err = sender.Ch.Publish(
		common.ExchangeName,
		sale.Name,
		false,
		false,
		amqp.Publishing{
			Body: signedMsgBytes,
		},
	)
	if err != nil {
		logger.Error(err.Error())
	}
}
