package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/libr-forum/Libr/core/libr/alias"
	libr "github.com/libr-forum/Libr/core/libr/app"
	"github.com/libr-forum/Libr/core/libr/avatar"
	cache "github.com/libr-forum/Libr/core/libr/cache"
	"github.com/libr-forum/Libr/core/libr/crypto/cryptoutils"
	"github.com/libr-forum/Libr/core/libr/identity"
	"github.com/libr-forum/Libr/core/libr/keycache"
	"github.com/libr-forum/Libr/core/libr/logger"
	service "github.com/libr-forum/Libr/core/libr/mod/service"
	"github.com/libr-forum/Libr/core/libr/models"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails adapter over the shared v2 node (core/libr/app).
type App struct {
	ctx context.Context

	mu   sync.Mutex
	node *libr.App
}

func NewApp() *App {
	cache.InitCacheFile()
	keycache.InitKeys()
	_ = identity.EnsureInitialized()
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logger.LogToFile("App started")
	runtime.WindowMaximise(ctx)
	go func() {
		runtime.EventsEmit(ctx, "navigate-to-root")
	}()
}

// shutdown is called by Wails before the window closes.
func (a *App) shutdown(_ context.Context) bool {
	a.stop()
	return false // do not prevent the window from closing
}

func (a *App) emit(name string, data ...interface{}) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, name, data...)
	}
}

// Connect joins the LIBR network as the active identity. Relays, bootstrap peers and
// seed servers come from LIBR_RELAYS, LIBR_BOOTSTRAP and LIBR_SEED_URLS (plus the
// peer cache and the seeds shipped in the binary).
func (a *App) Connect() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.node != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	n, err := libr.Start(ctx, libr.Config{
		DataDir: dataDir(activeIdentity()),
		Key:     keycache.PrivKey,
		Policy:  postPolicy,
		OnFinalized: func(p libr.PendingInfo) {
			a.emit("moderation_finalized", p)
		},
	})
	if err != nil {
		a.emit("relay_status_changed", "offline")
		return err
	}
	a.node = n
	a.emit("relay_status_changed", "online")
	return nil
}

// dataDir is per identity, so pending reports of an incognito identity are never
// resumed (signed) by the main one.
func dataDir(identityID string) string {
	return filepath.Join(cache.GetCacheDir(), "v2", identityID)
}

func activeIdentity() string {
	id, err := identity.GetActiveIdentityID()
	if err != nil {
		return identity.MainIdentityID
	}
	return id
}

func hasPending(identityID string) bool {
	files, _ := filepath.Glob(filepath.Join(dataDir(identityID), "pending", "*.json"))
	return len(files) > 0
}

func (a *App) stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.node != nil {
		a.node.Close()
		a.node = nil
	}
}

// restart reconnects after the identity changed (moderator duties follow the key).
func (a *App) restart() {
	a.stop()
	go func() {
		if err := a.Connect(); err != nil {
			log.Printf("reconnect: %v", err)
		}
	}()
}

var errOffline = errors.New("not connected to the LIBR network")

func (a *App) n() (*libr.App, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.node == nil {
		return nil, errOffline
	}
	return a.node, nil
}

func opCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 90*time.Second)
}

// GetRelayStatus is "online" or "offline" for the status badge.
func (a *App) GetRelayStatus() string {
	n, err := a.n()
	if err != nil || !n.Status().Online {
		return "offline"
	}
	return "online"
}

// GetNetworkStatus is the detailed network state.
func (a *App) GetNetworkStatus() libr.Status {
	n, err := a.n()
	if err != nil {
		return libr.Status{}
	}
	return n.Status()
}

func (a *App) TitleBarTheme(isDark bool) {
	if isDark {
		runtime.WindowSetDarkTheme(a.ctx)
	} else {
		runtime.WindowSetLightTheme(a.ctx)
	}
}

func (a *App) FetchPubKey() string {
	return keycache.LoadPubKey()
}

// ListCommunities lists the communities this build knows, with the user's role.
func (a *App) ListCommunities() ([]libr.Community, error) {
	n, err := a.n()
	if err != nil {
		return nil, err
	}
	return n.Communities(), nil
}

// IsMod reports whether the active identity moderates community.
func (a *App) IsMod(community string) bool {
	n, err := a.n()
	return err == nil && n.IsMod(community)
}

