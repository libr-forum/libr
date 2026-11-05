package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/libr-forum/Libr/core/crypto/cryptoutils"
	"github.com/libr-forum/Libr/core/mod_client/alias"
	"github.com/libr-forum/Libr/core/mod_client/avatar"
	cache "github.com/libr-forum/Libr/core/mod_client/cache_handler"
	"github.com/libr-forum/Libr/core/mod_client/config"
	"github.com/libr-forum/Libr/core/mod_client/core"
	"github.com/libr-forum/Libr/core/mod_client/identity"
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
	isMod       bool
}

func NewApp() *App {
	cache.InitCacheFile()
	keycache.InitKeys()
	_ = identity.EnsureInitialized()
	config.LoadConfig()
	util.InitServerClient(config.GetServerURL())
	amImod, _ := util.AmIMod(base64.StdEncoding.EncodeToString(keycache.PubKey))
	if amImod {
		config.InitDB()
	}
	core.MaybeStartCron()
	return &App{relayStatus: "offline", isMod: amImod}
}

// SetContext passes the Wails context from main to core for event emission
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	core.WailsCtx = ctx
	logger.LogToFile("App started")
	runtime.WindowMaximise(ctx)
	go func() {
		runtime.EventsEmit(ctx, "navigate-to-root")
	}()
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

func (a *App) RegenKeys() string {
	pub, _, _ := cryptoutils.GenerateKeyPair()
	keycache.InitKeys()
	_ = identity.SyncActiveIdentity()
	a.refreshModeratorState()
	return base64.StdEncoding.EncodeToString(pub)
}

func (a *App) refreshModeratorState() bool {
	amImod, _ := util.AmIMod(base64.StdEncoding.EncodeToString(keycache.PubKey))
	a.isMod = amImod
	return amImod
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

	a.refreshModeratorState()
	return pubKey
}

