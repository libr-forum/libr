package msgcert

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
	"github.com/devlup-labs/Libr/core/db/models"
)

func ValidateMsgCert(msgcert models.MsgCert) error {

	if err := hasRequiredFields(msgcert); err != nil {
		return err
	}

	sort.SliceStable(msgcert.ModCerts, func(i, j int) bool {
		return msgcert.ModCerts[i].PublicKey < msgcert.ModCerts[j].PublicKey
	})

	dataToVerify := models.DataToSign{
		Content:   msgcert.Msg.Content,
		Timestamp: msgcert.Msg.Ts,
		ModCerts:  msgcert.ModCerts,
	}

	jsonBytes, err := json.Marshal(dataToVerify)
	if err != nil {
		log.Printf("DB failed to marshal DataToSign: %v", err)
		return err
	}

	if !cryptoutils.VerifySignature(msgcert.PublicKey, string(jsonBytes), msgcert.Sign) {
		return fmt.Errorf("❌ Invalid MsgCert signature")
	}

	return nil
}

func hasRequiredFields(msgcert models.MsgCert) error {
	if msgcert.PublicKey == "" {
		return errors.New("sender public key is required")
	}
	if msgcert.Msg.Content == "" {
		return errors.New("message content is required")
	}
	if msgcert.Msg.Ts == 0 {
		return errors.New("timestamp is required")
	}
	if len(msgcert.ModCerts) == 0 {
		return errors.New("at least one mod_cert is required")
	}
	return nil
}

// func IsAllocatedToThisNode(msgcert models.MsgCert,nodeId string) bool {

// }
