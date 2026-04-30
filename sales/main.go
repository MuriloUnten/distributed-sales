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
)

func main() {
	l, err := common.ConnectToLoggingService("sales")
	if err != nil {
		log.Fatal("failed to connect to logging service")
	}
	logger = l
	defer logger.Disconnect()

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
		common.ExchangeName,
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

	err = ch.QueueBind(queue.Name, common.ReceivedKey , common.ExchangeName, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	messages, err := ch.Consume(queue.Name, "", false, true, false, false, nil)

	var forever chan struct{}

	go listen(messages)

	log.Printf("Waiting for messages")
	<-forever
}

func listen(messages <-chan amqp.Delivery) {
	key, err := common.LoadPrivateKeyFromFile("./keys/private/private_key.pem")
	if err != nil {
		log.Fatal("cannot continue due to failure loading private key: ", err)
	}

	registeredPubKeys, err := common.LoadPublicKeysFromDirectory("./keys/public")
	if err != nil {
		log.Fatal("cannot continue due to failure loading public keys: ", err)
	}

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
		handleMessage(msg.Body, ch, key, registeredPubKeys)
		msg.Ack(false)
	}
}

func handleMessage(msg []byte, ch *amqp.Channel, privateKey *rsa.PrivateKey, registeredPubKeys []*rsa.PublicKey) {
	signedMessage := new(common.SignedMessage)
	err := json.Unmarshal(msg, signedMessage)
	if err != nil {
		logger.Error("error decoding signed message: " + err.Error())
		return
	}
	fmt.Println(signedMessage)

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

	signature, err := common.Sign(privateKey, crypto.SHA256, signedMessage.Payload)
	if err != nil {
		logger.Error("dropping message due to failure signing: " + err.Error())
		return
	}

	outputMessage := common.SignedMessage{
		Signature: signature,
		Payload: signedMessage.Payload,
	}

	outputBytes, err := json.Marshal(outputMessage)
	if err != nil {
		logger.Error("dropping message due to failure encoding output: " + err.Error())
		return
	}

	err = ch.Publish(
		common.ExchangeName,
		common.PublishedKey,
		false,
		false,
		amqp.Publishing{
			Body: outputBytes,
		},
	)
	if err != nil {
		logger.Error("failed to publish sale: " + err.Error())
		return
	}
}
