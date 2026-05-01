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

type UiActionChoice int

const (
	HOT_DEAL_THRESHOLD = 3
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

	listener, err := common.InitListener(common.VoteKey)
	if err != nil {
		log.Fatal("error starting listener: ", err)
	}
	defer listener.Deinit()

	sender, err := common.InitSender()
	if err != nil {
		log.Fatal("error starting sender: ", err)
	}
	defer sender.Deinit()

	logger, err := common.ConnectToLoggingService("ranking")
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
		handleMessage(msg.Body, registeredPubKeys, sender)
		msg.Ack(false)
	}
}

func handleMessage(msg []byte, registeredPubKeys []*rsa.PublicKey, sender *common.RabbitMQSender) {
	signedMessage := new(common.SignedMessage)
	err := json.Unmarshal(msg, signedMessage)
	if err != nil {
		logger.Error("error decoding signed message: " + err.Error())
		return
	}

	vote := new(common.VoteMessage)
	err = json.Unmarshal(signedMessage.Payload, vote)
	if err != nil {
		logger.Error("error decoding vote message: " + err.Error())
		return
	}

	validated := common.ValidateSignature(signedMessage.Signature, signedMessage.Payload, registeredPubKeys)
	if !validated {
		logger.Info("dropping message due to failed validation")
		return
	}

	count, _ := votes[vote.Name]
	newValue := count + 1
	votes[vote.Name] = count + 1
	fmt.Printf("Casting vote to: %s | Total: %d\n", vote.Name, newValue)

	if newValue == HOT_DEAL_THRESHOLD {
		fmt.Println("New popular sale: ", vote.Name)
		sendPopular(vote.Name, sender)
	}
}

func sendPopular(name string, sender *common.RabbitMQSender) error {
	payload := common.SalePayload{
		Name: name,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	signature, err := common.Sign(privateKey, crypto.SHA256, payloadBytes)
	signed := common.SignedMessage{
		Signature: signature,
		Payload: payloadBytes,
	}
	signedMsgBytes, err := json.Marshal(signed)
	if err != nil {
		return err
	}

	return sender.Ch.Publish(
		common.ExchangeName,
		common.PopularKey,
		false,
		false,
		amqp.Publishing{
			Body: signedMsgBytes,
		},
	)
}
