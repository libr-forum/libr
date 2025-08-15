package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// PublicKeyPath and PrivateKeyPath store the file paths for the
// Ed25519 public and private keys, dynamically determined based on the OS.
var (
	PublicKeyPath  string
	PrivateKeyPath string
)

// init initializes the key file paths depending on the operating system.
// It sets sensible default paths based on platform conventions.
func init() {
	switch runtime.GOOS {
	case "windows":
		// %APPDATA%\libr\keys\
		appData := os.Getenv("APPDATA")
		PrivateKeyPath = filepath.Join(appData, "libr", "keys", "priv.key")
		PublicKeyPath = filepath.Join(appData, "libr", "keys", "pub.key")

	case "darwin":
		// ~/Library/Application Support/libr/keys/
		home, err := os.UserHomeDir()
		if err != nil {
			panic("unable to get user home directory: " + err.Error())
		}
		PrivateKeyPath = filepath.Join(home, "Library", "Application Support", "libr", "keys", "priv.key")
		PublicKeyPath = filepath.Join(home, "Library", "Application Support", "libr", "keys", "pub.key")

	case "linux":
		// ~/.config/libr/keys/
		home, err := os.UserHomeDir()
		if err != nil {
			
			panic("unable to get user home directory: " + err.Error())
		}
		PrivateKeyPath = filepath.Join(home, ".config", "libr", "keys", "priv.key")
		PublicKeyPath = filepath.Join(home, ".config", "libr", "keys", "pub.key")

	default:
		// Fallback to relative path: ./keys/
		PrivateKeyPath = filepath.Join("keys", "priv.key")
		PublicKeyPath = filepath.Join("keys", "pub.key")
	}
}
