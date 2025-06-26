package models

type UserMsg struct {
	Content   string `json:"content"`
	TimeStamp string `json:"timestamp"`
}

type ModSign struct {
	Content   string `json:"content"`
	TimeStamp string `json:"timestamp"`
	Status    string `json:"status"`
}

type ModResponse struct {
	Sign      string `json:"sign"`
	Status    string `json:"status"` //--> to check
	PublicKey string `json:"pub_key"`
}
