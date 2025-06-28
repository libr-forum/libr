package core

import (
	"fmt"
	"time"

	util "github.com/devlup-labs/Libr/core/client/utils"
)

func FetchMsgAll() {
	for _, msg := range util.Stored {
		if msg.Timestamp < time.Now().Unix() {
			fmt.Printf("\nSender: %s\n%s\nTime:%d\n", msg.PublicKey, msg.Content, msg.Timestamp)
		}
	}
}

func Fetch(ts int64) {
	for _, msg := range util.Stored {
		if msg.Timestamp == ts {
			fmt.Printf("\nSender: %s\n%s\nTime:%d\n", msg.PublicKey, msg.Content, msg.Timestamp)
		}
	}
}
