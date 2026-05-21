package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"

	"github.com/MuriloUnten/distributed-sales/common"
	amqp "github.com/rabbitmq/amqp091-go"
)

var interests = [...]string{"banana", "abacaxi", "abacate"}

func main() {
	listener, err := common.InitListener(common.PopularKey)
	if err != nil {
		log.Fatal("error starting listener: ", err)
	}
	defer listener.Deinit()

	for _, interest := range interests {
		routingKey := "promocao." + interest
		err := listener.Ch.QueueBind(listener.Queue.Name, routingKey, common.ExchangeName, false, nil)
		if err != nil {
			log.Fatal("error starting listener: ", err)
		}
	}

	messages, err := listener.Ch.Consume(listener.Queue.Name, "", false, true, false, false, nil)

	var forever chan struct{}

	go listen(messages)

	log.Printf("Waiting for messages")
	<-forever
}

func listen(messages <-chan amqp.Delivery) {

	registeredPubKeys, err := common.LoadPublicKeysFromDirectory("./keys/public")
	if err != nil {
		log.Fatal("cannot continue due to failure loading public keys: ", err)
	}

	for msg := range messages {
		handleMessage(msg.Body, msg.RoutingKey, registeredPubKeys)
		msg.Ack(false)
	}
}

func handleMessage(msg []byte, routingKey string, registeredPubKeys []*rsa.PublicKey) {
	signedMessage := new(common.SignedMessage)
	err := json.Unmarshal(msg, signedMessage)
	if err != nil {
		log.Println("error decoding signed message: " + err.Error())
		return
	}

	notification := new(common.NotificationMessage)
	err = json.Unmarshal(signedMessage.Payload, notification)
	if err != nil {
		log.Println("error decoding sale message: " + err.Error())
		return
	}

	validated := common.ValidateSignature(signedMessage.Signature, signedMessage.Payload, registeredPubKeys)
	if !validated {
		log.Println("dropping message due to failed validation")
		return
	}

	fmt.Printf("Received at routing key %s: %+v\n", routingKey, notification.Event)
}
