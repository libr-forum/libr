package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/libr-forum/Libr/core/crypto/cryptoutils"
	"github.com/libr-forum/Libr/core/mod_client/alias"
	"github.com/libr-forum/Libr/core/mod_client/avatar"
	cache "github.com/libr-forum/Libr/core/mod_client/cache_handler"
	"github.com/libr-forum/Libr/core/mod_client/config"
	"github.com/libr-forum/Libr/core/mod_client/core"
	moddb "github.com/libr-forum/Libr/core/mod_client/internal/mod_db"
	service "github.com/libr-forum/Libr/core/mod_client/internal/service"
	"github.com/libr-forum/Libr/core/mod_client/keycache"
	"github.com/libr-forum/Libr/core/mod_client/logger"
	"github.com/libr-forum/Libr/core/mod_client/models"
	Peers "github.com/libr-forum/Libr/core/mod_client/peers"
	"github.com/libr-forum/Libr/core/mod_client/types"
	util "github.com/libr-forum/Libr/core/mod_client/util"
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
	util.SetupMongo(config.MongoURI)
	amImod, _ := util.AmIMod(base64.StdEncoding.EncodeToString(keycache.PubKey))
	if amImod {
		config.InitDB()
	}
	core.MaybeStartCron()
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

func (a *App) GetRelayAddr() ([]string, error) {
	return util.GetRelayAddr()
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
	keycache.InitKeys()
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

func (a *App) SendInput(input string) (string, []types.ModCert) {
	if a.relayStatus != "online" {
		return "Offline", nil
	}

	ts := time.Now().Unix()

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// Run SendToMods with timeout
	modChan := make(chan []types.ModCert, 1)
	var err error
	go func() {
		var modcerts []types.ModCert
		modcerts, err = core.AutoSendToMods(input, ts)
		modChan <- modcerts
	}()

	if err != nil {
		return ":x: Moderator timeout", nil
	}

	var modcertlist []types.ModCert
	select {
	case modcertlist = <-modChan:
	case <-ctx.Done():
		return ":x: Moderator timeout", nil
	}

	if len(modcertlist) == 0 {
		return ":x: Message rejected by moderators.", nil
	}

	fmt.Println("ModCerts received:", modcertlist)

	msgCert := core.CreateMsgCert(input, ts, modcertlist)
	tsmin := msgCert.Msg.Ts - (msgCert.Msg.Ts % 60)
	key := util.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	core.SendToDb(key, msgCert, "/route=store")

	return fmt.Sprintf(":white_check_mark: Sent to DB. Time: %d Sign: %s", msgCert.Msg.Ts, msgCert.Sign), modcertlist
}

func (a *App) Report(msgcert types.MsgCert, reason *string) string {
	if a.relayStatus != "online" {
		return "Offline"
	}

	// ✅ Check if msgcert already exists in pending moderation files
	dir := filepath.Join(cache.GetCacheDir(), "pending_mods", "*.json")
	files, err := filepath.Glob(dir)
	if err != nil {
		logger.LogToFile("[DEBUG] Failed to list pending moderation files")
		log.Printf("Failed to list pending moderation files: %v", err)
	} else {
		for _, filePath := range files {
			pending, err := cache.LoadPendingModeration(filePath)
			if err != nil {
				logger.LogToFile(fmt.Sprintf("[DEBUG] Failed to load pending moderation file %s: %v", filePath, err))
				continue
			}
			if pending.MsgSign == msgcert.Sign {
				return ":white_check_mark: Already reported and pending moderation."
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	modChan := make(chan []types.ModCert, 1)
	mods, _ := util.GetOnlineMods()
	go func() {
		modcerts := core.ManualSendToMods(msgcert, mods, *reason, true)
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

	return fmt.Sprintf(":white_check_mark: Sent to Mods for approval. Time: %d", time.Now().Unix())
}

func (a *App) Delete(msgcert types.MsgCert) string {
	if a.relayStatus != "online" {
		return "Offline"
	}
	fmt.Println("Preparing to delete message with cert:", msgcert)
	payload := msgcert.Sign
	pubkey, sign, err := cryptoutils.SignMessage(keycache.PrivKey, payload)
	if err != nil {
		log.Println("Error signing delete cert: ", err)
	}
	delcert := []types.ModCert{{
		Sign:      sign,
		PublicKey: string(pubkey),
		Status:    "",
	},
	}
	fmt.Println("msgcert:", msgcert)
	repCert := core.CreateRepCert(msgcert, delcert, "delete")
	tsmin := msgcert.Msg.Ts - (msgcert.Msg.Ts % 60)
	key := util.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	core.SendToDb(key, repCert, "/route=delete")

	return fmt.Sprintf(":white_check_mark: Sent to DB. Time: %d", time.Now().Unix())
}

func (a *App) FetchAll() []types.RetMsgCert {
	messages := core.FetchRecent(context.Background())
	return messages
}

func (a *App) FetchMessageReports() []models.MsgCert {
	reports, err := moddb.GetUnmoderatedMsgs()
	if err != nil {
		log.Printf("Error fetching unmoderated messages: %v", err)
		return nil
	}
	return reports
}

func (a *App) ManualModerate(cert types.MsgCert, moderated int) {
	modsign, _ := moddb.ReportModSign(&cert, strconv.Itoa(moderated), keycache.PrivKey, keycache.PubKey)
	moddb.UpdateModerationStatus(cert.Sign, modsign, moderated)
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
	path := service.GetModKeysPath()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create modkeys directory: %w", err)
	}

	data := map[string]string{
		"GOOGLE_NLP_API_KEY": key,
	}

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode API key as JSON: %w", err)
	}

	err = os.WriteFile(path, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write modkeys.json: %w", err)
	}

	return nil
}

func (a *App) LogToFile(msg string) {
	logger.LogToFile(msg)
}
