package service

import (
	"crypto/ed25519"
	"fmt"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
	"github.com/devlup-labs/Libr/core/mod/models"
)

func ModSign(req models.Msg, status string, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) (models.ModResponse, error) {

	payload := fmt.Sprintf("%s|%d|%s", req.Content, req.Ts, status)

	public_key, sign, err := cryptoutils.SignMessage(privateKey, payload)
	if err != nil {
		return models.ModResponse{}, err
	}

	return models.ModResponse{
		Sign:      sign,
		Status:    status,
		PublicKey: public_key,
	}, nil
}
