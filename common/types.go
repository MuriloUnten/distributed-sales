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
