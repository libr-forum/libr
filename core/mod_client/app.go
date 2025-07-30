package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
	"github.com/devlup-labs/Libr/core/mod_client/alias"
	"github.com/devlup-labs/Libr/core/mod_client/avatar"
	cache "github.com/devlup-labs/Libr/core/mod_client/cache_handler"
	"github.com/devlup-labs/Libr/core/mod_client/config"
	"github.com/devlup-labs/Libr/core/mod_client/core"
	service "github.com/devlup-labs/Libr/core/mod_client/internal/service"
	"github.com/devlup-labs/Libr/core/mod_client/keycache"
	"github.com/devlup-labs/Libr/core/mod_client/models"
	Peers "github.com/devlup-labs/Libr/core/mod_client/peers"
	"github.com/devlup-labs/Libr/core/mod_client/types"
	util "github.com/devlup-labs/Libr/core/mod_client/util"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	relayStatus string
}

func NewApp() *App {
	cache.InitCacheFile()
	keycache.InitKeys()
	config.LoadConfig()
	return &App{relayStatus: "offline"}
}

func (a *App) FetchPubKey() string {
	pubStr := keycache.LoadPubKey()
	return pubStr
}

func (a *App) ModAuthentication(myKey string) bool {
	val, err := util.AmIMod(myKey)
	if err != nil {
		return false
	}
	return val
}

func (a *App) GetOnlineMods() []string {
	onlineMods, err := util.GetOnlineMods()
	if err != nil {
		return nil
	}

	var publicKeys []string
	for _, mod := range onlineMods {
		publicKeys = append(publicKeys, mod.PublicKey)
	}

	return publicKeys
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

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.WindowMaximise(ctx)
	go func() {
		runtime.EventsEmit(ctx, "navigate-to-root")
	}()
}

func (a *App) RegenKeys() string {
	pub, _, _ := cryptoutils.GenerateKeyPair()
	return base64.StdEncoding.EncodeToString(pub)
}

func (a *App) Connect(relayAdds []string) error {
	err := Peers.StartNode(relayAdds)
	if err != nil {
		a.relayStatus = "offline"
		return err
	}
	a.relayStatus = "online"
	return nil
}

func (a *App) GetRelayStatus() string {
	return a.relayStatus
}

func (a *App) TitleBarTheme(isDark bool) {
	if isDark {
		runtime.WindowSetDarkTheme(a.ctx)
	} else {
		runtime.WindowSetLightTheme(a.ctx)
	}
}

func (a *App) SendInput(input string) string {
	if a.relayStatus != "online" {
		return "Offline"
	}

	ts := time.Now().Unix()

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// Run SendToMods with timeout
	modChan := make(chan []types.ModCert, 1)

	go func() {
		modcerts := core.SendToMods(input, ts, "", "")
		modChan <- modcerts
	}()

	var modcertlist []types.ModCert
	select {
	case modcertlist = <-modChan:
	case <-ctx.Done():
		return ":x: Moderator timeout"
	}

	if len(modcertlist) == 0 {
		return ":x: Message rejected by moderators."
	}

	msgCert := core.CreateMsgCert(input, ts, modcertlist)
	tsmin := msgCert.Msg.Ts - (msgCert.Msg.Ts % 60)
	key := util.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	core.SendToDb(key, msgCert, "/route=store")

	return fmt.Sprintf(":white_check_mark: Sent to DB. Time: %d", msgCert.Msg.Ts)
}

func (a *App) Report(message string, ts int64, reason string, originalsender string) string {
	if a.relayStatus != "online" {
		return "Offline"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// Run SendToMods with timeout
	modChan := make(chan []types.ModCert, 1)

	go func() {
		modcerts := core.SendToMods(message, ts, reason, "manual")
		modChan <- modcerts
	}()

	var modcertlist []types.ModCert
	select {
	case modcertlist = <-modChan:
	case <-ctx.Done():
		return ":x: Moderator timeout"
	}

	if len(modcertlist) == 0 {
		return ":x: Message rejected by moderators."
	}

	repCert := core.CreateRepCert(originalsender, message, ts, modcertlist)
	tsmin := repCert.ReportMsg.Msg.Ts - (repCert.ReportMsg.Msg.Ts % 60)
	key := util.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	core.SendToDb(key, repCert, "/route=delete")

	return fmt.Sprintf(":white_check_mark: Sent to DB. Time: %d", repCert.ReportMsg.Msg.Ts)
}

func (a *App) FetchAll() []string {
	messages := core.FetchRecent(context.Background())
	var out []string
	for _, cert := range messages {
		out = append(out, fmt.Sprintf("Sender: %s | Msg: %s | Time: %d | ModCerts: %s", cert.PublicKey, cert.Msg.Content, cert.Msg.Ts, cert.ModCerts))
	}
	return out
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
	path := service.GetModEnvPath()

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read env file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	found := false

	for i, line := range lines {
		if strings.HasPrefix(line, "GOOGLE_NLP_API_KEY=") {
			lines[i] = "GOOGLE_NLP_API_KEY=" + key
			found = true
			break
		}
	}

	if !found {
		// If not found, append at the end
		lines = append(lines, "GOOGLE_NLP_API_KEY="+key)
	}

	newContent := strings.Join(lines, "\n")

	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	return nil
}
