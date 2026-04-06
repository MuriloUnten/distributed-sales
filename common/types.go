package common

import "encoding/json"

type SignedMessage struct {
	Signature string          `json:"signature"`
	Payload   json.RawMessage `json:"payload"`
}

type SalePayload struct {
	Name     string `json:"name"`
}

type VoteMessage struct {
	Name     string `json:"name"`
	Positive bool   `json:"positive"`
}

type LogMessage struct {
	Timestamp string `json:"timestamp"`
	Sender    string `json:"sender"`
	Payload   string `json:"payload"`
}
