package common

import (
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)


type DistributedLogger struct {
	Sender string
	ch *amqp.Channel
	conn *amqp.Connection
}

func (l *DistributedLogger) Info(s string) {
	err := l.send(
		InfoKey,
		LogMessage{
			Payload: s,
			Sender: l.Sender,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)

	if err != nil {
		log.Println(err)
	}
}

func (l *DistributedLogger) Warn(s string) {
	err := l.send(
		WarningKey,
		LogMessage{
			Payload: s,
			Sender: l.Sender,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	if err != nil {
		log.Println(err)
	}
}

func (l *DistributedLogger) Error(s string) {
	err := l.send(
		ErrorKey,
		LogMessage{
			Payload: s,
			Sender: l.Sender,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	if err != nil {
		log.Println(err)
	}
}

func (l *DistributedLogger) Debug(s string) {
	err := l.send(
		DebugKey,
		LogMessage{
			Payload: s,
			Sender: l.Sender,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	)
	if err != nil {
		log.Println(err)
	}
}

func (l *DistributedLogger) send(msgType string, msg LogMessage) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return l.ch.Publish(
		LogsExchangeName,
		msgType,
		false,
		false,
		amqp.Publishing{
			Body: bytes,
		},
	)
}

func (l *DistributedLogger) Disconnect() {
	l.ch.Close()
	l.conn.Close()
}

/*
  This opens a new connection, please call Disconnect on the returned DistributedLogger
  after use is finished 
*/
func ConnectToLoggingService(sender string) (*DistributedLogger, error) {
	connection, err := amqp.Dial(Url)
	if err != nil {
		log.Fatal(err)
	}

	ch, err := connection.Channel()
	if err != nil {
		log.Fatal(err)
	}

	err = ch.ExchangeDeclare(
		LogsExchangeName,
		"topic",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	
	logger := DistributedLogger{
		Sender: sender,
		ch: ch,
		conn: connection,
	}
	return &logger, nil
}
