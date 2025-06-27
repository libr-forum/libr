package service

import (
	"crypto/ed25519"
	"encoding/json"
	"strconv"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
	"github.com/devlup-labs/Libr/core/mod/models"
)

func ModSign(req models.UserMsg, status string, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) (string, error) {

	payload := req.Content + strconv.FormatInt(req.TimeStamp, 10) + status
	public_key, sign, err := cryptoutils.SignMessage(privateKey, payload)
	if err != nil {
		return "", err
	}

	ModResp := models.ModResponse{
		Sign:      sign,
		Status:    status,
		PublicKey: public_key,
	}
	ModResponseString, _ := CanonicalizeModResp(ModResp)
	return ModResponseString, nil
}

func CanonicalizeModResp(ModResp models.ModResponse) (string, error) {
	canonical, err := json.Marshal(struct {
		Sign      string `json:"sign"`
		PublicKey string `json:"public_key"`
		Status    string `json:"status"`
	}{
		Sign:      ModResp.Sign,
		PublicKey: ModResp.PublicKey,
		Status:    ModResp.Status,
	})
	if err != nil {
		return "", err
	}

	return string(canonical), nil
}
