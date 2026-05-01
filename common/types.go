package common

import "encoding/json"

type NotificationEvent string

const (
	CreatedEvent NotificationEvent = "created"
	PopularEvent NotificationEvent = "hot deal"
)

type SignedMessage struct {
	Signature string          `json:"signature"`
	Payload   json.RawMessage `json:"payload"`
}

type SalePayload struct {
	Name string `json:"name"`
}

type VoteMessage struct {
	Name     string `json:"name"`
	// Note (Murilo): Positive was put here to allow a downvote in the future
	Positive bool   `json:"positive"`
}

type NotificationMessage struct {
	Event NotificationEvent `json:"message"`
}

type LogMessage struct {
	Timestamp string `json:"timestamp"`
	Sender    string `json:"sender"`
	Payload   string `json:"payload"`
}
