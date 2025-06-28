package util

import (
	"time"

	"github.com/devlup-labs/Libr/core/client/types"
)

var Stored []types.StoredMsg

func InitMockDB() {
	curTime := time.Now().Unix()
	Prev := []types.StoredMsg{
		{
			PublicKey: "PubKeySender1",
			Content:   "Hello this is message sent first",
			Timestamp: curTime - 180,
		},
		{
			PublicKey: "PubKeySender2",
			Content:   "Hello this is message sent second",
			Timestamp: curTime - 120,
		},
		{
			PublicKey: "PubKeySender3",
			Content:   "Hello this is message sent third",
			Timestamp: curTime - 60,
		},
		{
			PublicKey: "PubKeySender2",
			Content:   "Hello this is message sent fourth",
			Timestamp: curTime - 60,
		},
		{
			PublicKey: "PubKeySender3",
			Content:   "Hello this is message sent fifth",
			Timestamp: curTime + 30,
		},
	}
	Stored = append(Stored, Prev...)
}

func Store(MsgCert types.MsgCert) {
	New := types.StoredMsg{
		PublicKey: MsgCert.PublicKey,
		Content:   MsgCert.Msg.Content,
		Timestamp: MsgCert.Msg.Ts,
	}
	Stored = append(Stored, New)
}
