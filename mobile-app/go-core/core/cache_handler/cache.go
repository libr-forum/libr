package cache

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/libr-forum/Libr/core/crypto/config"
	"github.com/libr-forum/Libr/core/mod_client/logger"
	"github.com/libr-forum/Libr/core/mod_client/types"
)

const cacheFileName = "alias_cache.json"

type CacheEntry struct {
	PublicKey string `json:"public_key"`
	AvatarSVG string `json:"avatar_svg"` // base64-encoded
	Alias     string `json:"alias"`
}

type CacheData map[string]*CacheEntry

var (
	cacheData CacheData
	cachePath string
	cacheMu   sync.Mutex
)

// GetCacheDir returns the directory used for caching.
// On mobile, we use the same directory configured for keys.
func GetCacheDir() string {
	return filepath.Dir(config.PrivateKeyPath)
}

func getCachePath() string {
	if cachePath == "" {
		cachePath = filepath.Join(GetCacheDir(), cacheFileName)
	}
	return cachePath
}

// InitCacheFile creates the JSON alias cache file if it doesn't exist.
func InitCacheFile() error {
	dir := GetCacheDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := getCachePath()
	if _, err := os.Stat(path); err == nil {
		// File exists, load it
		f, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		json.Unmarshal(f, &cacheData)
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	cacheData = make(CacheData)
	return saveCache()
}

func saveCache() error {
	f, err := os.Create(getCachePath())
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cacheData)
}

// GetFromCache looks up a public key in the alias cache.
func GetFromCache(pubKey string) (*CacheEntry, error) {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	if entry, ok := cacheData[pubKey]; ok {
		svgBytes, _ := base64.StdEncoding.DecodeString(entry.AvatarSVG)
		return &CacheEntry{
			PublicKey: entry.PublicKey,
			AvatarSVG: string(svgBytes),
			Alias:     entry.Alias,
		}, nil
	}
	return nil, nil
}

// AddToCache adds a new publicKey/avatar/alias entry to the alias cache.
func AddToCache(key string, svg string, alias string) error {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	encodedSVG := base64.StdEncoding.EncodeToString([]byte(svg))
	cacheData[key] = &CacheEntry{
		PublicKey: key,
		AvatarSVG: encodedSVG,
		Alias:     alias,
	}
	return saveCache()
}

// ── Pending Moderation Cache ──────────────────────────────────────────────────

// SavePendingModeration saves a PendingModeration object to disk.
func SavePendingModeration(pending types.PendingModeration) error {
	dir := filepath.Join(GetCacheDir(), "pending_mods")
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.LogToFile("[DEBUG] Failed to create pending_mods dir")
		return fmt.Errorf("failed to create pending_mods dir: %w", err)
	}

	filePath := filepath.Join(dir, sanitizeFileName(pending.MsgSign)+".json")
	data, err := json.MarshalIndent(pending, "", "  ")
	if err != nil {
		logger.LogToFile("[DEBUG] Failed to marshal pending moderation")
		return fmt.Errorf("failed to marshal pending moderation: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write pending file: %w", err)
	}

	return nil
}

// LoadPendingModeration loads a PendingModeration object from a specific file path.
func LoadPendingModeration(filePath string) (types.PendingModeration, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return types.PendingModeration{}, fmt.Errorf("failed to read pending file: %w", err)
	}

	var pending types.PendingModeration
	if err := json.Unmarshal(data, &pending); err != nil {
		return types.PendingModeration{}, fmt.Errorf("failed to unmarshal pending data: %w", err)
	}

	return pending, nil
}

// DeletePendingModeration deletes a pending moderation file by its message signature.
func DeletePendingModeration(msgSign string) error {
	dir := filepath.Join(GetCacheDir(), "pending_mods")
	filePath := filepath.Join(dir, sanitizeFileName(msgSign)+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete pending moderation file: %w", err)
	}
	return nil
}

// GetAllPendingModerations returns a list of all currently pending moderations.
func GetAllPendingModerations() ([]types.PendingModeration, error) {
	pattern := filepath.Join(GetCacheDir(), "pending_mods", "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending files: %w", err)
	}

	var results []types.PendingModeration
	for _, filePath := range files {
		if pending, err := LoadPendingModeration(filePath); err == nil {
			results = append(results, pending)
		}
	}
	return results, nil
}

func HasPendingModerationForIdentity(identityID string) bool {
	if identityID == "" {
		return false
	}

	pendings, err := GetAllPendingModerations()
	if err != nil {
		return false
	}

	for _, pending := range pendings {
		if pending.SignerIdentityID == identityID {
			return true
		}
	}
	return false
}

// ── Finalized Moderation Cache ───────────────────────────────────────────────

// SaveFinalizedModeration saves a FinalizedModeration record to disk.
func SaveFinalizedModeration(finalized types.FinalizedModeration) error {
	dir := filepath.Join(GetCacheDir(), "finalized_mods")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create finalized_mods dir: %w", err)
	}

	filePath := filepath.Join(dir, sanitizeFileName(finalized.MsgSign)+".json")
	data, err := json.MarshalIndent(finalized, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal finalized moderation: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write finalized file: %w", err)
	}

	return nil
}

// LoadFinalizedModeration loads a FinalizedModeration record from disk.
func LoadFinalizedModeration(filePath string) (types.FinalizedModeration, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return types.FinalizedModeration{}, fmt.Errorf("failed to read finalized file: %w", err)
	}

	var finalized types.FinalizedModeration
	if err := json.Unmarshal(data, &finalized); err != nil {
		return types.FinalizedModeration{}, fmt.Errorf("failed to unmarshal finalized data: %w", err)
	}

	return finalized, nil
}

// DeleteFinalizedModeration deletes a finalized moderation file.
func DeleteFinalizedModeration(msgSign string) error {
	dir := filepath.Join(GetCacheDir(), "finalized_mods")
	filePath := filepath.Join(dir, sanitizeFileName(msgSign)+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete finalized moderation file: %w", err)
	}
	return nil
}

// GetAllFinalizedModerations returns all finalized moderations from disk.
func GetAllFinalizedModerations() ([]types.FinalizedModeration, error) {
	pattern := filepath.Join(GetCacheDir(), "finalized_mods", "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list finalized files: %w", err)
	}

	var results []types.FinalizedModeration
	for _, filePath := range files {
		if finalized, err := LoadFinalizedModeration(filePath); err == nil {
			results = append(results, finalized)
		}
	}
	return results, nil
}

func sanitizeFileName(msgSign string) string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(msgSign))
}

// ── Moderator List Cache ──────────────────────────────────────────────────────

const modsCacheFileName = "mods_cache.json"

func getModsCachePath() string {
	return filepath.Join(GetCacheDir(), modsCacheFileName)
}

// SaveModsCache persists the list of online moderators to disk.
func SaveModsCache(mods []types.Mod) error {
	dir := GetCacheDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(mods, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getModsCachePath(), data, 0644)
}

// GetModsCache retrieves the last known list of online moderators from disk.
func GetModsCache() ([]types.Mod, error) {
	data, err := os.ReadFile(getModsCachePath())
	if err != nil {
		return nil, err
	}

	var mods []types.Mod
	if err := json.Unmarshal(data, &mods); err != nil {
		return nil, err
	}

	return mods, nil
}
