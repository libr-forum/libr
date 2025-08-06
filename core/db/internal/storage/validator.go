package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/models"
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
	validMap := make(map[string]struct{})
	for _, mod := range validMods {
		validMap[mod.PublicKey] = struct{}{}
	}
	totalMods := len(validMods)

	msg := repCert.Msgcert.Msg
	msgCertSign := repCert.Msgcert.Sign
	msgCertPubKey := repCert.Msgcert.PublicKey

	approveCount := 0

	for _, repmodcert := range repCert.RepModCerts {
		if repCert.Mode == "delete" {
			// In delete mode, must be signed by the original author
			if repmodcert.PublicKey != msgCertPubKey {
				return fmt.Errorf("❌ Delete mode: modcert.PublicKey (%s) != MsgCert.PublicKey (%s)", repmodcert.PublicKey, msgCertPubKey)
			}
			if !cryptoutils.VerifySignature(repmodcert.PublicKey, msgCertSign, repmodcert.Sign) {
				return fmt.Errorf("❌ Delete mode: invalid signature from original user %s", repmodcert.PublicKey)
			}
			approveCount = totalMods
			break
		}

		// Manual (report) mode
		if _, ok := validMap[repmodcert.PublicKey]; !ok {
			return fmt.Errorf("❌ Unauthorized moderator: %s", repmodcert.PublicKey)
		}

		payload := msg.Content + strconv.FormatInt(msg.Ts, 10) + repmodcert.Status
		if !cryptoutils.VerifySignature(repmodcert.PublicKey, payload, repmodcert.Sign) {
			return fmt.Errorf("❌ Invalid signature from mod: %s", repmodcert.PublicKey)
		}

		if repmodcert.Status == "approve" {
			approveCount++
		}
	}

	// Check if majority of valid mods approved
	threshold := int(math.Ceil(float64(len(validMods)) * float64(config.RepMajority)))
	if approveCount < threshold {
		return fmt.Errorf("❌ Not enough approvals: got %d, need more than %d", approveCount, threshold)
	}

	return nil
}

func ValidateModCert(msgCert *models.MsgCert) error {

	for _, modcert := range msgCert.ModCerts {
		payload := msgCert.Msg.Content + strconv.FormatInt(msgCert.Msg.Ts, 10) + modcert.Status
		fmt.Println("Payload:", payload)
		if !cryptoutils.VerifySignature(modcert.PublicKey, payload, modcert.Sign) {
			return fmt.Errorf("❌ Invalid ModCert signature")
		}
	}

	return nil
}
