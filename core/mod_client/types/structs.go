package types

import "time"

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
	Reason    string    `json:"reason,omitempty"`
}

type DataToSign struct {
	Content   string    `json:"message"`
	Timestamp int64     `json:"timestamp"`
	ModCerts  []ModCert `json:"mod_certs"`
}

type Mod struct {
	PeerId    string `json:"peer_id"`
	PublicKey string `json:"public_key"`
}

// type StoredMsg struct {
// 	PublicKey string `json:"public_key"`
// 	Content   string `json:"content"`
// 	Timestamp int64  `json:"timestamp"`
// }

type Node struct {
	NodeId [20]byte `json:"node_id"`
	PeerId string   `json:"peer_id"`
}

type ReportMsg struct {
	PublicKey string `json:"public_key"`
	Msg       Msg    `json:"msg"`
}

type SubmitMsg struct {
	Content string  `json:"content"`
	Ts      int64   `json:"ts"`
	Reason  *string `json:"reason,omitempty"`
	Mode    string  `json:"mode"`
	Sign    *string `json:"sign,omitempty"`
}

type ReportCert struct {
	Msgcert     MsgCert   `json:"msgcert"`
	RepModCerts []ModCert `json:"repmod_certs"`
	Mode        string    `json:"mode"`
}

type DeleteCert struct {
	PublicKey string    `json:"public_key"`
	ReportMsg ReportMsg `json:"report_msg"`
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

type PendingModeration struct {
	MsgSign      string    `json:"msg_sign"`      // cert.Sign
	MsgCert      MsgCert   `json:"msg_cert"`      // full original cert
	PartialCerts []ModCert `json:"partial_certs"` // modcerts already received
	AwaitingMods []string  `json:"awaiting_mods"` // public keys of mods yet to respond
	AckCount     int       `json:"ack_count"`
	CreatedAt    time.Time `json:"created_at"` // timestamp for cron retry
}
