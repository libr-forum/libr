package models

type MsgCert struct {
    Sender    string     `json:"sender"`
    Msg       string     `json:"msg"`
    Timestamp string     `json:"ts"`       
    ModCert   []ModCert  `json:"mod_cert"` 
}

type ModCert struct {
    PublicKey string `json:"public_key"`
    Sign      string `json:"sign"`
    Status    string `json:"status"`    
}