func (a *App) DisableIncognito() string {
	if !a.IsIncognitoEnabled() {
		keycache.InitKeys()
		a.refreshModeratorState()
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

	a.refreshModeratorState()
	if activeID != identity.MainIdentityID && !cache.HasPendingModerationForIdentity(activeID) {
		_ = identity.DeleteIdentity(activeID)
	}

	return pubKey
}

func (a *App) Connect(relayAdds []string) error {
	if len(relayAdds) == 0 {
		return fmt.Errorf("No relay addresses provided")
	}
	err := Peers.StartNode(relayAdds, func(status string) {
		a.relayStatus = status
		runtime.EventsEmit(a.ctx, "relay_status_changed", status)
	})
	if err != nil {
		a.relayStatus = "offline"
		runtime.EventsEmit(a.ctx, "relay_status_changed", "offline")
		return err
	}
	a.relayStatus = "online"
	runtime.EventsEmit(a.ctx, "relay_status_changed", "online")

	if a.isMod && Peers.Peer != nil {
		peerId := Peers.Peer.Host.ID().String()
		pubKeyB64 := base64.StdEncoding.EncodeToString(keycache.PubKey)
		if err := util.RegisterAsMod(peerId, pubKeyB64, keycache.PrivKey); err != nil {
			log.Printf("⚠️ RegisterAsMod failed: %v", err)
		} else {
			core.StartRefreshCron(pubKeyB64, keycache.PrivKey, config.RegistryRefreshSeconds)
		}
	}
	return nil
}

// shutdown is called by Wails before the window closes.
// It deregisters the mod from the discovery server (best-effort).
func (a *App) shutdown(_ context.Context) bool {
	if a.isMod {
		pubKeyB64 := base64.StdEncoding.EncodeToString(keycache.PubKey)
		core.StopRefreshCron()
		if err := util.DeregisterAsMod(pubKeyB64, keycache.PrivKey); err != nil {
			log.Printf("⚠️ DeregisterAsMod on shutdown failed: %v", err)
		}
	}
	return false // do not prevent the window from closing
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

func (a *App) SendInput(input string) (types.SendResult, error) {
	if a.relayStatus != "online" {
		return types.SendResult{Status: "offline"}, nil
	}

	// Defensive Size Check: Enforce ~1MB limit
	if len(input) > 1100000 {
		return types.SendResult{Status: "error"}, fmt.Errorf("message too large (max 1MB)")
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
		return types.SendResult{Status: "timeout"}, nil
	}

	var modcertlist []types.ModCert
	select {
	case modcertlist = <-modChan:
	case <-ctx.Done():
		return types.SendResult{Status: "timeout"}, nil
	}

	if len(modcertlist) == 0 {
		return types.SendResult{Status: "rejected"}, nil
	}

	fmt.Println("ModCerts received:", modcertlist)

	msgCert := core.CreateMsgCert(input, ts, modcertlist)
	tsmin := msgCert.Msg.Ts - (msgCert.Msg.Ts % 60)
	key := util.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	err = core.SendToDb(key, msgCert, "/route=store")
	if err != nil {
		return types.SendResult{
			Status:   "storage_failed",
			ModCerts: modcertlist,
			Sign:     msgCert.Sign,
			Ts:       msgCert.Msg.Ts,
		}, nil
	}

	// Add to in-memory cache for immediate UI refresh
	cache.AddToSentCache(msgCert)

	return types.SendResult{
		Status:   "sent",
		ModCerts: modcertlist,
		Sign:     msgCert.Sign,
		Ts:       msgCert.Msg.Ts,
	}, nil
}

func (a *App) ResendToDb(msgCert types.MsgCert) (types.SendResult, error) {
	tsmin := msgCert.Msg.Ts - (msgCert.Msg.Ts % 60)
	key := util.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	err := core.SendToDb(key, msgCert, "/route=store")
	if err != nil {
		return types.SendResult{
			Status:   "storage_failed",
			ModCerts: msgCert.ModCerts,
			Sign:     msgCert.Sign,
			Ts:       msgCert.Msg.Ts,
		}, nil
	}

	// Add to in-memory cache for immediate UI refresh
	cache.AddToSentCache(msgCert)

	return types.SendResult{
		Status:   "sent",
		ModCerts: msgCert.ModCerts,
		Sign:     msgCert.Sign,
		Ts:       msgCert.Msg.Ts,
	}, nil
}

func (a *App) SendImageInput(input string) types.SendResult {
	if a.relayStatus != "online" {
		return types.SendResult{Status: "offline"}
	}

	// Defensive Size Check: Enforce ~1MB limit
	if len(input) > 1100000 {
		return types.SendResult{Status: "error"}
	}

	ts := time.Now().Unix()

	// 1. Create a raw MsgCert with empty ModCerts list and our own pubkey
	msgCert := types.MsgCert{
		PublicKey: base64.StdEncoding.EncodeToString(keycache.PubKey),
		Msg: types.Msg{
			Content: input,
			Ts:      ts,
		},
		Reason: "Image attached",
		Type:   "manual_mod",
	}

	// Calculate its signature right away so manual mod queue tracks it cleanly
	dataToSign := types.DataToSign{
		Content:   input,
		Timestamp: ts,
		ModCerts:  []types.ModCert{},
	}
	jsonBytes, _ := json.Marshal(dataToSign)
	_, sign, err := cryptoutils.SignMessage(keycache.PrivKey, string(jsonBytes))
	if err != nil {
		return types.SendResult{Status: "error"}
	}
	msgCert.Sign = sign

	// 2. Look up mods
	mods, err := util.GetOnlineMods()
	if err != nil || len(mods) == 0 {
		return types.SendResult{Status: "No online moderators available"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	modChan := make(chan []types.ModCert, 1)
	go func() {
		modcerts := core.ManualSendToMods(msgCert, mods, "Image attached", true)
		modChan <- modcerts
	}()

	var modcertlist []types.ModCert
	select {
	case modcertlist = <-modChan:
	case <-ctx.Done():
		return types.SendResult{Status: "timeout"}
	}

	if modcertlist == nil {
		return types.SendResult{Status: "No moderators available"}
	}

	// modcertlist empty means it successfully lodged into pending queue but no immediate signatures
	if len(modcertlist) == 0 {
		return types.SendResult{
			Status: "pending_manual",
			Sign:   msgCert.Sign,
			Ts:     msgCert.Msg.Ts,
		}
	}

	// If by some miracle they all rejected it synchronously:
	return types.SendResult{
		Status:   "sent",
		ModCerts: modcertlist,
		Sign:     msgCert.Sign,
		Ts:       msgCert.Msg.Ts,
	}
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

	var reasonStr string
	if reason != nil {
		reasonStr = *reason
	}
	msgcert.Type = "report"

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	modChan := make(chan []types.ModCert, 1)
	mods, _ := util.GetOnlineMods()
	go func() {
		modcerts := core.ManualSendToMods(msgcert, mods, reasonStr, true)
		modChan <- modcerts
	}()

	var modcertlist []types.ModCert
	select {
	case modcertlist = <-modChan:
	case <-ctx.Done():
		return ":x: Moderator timeout"
	}

	// Empty modcertlist means all mods acknowledged (pending manual review) — not a rejection.
	// A nil modcertlist (no mods reachable at all) is the real failure case.
	if modcertlist == nil {
		return ":x: No moderators available to process this report."
	}

	if len(modcertlist) == 0 {
		return ":white_check_mark: Report submitted. Awaiting moderator review."
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

// IsModerationCronRunning exposes the cron status to the frontend via Wails.
func (a *App) IsModerationCronRunning() bool {
	return core.IsModerationCronRunning()
}

// ModerateBySign is a simpler Wails-callable that takes just the sign string,
// avoiding any struct-mapping ambiguity across MsgCert type variants.
func (a *App) ModerateBySign(sign string, moderated int) {
	// Fetch full message record so we can choose the correct signing payload.
	content, ts, msgType, err := moddb.GetMsgBySign(sign)
	if err != nil {
		log.Printf("[ModerateBySign] db lookup error: %v", err)
	}
	cert := types.MsgCert{
		Sign: sign,
		Type: msgType,
		Msg: types.Msg{
			Content: content,
			Ts:      ts,
		},
	}
	modsign, err := moddb.ReportModSign(&cert, strconv.Itoa(moderated), keycache.PrivKey, keycache.PubKey)
	if err != nil {
		log.Printf("[ModerateBySign] signing error: %v", err)
	}
	if _, err := moddb.UpdateModerationStatus(sign, modsign, moderated); err != nil {
		log.Printf("[ModerateBySign] update error: %v", err)
	} else {
		log.Printf("[ModerateBySign] sign=%s moderated=%d updated OK", sign, moderated)
	}
}

// PendingItemStat is returned by GetPendingModerationStats for each pending message.
type PendingItemStat struct {
	MsgSign  string `json:"msg_sign"`
	Ts       int64  `json:"ts"`
	Content  string `json:"content"`
	Reason   string `json:"reason"`
	Approved int    `json:"approved"`
	Rejected int    `json:"rejected"`
	Awaiting int    `json:"awaiting"`
	IsImage  bool   `json:"is_image"`
}

// PendingModerationStats is the full response from GetPendingModerationStats.
type PendingModerationStats struct {
	Items      []PendingItemStat `json:"items"`
	CronActive bool              `json:"cron_active"`
}

// GetPendingModerationStats returns vote tallies for all pending moderation files.
func (a *App) GetPendingModerationStats() PendingModerationStats {
	pattern := filepath.Join(cache.GetCacheDir(), "pending_mods", "*.json")
	files, _ := filepath.Glob(pattern)

	stats := PendingModerationStats{
		Items:      []PendingItemStat{},
		CronActive: core.IsModerationCronRunning(),
	}

	for _, f := range files {
		pending, err := cache.LoadPendingModeration(f)
		if err != nil {
			continue
		}
		var approved, rejected int
		for _, cert := range pending.PartialCerts {
			switch cert.Status {
			case "1":
				approved++
			case "0":
				rejected++
			}
		}
		stats.Items = append(stats.Items, PendingItemStat{
			MsgSign:  pending.MsgSign,
			Ts:       pending.MsgCert.Msg.Ts,
			Content:  pending.MsgCert.Msg.Content,
			Reason:   pending.MsgCert.Reason,
			Approved: approved,
			Rejected: rejected,
			Awaiting: len(pending.AwaitingMods),
			IsImage:  pending.MsgCert.Reason == "Image attached",
		})
	}
	return stats
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