package model

import "time"

type CommandResponse struct {
	Uuid string `json:"uuid"`
}

type DeactivateNodePayload struct {
	Certname          string    `json:"certname"`
	ProducerTimestamp time.Time `json:"producer_timestamp"`
}