// SendInput posts text or an image (an <img src="data:..."> tag, at most 1 MiB in
// total) to community.
func (a *App) SendInput(community, input string) (libr.Outcome, error) {
	n, err := a.n()
	if err != nil {
		return libr.Outcome{}, err
	}
	ctx, cancel := opCtx()
	defer cancel()
	return n.Post(ctx, community, input)
}

// FetchAll returns the last 24 hours of community's verified posts, newest first.
func (a *App) FetchAll(community string) ([]libr.Post, error) {
	n, err := a.n()
	if err != nil {
		return nil, err
	}
	ctx, cancel := opCtx()
	defer cancel()
	return n.Feed(ctx, community, 24*time.Hour)
}

// Report asks the community's moderators to remove the post with hash.
func (a *App) Report(hash, reason string) (libr.Outcome, error) {
	n, err := a.n()
	if err != nil {
		return libr.Outcome{}, err
	}
	ctx, cancel := opCtx()
	defer cancel()
	return n.Report(ctx, hash, reason)
}

// Delete withdraws one of the user's own posts.
func (a *App) Delete(hash string) error {
	n, err := a.n()
	if err != nil {
		return err
	}
	ctx, cancel := opCtx()
	defer cancel()
	return n.Delete(ctx, hash)
}

// GetPending lists the user's posts and reports waiting for moderators.
func (a *App) GetPending() []libr.PendingInfo {
	n, err := a.n()
	if err != nil {
		return nil
	}
	return n.Pending()
}

// QueueItem is one entry of the moderator's human review queue.
type QueueItem struct {
	Key       string `json:"key"`
	Kind      string `json:"kind"` // post | report
	Community string `json:"community"`
	Content   string `json:"content"`
	Reason    string `json:"reason"`
	Received  int64  `json:"received"`
	Decided   bool   `json:"decided"`
	Approved  bool   `json:"approved"`
}

// GetModQueue lists posts and reports waiting for this moderator's decision.
func (a *App) GetModQueue() []QueueItem {
	n, err := a.n()
	if err != nil {
		return nil
	}
	var out []QueueItem
	for _, it := range n.ModQueue() {
		q := QueueItem{Key: it.Key, Kind: it.Kind, Community: it.Cert.Msg.Community, Content: string(it.Cert.Msg.Content),
			Reason: it.Reason, Received: it.Received, Decided: it.Vote != nil}
		if it.Vote != nil {
			q.Approved = it.Vote.Approve
		}
		out = append(out, q)
	}
	return out
}

// ModDecide records this moderator's decision. For a post, approve publishes; for a
// report, approve removes the post.
func (a *App) ModDecide(key string, approve bool) error {
	n, err := a.n()
	if err != nil {
		return err
	}
	return n.ModDecide(key, approve)
}

func (a *App) GenerateAvatar(key string) string {
	// Check cache
	record, err := cache.GetFromCache(key)
	if err == nil && record != nil && record.AvatarSVG != "" {
		return base64.StdEncoding.EncodeToString([]byte(record.AvatarSVG))
	}

	// Not cached, generate
	svg := avatar.GenerateAvatar(key)
	encodedSVG := base64.StdEncoding.EncodeToString([]byte(svg))

	// Get alias if available, else empty
	alias := ""
	if record != nil {
		alias = record.Alias
	}

	// Write to cache
	_ = cache.AddToCache(key, svg, alias)

	return encodedSVG
}

func (a *App) GenerateAlias(key string) string {
	// Check cache
	record, err := cache.GetFromCache(key)
	if err == nil && record != nil && record.Alias != "" {
		return record.Alias
	}

	// Not cached, generate
	genAlias := alias.GenerateAlias(key)

	// Get SVG if available, else empty
	svg := ""
	if record != nil {
		svg = record.AvatarSVG
	}

	// Write to cache
	_ = cache.AddToCache(key, svg, genAlias)

	return genAlias
}

func (a *App) RegenKeys() string {
	pub, _, _ := cryptoutils.GenerateKeyPair()
	keycache.InitKeys()
	_ = identity.SyncActiveIdentity()
	a.restart()
	return base64.StdEncoding.EncodeToString(pub)
}

func (a *App) IsIncognitoEnabled() bool {
	activeID, err := identity.GetActiveIdentityID()
	return err == nil && activeID != identity.MainIdentityID
}

