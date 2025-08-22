package cache

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

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

func GetCacheDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		return filepath.Join(appData, "libr", "cache")
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "./keys"
		}
		return filepath.Join(home, "Library", "Application Support", "libr", "cache")
	case "linux":
		home, err := os.UserHomeDir()
		if err != nil {
			return "./keys"
		}
		return filepath.Join(home, ".config", "libr", "cache")
	default:
		return "./keys"
	}
}

func getCachePath() string {
	if cachePath == "" {
		cachePath = filepath.Join(GetCacheDir(), cacheFileName)
	}
	return cachePath
}

// InitCacheFile creates the JSON cache file if it doesn't exist.
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

// GetFromCache looks up a public key in the cache.
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

// AddToCache adds a new publicKey/avatar/alias entry to the JSON file.
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

// Func for pending moderation
func SavePendingModeration(pending types.PendingModeration) error {
	dir := filepath.Join(GetCacheDir(), "pending_mods")
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.LogToFile("[DEBUG]Failed to create pending_mods dir")
		return fmt.Errorf("failed to create pending_mods dir: %w", err)
	}

	filePath := filepath.Join(dir, sanitizeFileName(pending.MsgSign)+".json")
	data, err := json.MarshalIndent(pending, "", "  ")
	if err != nil {
		logger.LogToFile("[DEBUG]Failed to marshal pending moderation")
		return fmt.Errorf("failed to marshal pending moderation: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write pending file: %w", err)
	}

	return nil
}

func LoadPendingModeration(path string) (types.PendingModeration, error) {
	filePath := filepath.Join(path)
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

func DeletePendingModeration(msgSign string) error {
	dir := filepath.Join(GetCacheDir(), "pending_mods")
	filePath := filepath.Join(dir, sanitizeFileName(msgSign)+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete pending moderation file: %w", err)
	}
	return nil
}

func sanitizeFileName(msgSign string) string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(msgSign))
}
