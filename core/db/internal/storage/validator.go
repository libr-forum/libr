package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/libr-forum/Libr/core/crypto/cryptoutils"
	"github.com/libr-forum/Libr/core/db/internal/models"
)

func ValidateMsgCert(msgcert *models.MsgCert) error {
	fmt.Println("Validating MsgCert")
	if err := ValidateMsgCertFields(msgcert); err != nil {
		fmt.Println("Error validating MsgCert fields:", err)
		return err
	}

	sort.SliceStable(msgcert.ModCerts, func(i, j int) bool {
		return msgcert.ModCerts[i].PublicKey < msgcert.ModCerts[j].PublicKey
	})

	dataToVerify := models.DataToSign{
		Content:  msgcert.Msg.Content,
		Ts:       msgcert.Msg.Ts,
		ModCerts: msgcert.ModCerts,
	}

	jsonBytes, err := json.Marshal(dataToVerify)
	fmt.Println("Data to verify:", string(jsonBytes))
	if err != nil {
		log.Printf("DB failed to marshal DataToSign: %v", err)
		return err
	}

	fmt.Print("Public Key:", msgcert.PublicKey)

	if !cryptoutils.VerifySignature(msgcert.PublicKey, string(jsonBytes), msgcert.Sign) {
		return fmt.Errorf("❌ Invalid MsgCert signature")
	}

	return nil
}

func ValidateMsgCertFields(msgcert *models.MsgCert) error {
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

func ValidateRepCertFields(repCert *models.ReportCert) error {
	if repCert.Msgcert.PublicKey == "" {
		return errors.New("sender public key is required")
	}
	if repCert.Msgcert.Msg.Content == "" {
		return errors.New("message content is required")
	}
	if repCert.Msgcert.Msg.Ts == 0 {
		return errors.New("timestamp is required")
	}
	return nil
}

func ValidateRepCert(repCert *models.ReportCert, validMods []*models.Mod) error {
	if err := ValidateRepCertFields(repCert); err != nil {
		return err
	}

	// Validate MsgCert Signature (signing the message data + modcerts)
	sort.SliceStable(repCert.Msgcert.ModCerts, func(i, j int) bool {
		return repCert.Msgcert.ModCerts[i].PublicKey < repCert.Msgcert.ModCerts[j].PublicKey
	})

	dataToVerify := models.DataToSign{
		Content:  repCert.Msgcert.Msg.Content,
		Ts:       repCert.Msgcert.Msg.Ts,
		ModCerts: repCert.Msgcert.ModCerts,
	}
	fmt.Println("Data to verify:", dataToVerify)
	jsonBytes, err := json.Marshal(dataToVerify)
	if err != nil {
		log.Printf("DB failed to marshal DataToSign: %v", err)
		return err
	}
	fmt.Println("Data to verify JSON:", string(jsonBytes))
	if !cryptoutils.VerifySignature(repCert.Msgcert.PublicKey, string(jsonBytes), repCert.Msgcert.Sign) {
		return fmt.Errorf("❌ Invalid MsgCert signature")
	}

	// Validate associated moderator certs
	return ValidateRepModCerts(repCert, validMods)
}

func ValidateRepModCerts(repCert *models.ReportCert, validMods []*models.Mod) error {
	if repCert.Mode == "delete" {
		// Author self-delete: needs the original author's signature over the cert sign; no mod quorum.
		for _, c := range repCert.RepModCerts {
			if c.PublicKey == repCert.Msgcert.PublicKey && cryptoutils.VerifySignature(c.PublicKey, repCert.Msgcert.Sign, c.Sign) {
				return nil
			}
		}
		return errors.New("delete mode: no valid signature from the original author")
	}
	return checkQuorum(repCert.RepModCerts, validMods, func(c models.ModCert) string {
		return repCert.Msgcert.Sign + c.Status
	})
}

// ValidateModCert checks the moderator approvals on a MsgCert against the given moderator set.
func ValidateModCert(msgCert *models.MsgCert, validMods []*models.Mod) error {
	return checkQuorum(msgCert.ModCerts, validMods, func(c models.ModCert) string {
		return msgCert.Msg.Content + strconv.FormatInt(msgCert.Msg.Ts, 10) + c.Status
	})
}

// checkQuorum counts each moderator key in validMods at most once, after verifying its
// signature over payload(cert). It fails closed: an empty moderator set, an unknown key,
// a repeated key or a bad signature is an error.
func checkQuorum(certs []models.ModCert, validMods []*models.Mod, payload func(models.ModCert) string) error {
	total := len(validMods)
	if total == 0 {
		return errors.New("moderator set unavailable or empty")
	}
	valid := make(map[string]bool, total)
	for _, m := range validMods {
		valid[m.PublicKey] = true
	}

	seen := make(map[string]bool, len(certs))
	appr, rej := 0, 0
	for _, c := range certs {
		if !valid[c.PublicKey] {
			return fmt.Errorf("unauthorized moderator: %s", c.PublicKey)
		}
		if seen[c.PublicKey] {
			return fmt.Errorf("duplicate vote from moderator: %s", c.PublicKey)
		}
		seen[c.PublicKey] = true
		if !cryptoutils.VerifySignature(c.PublicKey, payload(c), c.Sign) {
			return fmt.Errorf("invalid signature from moderator: %s", c.PublicKey)
		}
		switch c.Status {
		case "1":
			appr++
		case "0":
			rej++
		}
	}

	// v1 rule kept: rejections <= half the set, approvals > half of votes and > 30% of the set.
	if rej > total/2 {
		return fmt.Errorf("too many rejections: %d of %d", rej, total)
	}
	if 2*appr <= appr+rej || 10*appr <= 3*total {
		return fmt.Errorf("not enough approvals: %d of %d (rejections %d)", appr, total, rej)
	}
	return nil
}
