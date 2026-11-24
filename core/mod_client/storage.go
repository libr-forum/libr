package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"

	"github.com/libr-forum/Libr/core/db/node"
	cache "github.com/libr-forum/Libr/core/libr/cache"
	"github.com/libr-forum/Libr/core/libr/seeds"
)

// "Contribute storage": the desktop also runs a DB node for the network it joined. The
// node has its own key (not the user's identity, so storage is not linkable to posts)
// and its own data dir, shared by all identities on this machine.

func storageDir() string { return filepath.Join(cache.GetCacheDir(), "v2", "dbnode") }

func storageFlag() string { return filepath.Join(storageDir(), "enabled") }

// storageChoice holds the communities this machine stores; absent means all of them.
func storageChoice() string { return filepath.Join(storageDir(), "communities.json") }

// StorageCommunity is one community the DB node could store.
type StorageCommunity struct {
	Name   string `json:"name"`
	Stored bool   `json:"stored"`
}

// StorageStatus reports whether this machine contributes storage, for which
// communities, and the DB node's peer ID while it runs.
type StorageStatus struct {
	Enabled     bool               `json:"enabled"`
	Running     bool               `json:"running"`
	PeerID      string             `json:"peer_id"`
	Communities []StorageCommunity `json:"communities"`
}

func (a *App) GetStorageStatus() StorageStatus {
	_, err := os.Stat(storageFlag())
	a.mu.Lock()
	defer a.mu.Unlock()
	s := StorageStatus{Enabled: err == nil, Running: a.db != nil}
	if a.db != nil {
		s.PeerID = a.db.Host.ID().String()
	}
	if a.node != nil {
		_, _, _, sets := a.node.Network()
		chosen, all := loadStorageChoice()
		for _, c := range sets.Communities() {
			s.Communities = append(s.Communities, StorageCommunity{Name: c, Stored: all || slices.Contains(chosen, c)})
		}
	}
	return s
}

func loadStorageChoice() (chosen []string, all bool) {
	raw, err := os.ReadFile(storageChoice())
	if err != nil || json.Unmarshal(raw, &chosen) != nil {
		return nil, true
	}
	return chosen, false
}

// SetStorageCommunities picks the communities this machine stores (at least one) and
// restarts the DB node if it runs, so the choice applies at once.
func (a *App) SetStorageCommunities(names []string) error {
	if len(names) == 0 {
		return errors.New("choose at least one community, or turn storage off")
	}
	if err := os.MkdirAll(storageDir(), 0o700); err != nil {
		return err
	}
	raw, _ := json.Marshal(names)
	if err := os.WriteFile(storageChoice(), raw, 0o600); err != nil {
		return err
	}
	a.stopStorage()
	return a.startStorage()
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
	var optOut []string
	if chosen, all := loadStorageChoice(); !all {
		for _, c := range sets.Communities() {
			if !slices.Contains(chosen, c) {
				optOut = append(optOut, c)
			}
		}
	}
	cfg := node.Config{
		OptOut:    optOut,
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
