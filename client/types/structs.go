package types

import "crypto/ed25519"

// ## 🧩 Structs (All JSON format)

// ### 🔸 `ModCert`
// ```json
// {
//   "Sign": "string",
//   "Pub_key": "string",
//   "Status": "string(approved or rejected)"
// }
// ```

// ### 🔸 `MsgCert`
// ```json
// {
//   "Public_key": "string",
//   "Msg": {
//     "message": "string", // In Msg changed it from message to content(chage it in doc)
//     "ts": 1234567890123
//   },
//   "ts": 1234567890123,  // later remove this ts property also from document
//   "Modcert": ["array of ModCert"]
//   "sign": "string"
// }
// ```

type Msg struct {
	Content string `json:"content"`
	Ts      int64  `json:"ts"`
}

type ModCert struct {
	Sign      string            `json:"sign"`
	PublicKey ed25519.PublicKey `json:"public_key"`
	Status    string            `json:"status"`
}

type MsgCert struct {
	PublicKey ed25519.PublicKey `json:"public_key"`
	Msg       Msg               `json:"msg"`
	ModCerts  []ModCert         `json:"modCerts"`
	Sign      string            `json:"sign"`
}

type DataToSign struct {
	Content   string    `json:"message"`
	Timestamp int64     `json:"timestamp"`
	ModCerts  []ModCert `json:"modcerts"`
}

type Mod struct {
	IP        string            `json:"ip"`
	Port      string            `json:"port"`
	PublicKey ed25519.PublicKey `json:"public_key"`
}
