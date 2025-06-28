package models

type Msg struct {
	Content string `json:"content"`
	Ts      int64  `json:"ts"`
}

type ModSign struct {
	Content   string `json:"content"`
	TimeStamp string `json:"timestamp"`
	Status    string `json:"status"`
}

type ModResponse struct {
	Sign      string `json:"sign"`
	Status    string `json:"status"`
	PublicKey string `json:"pub_key"`
}
