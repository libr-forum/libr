package models

type UserMsg struct {
	Content   string `json:"content"`
	TimeStamp int64  `json:"ts"`
}

type ModSign struct {
	Content   string `json:"content"`
	TimeStamp int64  `json:"timestamp"`
	Status    string `json:"status"`
}

type ModResponse struct {
	Sign      string `json:"sign"`
	PublicKey string `json:"public_key"`
	Status    string `json:"status"`
}
