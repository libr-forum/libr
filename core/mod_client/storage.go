package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/libr-forum/Libr/core/db/node"
	cache "github.com/libr-forum/Libr/core/libr/cache"
	"github.com/libr-forum/Libr/core/libr/seeds"
)

// "Contribute storage": the desktop also runs a DB node for the network it joined. The
// node has its own key (not the user's identity, so storage is not linkable to posts)
// and its own data dir, shared by all identities on this machine.

func storageDir() string { return filepath.Join(cache.GetCacheDir(), "v2", "dbnode") }

func storageFlag() string { return filepath.Join(storageDir(), "enabled") }

// StorageStatus reports whether this machine contributes storage, and the DB node's
// peer ID while it runs.
type StorageStatus struct {
	Enabled bool   `json:"enabled"`
	Running bool   `json:"running"`
	PeerID  string `json:"peer_id"`
}

func (a *App) GetStorageStatus() StorageStatus {
	_, err := os.Stat(storageFlag())
	a.mu.Lock()
	defer a.mu.Unlock()
	s := StorageStatus{Enabled: err == nil, Running: a.db != nil}
	if a.db != nil {
		s.PeerID = a.db.Host.ID().String()
	}
	return s
}

// SetStorageContribution turns contributing storage on or off (remembered across
// restarts). Turning it on while connected starts the DB node at once.
func (a *App) SetStorageContribution(on bool) error {
	if !on {
		a.stopStorage()
		return os.Remove(storageFlag())
	}
	if err := os.MkdirAll(storageDir(), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(storageFlag(), nil, 0o600); err != nil {
		return err
	}
	return a.startStorage()
}

// startStorage runs the DB node if enabled and connected; no-op otherwise.
func (a *App) startStorage() error {
	if _, err := os.Stat(storageFlag()); err != nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db != nil {
		return nil
	}
	if a.node == nil {
		return errOffline
	}
	relays, boot, seedURLs, sets := a.node.Network()
	cfg := node.Config{
		DataDir:   storageDir(),
		Sets:      sets,
		Listen:    []string{"/ip4/0.0.0.0/tcp/0", "/ip4/0.0.0.0/udp/0/quic-v1"},
		Relays:    seeds.Parse(relays),
		Bootstrap: seeds.Parse(boot),
		SeedURLs:  seedURLs,
		Private:   len(relays) > 0, // a desktop is behind NAT: reachable through relays
	}
	if len(cfg.Relays) == 0 && len(cfg.Bootstrap) == 0 && len(cfg.SeedURLs) == 0 && len(seeds.Shipped()) == 0 {
		return errors.New("no network configured to contribute storage to")
	}
	db, err := node.Start(cfg)
	if err != nil {
		return err
	}
	a.db = db
	return nil
}

func (a *App) stopStorage() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db != nil {
		a.db.Close()
		a.db = nil
	}
}
