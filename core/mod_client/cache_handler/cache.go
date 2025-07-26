package cache

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
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
