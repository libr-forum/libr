package main

import (
	"fmt"

	"github.com/devlup-labs/Libr/core/db/config"
	internal "github.com/devlup-labs/Libr/core/db/internal/msgcert"
	"github.com/devlup-labs/Libr/core/db/models"
)

func main() {
	config.InitConnection()
	msgcert := models.MsgCert{
		PublicKey: "Jl6u0CVdfVDfP9I56praRtqwn6uUuo4K3Wnt69aOwWo=",
		Msg: models.Msg{
			Content: "Hello",
			Ts:      1751113043,
		},
		ModCerts: []models.ModCert{{
			Sign:      "htkVG1bdGvKo+FPkTgmJL6XO+fWtk9Waz1OLIoq2/0ZP5DEPJaXCkOGqiaZo2eNqWqLLWmquyDrmbcSEDl+pAw==",
			PublicKey: "Jl6u0CVdfVDfP9I56praRtqwn6uUuo4K3Wnt69aOwWo=",
			Status:    "1",
		}},
		Sign: "zUJPOWJjCnMFISqbsbubsonJ6gBxqEWWuU0trjwUbInF65JRBdq+AX2n+thNGb6o0L3v2R/F+8wZi1U472vpDw==",
	}
	err := internal.ValidateMsgCert(msgcert)
	if err == nil {
		internal.StoreMsgCert(msgcert)
	}
	res := internal.GetMsgCert(1751113043)
	for _, cert := range res {
		fmt.Println(cert)
	}
}
