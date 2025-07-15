package types

type Msg struct {
	Content string `json:"content"`
	Ts      int64  `json:"ts"`
}

type ModCert struct {
	Sign      string `json:"sign"`
	PublicKey string `json:"public_key"`
	Status    string `json:"status"`
}

type MsgCert struct {
	PublicKey string    `json:"public_key"`
	Msg       Msg       `json:"msg"`
	ModCerts  []ModCert `json:"mod_certs"`
	Sign      string    `json:"sign"`
}

type DataToSign struct {
	Content   string    `json:"message"`
	Timestamp int64     `json:"timestamp"`
	ModCerts  []ModCert `json:"modcerts"`
}

type Mod struct {
	IP        string `json:"ip"`
	Port      string `json:"port"`
	PublicKey string `json:"public_key"`
}

// type StoredMsg struct {
// 	PublicKey string `json:"public_key"`
// 	Content   string `json:"content"`
// 	Timestamp int64  `json:"timestamp"`
// }

type Node struct {
	NodeId [20]byte `json:"nodeid"`
	IP     string   `json:"ip"`
	Port   string   `json:"port"`
}
