package cryptoutils

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/devlup-labs/Libr/core/crypto/config"
)

// GenerateKeyPair generates a new Ed25519 key pair,
// saves them to disk, and returns the public and private keys.
func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// Ensure the directory for the private key exists
	if err := os.MkdirAll(filepath.Dir(config.PrivateKeyPath), 0700); err != nil {
		return nil, nil, err
	}

	// Write private key to file with secure permissions
	if err := os.WriteFile(config.PrivateKeyPath, priv, 0600); err != nil {
		return nil, nil, err
	}

	// Write public key to file
	if err := os.WriteFile(config.PublicKeyPath, pub, 0644); err != nil {
		return nil, nil, err
	}

	return pub, priv, nil
}

// LoadKeys attempts to load the private and public keys from disk.
// If the private key is missing or invalid, a new key pair is generated.
// If the public key is missing or invalid, it is regenerated from the private key and saved.
func LoadKeys() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	privData, err := os.ReadFile(config.PrivateKeyPath)
	if err != nil {
		log.Println("Private key not found, generating new key pair.")
		return GenerateKeyPair()
	}

	if len(privData) != ed25519.PrivateKeySize {
		return nil, nil, errors.New("invalid private key size")
	}

	privKey := ed25519.PrivateKey(privData)
	// Derive the public key from the private key
	pubKey := privKey.Public().(ed25519.PublicKey)

	// Try to read the existing public key
	pubData, err := os.ReadFile(config.PublicKeyPath)
	if err != nil || len(pubData) != ed25519.PublicKeySize {
		log.Println("Public key missing or invalid, reconstructing from private key.")

		// Ensure the directory exists
		err := os.MkdirAll(filepath.Dir(config.PublicKeyPath), 0700)
		if err != nil {
			return nil, nil, err
		}

		// Save the derived public key to disk
		err = os.WriteFile(config.PublicKeyPath, pubKey, 0644)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// Use the valid public key from disk
		pubKey = ed25519.PublicKey(pubData)
	}

	return pubKey, privKey, nil
}

// SignMessage signs a string message using the provided private key,
// and returns a base64-encoded signature string.
func SignMessage(privateKey ed25519.PrivateKey, message string) (string, error) {
	sign := ed25519.Sign(privateKey, []byte(message))
	return base64.StdEncoding.EncodeToString(sign), nil
}

// VerifySignature checks whether the given base64-encoded signature
// matches the message using the public key.
func VerifySignature(publicKey ed25519.PublicKey, message string, sign string) bool {
	// Decode the base64-encoded signature
	decodedSign, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}
	// Verify the signature
	return ed25519.Verify(publicKey, []byte(message), decodedSign)
}
