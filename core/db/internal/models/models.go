package models

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

type RetMsgCert struct {
	PublicKey string    `json:"public_key"`
	Msg       Msg       `json:"msg"`
	ModCerts  []ModCert `json:"mod_certs"`
	Sign      string    `json:"sign"`
	Deleted   string    `json:"deleted"`
}

type DataToSign struct {
	Content  string    `json:"message"`
	Ts       int64     `json:"timestamp"`
	ModCerts []ModCert `json:"mod_certs"`
}

type Mod struct {
	IP        string `json:"ip"`
	Port      string `json:"port"`
	PublicKey string `json:"public_key"`
}

type ReportMsg struct {
	PublicKey string `json:"public_key"`
	Msg       Msg    `json:"msg"`
}

type ReportCert struct {
	Msgcert     MsgCert   `json:"msgcert"`
	RepModCerts []ModCert `json:"repmod_certs"`
	Mode        string    `json:"mode"`
}
type DeleteCert struct {
	PublicKey string `json:"public_key"`
	Msg       Msg    `json:"msg"`
	Sign      string `json:"sign"`
}

type Node struct {
	NodeId    [20]byte `json:"nodeid"`
	IP        string   `json:"ip"`
	Port      string   `json:"port"`
	LastSeen  int64    `json:"lastseen"`
	PublicKey string   `json:"public_key"`
}

type KBucket struct {
	Nodes []*Node
}