func (a *App) EnableIncognito() string {
	if a.IsIncognitoEnabled() {
		return base64.StdEncoding.EncodeToString(keycache.PubKey)
	}
	incognitoID, _, err := identity.CreateIncognitoIdentity()
	if err != nil {
		return fmt.Sprintf("error:create incognito identity: %v", err)
	}
	pubKey, err := identity.ActivateIdentity(incognitoID)
	if err != nil {
		_ = identity.DeleteIdentity(incognitoID)
		return fmt.Sprintf("error:activate incognito identity: %v", err)
	}
	a.restart()
	return pubKey
}

func (a *App) DisableIncognito() string {
	if !a.IsIncognitoEnabled() {
		keycache.InitKeys()
		return base64.StdEncoding.EncodeToString(keycache.PubKey)
	}
	activeID, err := identity.GetActiveIdentityID()
	if err != nil {
		return fmt.Sprintf("error:read active identity: %v", err)
	}
	pubKey, err := identity.ActivateIdentity(identity.MainIdentityID)
	if err != nil {
		return fmt.Sprintf("error:restore main identity: %v", err)
	}
	a.restart()
	if activeID != identity.MainIdentityID && !hasPending(activeID) {
		_ = identity.DeleteIdentity(activeID)
		_ = os.RemoveAll(dataDir(activeID))
	}
	return pubKey
}

func (a *App) GetModerationLogs() ([]models.ModLogEntry, error) {
	cacheDir := cache.GetCacheDir()
	filePath := filepath.Join(cacheDir, "modlog.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var logs []models.ModLogEntry
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, err
	}

	// Sort by TimeStamp (latest first) using string comparison
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].TimeStamp > logs[j].TimeStamp
	})

	return logs, nil
}

func (a *App) GetModConfig() (models.ModConfig, error) {
	config, err := service.ReadModConfigFile()
	if err != nil {
		return models.ModConfig{}, err
	}
	return config, nil
}

// SaveModConfig writes to centralized config file path
func (a *App) SaveModConfig(cfg models.ModConfig) error {
	path := service.GetModConfigPath()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create modconfig directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}

func (a *App) SaveGoogleApiKey(key string) error {
	return service.WriteModKey("GOOGLE_NLP_API_KEY", key)
}

func (a *App) GetModKeys() (map[string]string, error) {
	keys, err := service.ReadModKeys()
	if err != nil {
		return nil, err
	}
	masked := make(map[string]string)
	for k, v := range keys {
		if len(v) > 8 {
			masked[k] = v[:4] + "..." + v[len(v)-4:]
		} else if len(v) > 0 {
			masked[k] = "********"
		} else {
			masked[k] = ""
		}
	}
	return masked, nil
}

func (a *App) SaveModKey(keyName, keyValue string) error {
	return service.WriteModKey(keyName, keyValue)
}

func (a *App) TestLLMConnection(provider, model, apiKey string) (string, error) {
	return a.TestDirectCloudConnection(provider, model, apiKey)
}

func (a *App) LogToFile(msg string) {
	logger.LogToFile(msg)
}

func (a *App) TestDirectCloudConnection(provider, model, apiKey string) (string, error) {
	if apiKey == "" {
		keys, _ := service.ReadModKeys()
		if keys != nil {
			providerKeyName := strings.ToUpper(provider) + "_API_KEY"
			if k, ok := keys[providerKeyName]; ok && k != "" {
				apiKey = k
			} else if k, ok := keys["LLM_API_KEY"]; ok && k != "" {
				apiKey = k
			} else if k, ok := keys["GOOGLE_NLP_API_KEY"]; ok && (provider == "gemini" || provider == "google") {
				apiKey = k
			}
		}
	}

	url, reqModel := service.ResolveCloudEndpoint(provider, model)
	if url == "" {
		return "", fmt.Errorf("direct cloud endpoint for provider '%s' is not mapped (supported: groq, gemini, openai, deepseek)", provider)
	}
	if apiKey == "" {
		return "", fmt.Errorf("missing API key for %s", provider)
	}

	payload := map[string]interface{}{
		"model": reqModel,
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
		"max_tokens": 5,
	}
	jsonBody, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("direct cloud connection error: %w", err)
	}
	defer resp.Body.Close()

	elapsed := time.Since(start).Milliseconds()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("cloud provider returned %d: %s", resp.StatusCode, string(b))
	}

	return fmt.Sprintf("Direct connection verified! %s responded in %dms", reqModel, elapsed), nil
}
