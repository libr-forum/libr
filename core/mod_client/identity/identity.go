package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	cryptoconfig "github.com/libr-forum/Libr/core/crypto/config"
	"github.com/libr-forum/Libr/core/mod_client/keycache"
)

const MainIdentityID = "main"

func storeRoot() string {
	return filepath.Join(filepath.Dir(cryptoconfig.PrivateKeyPath), "identities")
}

func activeIdentityPath() string {
	return filepath.Join(storeRoot(), "active_identity")
}

func identityDir(id string) string {
	return filepath.Join(storeRoot(), id)
}

func identityKeyPaths(id string) (string, string) {
	dir := identityDir(id)
	return filepath.Join(dir, "priv.key"), filepath.Join(dir, "pub.key")
}

func EnsureInitialized() error {
	if err := os.MkdirAll(storeRoot(), 0700); err != nil {
		return err
	}

	if !IdentityExists(MainIdentityID) {
		if err := SaveIdentityFromActive(MainIdentityID); err != nil {
			return err
		}
	}

	activeID, err := GetActiveIdentityID()
	if err != nil || activeID == "" {
		activeID = MainIdentityID
		if err := setActiveIdentityID(activeID); err != nil {
			return err
		}
	}

	if !IdentityExists(activeID) {
		activeID = MainIdentityID
		if err := setActiveIdentityID(activeID); err != nil {
			return err
		}
	}

	_, err = ActivateIdentity(activeID)
	return err
}

func GetActiveIdentityID() (string, error) {
	data, err := os.ReadFile(activeIdentityPath())
	if err != nil {
		if os.IsNotExist(err) {
			return MainIdentityID, nil
		}
		return "", err
	}
	id := strings.TrimSpace(string(data))
	if id == "" {
		return MainIdentityID, nil
	}
	return id, nil
}

func setActiveIdentityID(id string) error {
	if err := os.MkdirAll(storeRoot(), 0700); err != nil {
		return err
	}
	return os.WriteFile(activeIdentityPath(), []byte(id), 0600)
}

func IdentityExists(id string) bool {
	privPath, pubPath := identityKeyPaths(id)
	if _, err := os.Stat(privPath); err != nil {
		return false
	}
	if _, err := os.Stat(pubPath); err != nil {
		return false
	}
	return true
}

func SaveIdentityFromActive(id string) error {
	privBytes, err := os.ReadFile(cryptoconfig.PrivateKeyPath)
	if err != nil {
		return err
	}
	pubBytes, err := os.ReadFile(cryptoconfig.PublicKeyPath)
	if err != nil {
		return err
	}
	return writeIdentity(id, privBytes, pubBytes)
}

func SyncActiveIdentity() error {
	activeID, err := GetActiveIdentityID()
	if err != nil {
		return err
	}
	return SaveIdentityFromActive(activeID)
}

func CreateIncognitoIdentity() (string, string, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}

	id := fmt.Sprintf("incognito-%d", time.Now().UnixNano())
	if err := writeIdentity(id, priv, pub); err != nil {
		return "", "", err
	}
	return id, base64.StdEncoding.EncodeToString(pub), nil
}

func ActivateIdentity(id string) (string, error) {
	privBytes, pubBytes, err := readIdentityBytes(id)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(cryptoconfig.PrivateKeyPath), 0700); err != nil {
		return "", err
	}
	if err := os.WriteFile(cryptoconfig.PrivateKeyPath, privBytes, 0600); err != nil {
		return "", err
	}
	if err := os.WriteFile(cryptoconfig.PublicKeyPath, pubBytes, 0644); err != nil {
		return "", err
	}
	if err := setActiveIdentityID(id); err != nil {
		return "", err
	}

	keycache.InitKeys()
	return base64.StdEncoding.EncodeToString(pubBytes), nil
}

func LoadPrivateKey(id string) (ed25519.PrivateKey, error) {
	privBytes, _, err := readIdentityBytes(id)
	if err != nil {
		return nil, err
	}
	if len(privBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size for identity %s", id)
	}
	return ed25519.PrivateKey(privBytes), nil
}

func DeleteIdentity(id string) error {
	if id == "" || id == MainIdentityID {
		return nil
	}
	return os.RemoveAll(identityDir(id))
}

func writeIdentity(id string, privBytes []byte, pubBytes []byte) error {
	privPath, pubPath := identityKeyPaths(id)
	if err := os.MkdirAll(filepath.Dir(privPath), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(privPath, privBytes, 0600); err != nil {
		return err
	}
	if err := os.WriteFile(pubPath, pubBytes, 0644); err != nil {
		return err
	}
	return nil
}

func readIdentityBytes(id string) ([]byte, []byte, error) {
	privPath, pubPath := identityKeyPaths(id)
	privBytes, err := os.ReadFile(privPath)
	if err != nil {
		return nil, nil, err
	}
	pubBytes, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, nil, err
	}
	if len(pubBytes) != ed25519.PublicKeySize {
		return nil, nil, fmt.Errorf("invalid public key size for identity %s", id)
	}
	return privBytes, pubBytes, nil
}
