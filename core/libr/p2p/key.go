package p2p

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// LoadOrCreateKey reads a 32-byte Ed25519 seed from path, creating it (mode 0600) if
// absent. A node's peer ID and its signatures derive from this key, so it must survive
// restarts.
func LoadOrCreateKey(path string) (ed25519.PrivateKey, error) {
	seed, err := os.ReadFile(path)
	switch {
	case err == nil:
		if len(seed) != ed25519.SeedSize {
			return nil, fmt.Errorf("%s: want %d-byte seed, got %d", path, ed25519.SeedSize, len(seed))
		}
		return ed25519.NewKeyFromSeed(seed), nil
	case errors.Is(err, fs.ErrNotExist):
		seed = make([]byte, ed25519.SeedSize)
		if _, err := rand.Read(seed); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, seed, 0o600); err != nil {
			return nil, err
		}
		return ed25519.NewKeyFromSeed(seed), nil
	default:
		return nil, err
	}
}
