package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"slices"

	"github.com/MuriloUnten/distributed-sales/common"
	amqp "github.com/rabbitmq/amqp091-go"
)

type UiActionChoice int

const (
	ActionCreate UiActionChoice = iota + 1
	ActionList
	ActionVote

	PUBLISHED_SALES_INITIAL_CAPACITY = 10
)

var (
	logger *common.DistributedLogger = nil
	publishedSales []string = nil
	privateKey *rsa.PrivateKey
)

func main() {
	publishedSales = make([]string, 0, PUBLISHED_SALES_INITIAL_CAPACITY)

	key, err := common.LoadPrivateKeyFromFile("./keys/private/private_key.pem")
	privateKey = key
	if err != nil {
		log.Fatal("cannot continue due to failure loading private key: ", err)
	}

	listener, err := InitListener()
	if err != nil {
		log.Fatal("error starting listener: ", err)
	}
	defer listener.Deinit()

	sender, err := InitSender()
	if err != nil {
		log.Fatal("error starting sender: ", err)
	}
	defer sender.Deinit()

	logger, err := common.ConnectToLoggingService("gateway")
	if err != nil {
		log.Fatal("failed to connect to logging service")
	}
	defer logger.Disconnect()

	messages, err := listener.ch.Consume(listener.queue.Name, "", true, true, false, false, nil)
	go listen(messages)


	runUi(sender)
}

func listen(messages <-chan amqp.Delivery) {

	registeredPubKeys, err := common.LoadPublicKeysFromDirectory("./keys/public")
	if err != nil {
		log.Fatal("cannot continue due to failure loading public keys: ", err)
	}

	for msg := range messages {
		handleMessage(msg.Body, registeredPubKeys)
		msg.Ack(false)
	}
}

func handleMessage(msg []byte, registeredPubKeys []*rsa.PublicKey) {
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

	alreadyExists := findSale(publishedSales, sale.Name)
	if alreadyExists {
		logger.Warn("Received published sale message for sale that already exists: " + sale.Name)
		return
	}

	publishedSales = append(publishedSales, sale.Name)
}

func runUi(sender *RabbitMQSender) {
	first := true
	for {
		if (!first) {
			fmt.Println("\nPress Enter to return...")
			fmt.Scanln()
		}
		first = false

		common.ClearTerminal()
		fmt.Println("Choose an action by entering its number")
		fmt.Printf("%d. Create a sale\n", ActionCreate)
		fmt.Printf("%d. List sales\n", ActionList)
		fmt.Printf("%d. Vote for a sale\n", ActionVote)

		var choice UiActionChoice
		_, err := fmt.Scanf("%d", &choice)
		if err != nil {
			fmt.Println("You entered an invalid action.")
			continue
		}

		switch choice {
		case ActionCreate:
			var saleName string
			common.ClearTerminal()
			fmt.Println("---- Create a sale ----")
			fmt.Printf("Enter the name of the sale you wish to create: ")
			_, err := fmt.Scanf("%s", &saleName)
			if err != nil {
				fmt.Println("Error reading sale name.")
				continue
			}

			alreadyExists := findSale(publishedSales, saleName)
			if alreadyExists {
				fmt.Println("Failed to create sale. A sale with the same name already exists.")
				continue
			}
			err = createSale(saleName, sender)
			if err != nil {
				fmt.Println("Error creating sale: ", err)
				continue
			}

		case ActionList:
			common.ClearTerminal()
			fmt.Println("---- List of sales ----")
			// Here we can print the main slice directly,
			// because there will be no subsequent writes
			printSales(publishedSales)
			continue

		case ActionVote:
			common.ClearTerminal()
			fmt.Println("---- Vote for sale ----")
			sales := getSales(publishedSales)
			if len(sales) == 0 {
				fmt.Println("There are no sales to vote to.")
				continue
			}
			printSales(sales)

			fmt.Printf("Enter the number of the sale you wish to vote to: ")
			var vote int
			_, err = fmt.Scanf("%d", &vote)
			if err != nil {
				fmt.Println("Failed to read sale number.")
				continue
			}

			if vote < 0 || vote >= len(sales) {
				fmt.Println("You entered an invalid sale number")
				continue
			}
			err = voteForSale(sales, vote, sender)
			if err != nil {
				fmt.Println("Error sending vote message: ", err)
			}

		default:
			continue
		}
	}
}

func createSale(name string, sender *RabbitMQSender) error {
	payload := common.SalePayload{
		Name: name,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256(payloadBytes)
	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, hashed[:])
	signatureString := base64.StdEncoding.EncodeToString(signature)
	signed := common.SignedMessage{
		Signature: signatureString,
		Payload: payloadBytes,
	}
	signedMsgBytes, err := json.Marshal(signed)
	if err != nil {
		return err
	}

	return sender.ch.Publish(
		common.ExchangeName,
		common.ReceivedKey,
		false,
		false,
		amqp.Publishing{
			Body: signedMsgBytes,
		},
	)
}

func printSales(sales []string) {
	if len(sales) == 0 {
		fmt.Println("")
		return
	}
	for id, sale := range sales {
		fmt.Printf("%d. %s\n", id, sale)
	}
	fmt.Println("")
}

/**
  when you wish to use the published sales,
  you must get a copy first in order to avoid concurrency problems in cases like voting
*/
func getSales(sales []string) []string {
	return slices.Clone(sales)
}

func voteForSale(sales []string, number int, sender *RabbitMQSender) error {
	// sanity check. This check guarantees the indexing won't cause a panic
	if number < 0 || number >= len(sales) {
		return fmt.Errorf("invalid sale number input")
	}

	chosenSale := sales[number]
	payload := common.VoteMessage{
		Name: chosenSale,
		Positive: true,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256(payloadBytes)
	signature, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, hashed[:])
	signed := common.SignedMessage{
		Signature: string(signature),
		Payload: payloadBytes,
	}
	signedMsgBytes, err := json.Marshal(signed)
	if err != nil {
		return err
	}

	return sender.ch.Publish(
		common.ExchangeName,
		common.VoteKey,
		false,
		false,
		amqp.Publishing{
			Body: signedMsgBytes,
		},
	)
}

func findSale(sales []string, name string) bool {
	for _, s := range publishedSales {
		if s == name {
			return true
		}
	}
	return false
}
