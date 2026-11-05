package bridge

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	cryptoconfig "github.com/libr-forum/Libr/core/crypto/config"
	"github.com/libr-forum/Libr/core/crypto/cryptoutils"
	"github.com/libr-forum/Libr/core/mod_client/alias"
	"github.com/libr-forum/Libr/core/mod_client/avatar"
	"github.com/libr-forum/Libr/core/mod_client/config"
	"github.com/libr-forum/Libr/core/mod_client/core"
	cache "github.com/libr-forum/Libr/core/mod_client/core/cache_handler"
	"github.com/libr-forum/Libr/core/mod_client/identity"
	"github.com/libr-forum/Libr/core/mod_client/keycache"
	peer "github.com/libr-forum/Libr/core/mod_client/peers"
	"github.com/libr-forum/Libr/core/mod_client/types"
	util "github.com/libr-forum/Libr/core/mod_client/util"
)

var (
	chatPeer *peer.ChatPeer
	peerMu   sync.Mutex
	cancel   context.CancelFunc
)

// init ensures a default server URL is always set even if InitApp is not called.
func init() {
	util.InitServerClient(config.GetServerURL())
}

// InitNode initializes the libp2p node and connects to the provided relays.
// relayAddrs should be a JSON array of multiaddress strings.
func InitNode(relayAddrsJson string) string {
	peerMu.Lock()
	defer peerMu.Unlock()

	if chatPeer != nil {
		return "already_initialized"
	}

	var addrs []string
	if err := json.Unmarshal([]byte(relayAddrsJson), &addrs); err != nil {
		return fmt.Sprintf("error_unmarshal_addrs: %v", err)
	}

	var err error
	chatPeer, err = peer.NewChatPeer(addrs)
	if err != nil {
		return fmt.Sprintf("error_new_peer: %v", err)
	}

	ctx, c := context.WithCancel(context.Background())
	cancel = c

	if err := chatPeer.Start(ctx); err != nil {
		chatPeer = nil
		cancel()
		return fmt.Sprintf("error_start_peer: %v", err)
	}

	// Expose to peers package so GET/POST in functions.go work
	peer.Peer = chatPeer

	return "success"
}

// GetPeerID returns the local node's Peer ID.
func GetPeerID() string {
	peerMu.Lock()
	defer peerMu.Unlock()

	if chatPeer == nil {
		return ""
	}
	return chatPeer.Host.ID().String()
}

// StopNode stops the libp2p node.
func StopNode() {
	peerMu.Lock()
	defer peerMu.Unlock()

	if cancel != nil {
		cancel()
	}
	if chatPeer != nil {
		chatPeer.Close()
		chatPeer = nil
	}
}

// SendMessage sends a simple text message to a target peer.
func SendMessage(targetPeerID string, msg string) string {
	peerMu.Lock()
	defer peerMu.Unlock()

	if chatPeer == nil {
		return "error_not_initialized"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Wrap in typical Libr request format
	reqParams := map[string]string{
		"Method": "POST",
		"route":  "chat",
	}
	jsonReq, _ := json.Marshal(reqParams)

	bodyBytes, _ := json.Marshal(msg)
	resp, err := chatPeer.Send(ctx, targetPeerID, jsonReq, bodyBytes)
	if err != nil {
		return fmt.Sprintf("error_send: %v", err)
	}

	return string(resp)
}

// SetMessageHandler is an interface for receiving messages.
type MessageHandler interface {
	OnMessage(peerID string, msg string)
}

func SetMessageHandler(h MessageHandler) {
	peer.SetMessageHandler(h)
}

// NotificationHandler is an interface implemented in native Android code to show push notifications.
type NotificationHandler interface {
	OnNotification(title string, message string)
}

func SetNotificationHandler(h NotificationHandler) {
	core.SetNotificationHandler(h)
}

// ── Key management ────────────────────────────────────────────────────────────

// InitApp initialises keys and the discovery server client. Must be called once
// at startup before any other functions that depend on keys or the server.
// keyDir is the directory where keys are stored.
// serverURL overrides the default; pass "" to use the compiled-in default.
func InitApp(keyDir string, serverURL string) string {
	if keyDir != "" {
		cryptoconfig.SetKeyDir(keyDir)
	}
	if serverURL != "" {
		util.InitServerClient(serverURL)
	}
	keycache.InitKeys()
	cache.InitCacheFile()
	if err := identity.EnsureInitialized(); err != nil {
		return base64.StdEncoding.EncodeToString(keycache.PubKey)
	}
	return base64.StdEncoding.EncodeToString(keycache.PubKey)
}

// GetPublicKey returns the current Ed25519 public key as a base64 string.
func GetPublicKey() string {
	return keycache.LoadPubKey()
}

// RegenKeys regenerates the Ed25519 key pair and returns the new public key.
func RegenKeys() string {
	cryptoutils.GenerateKeyPair() // generates + persists to disk
	keycache.InitKeys()           // reload from disk
	_ = identity.SyncActiveIdentity()
	return base64.StdEncoding.EncodeToString(keycache.PubKey)
}

func IsIncognitoEnabled() bool {
	activeID, err := identity.GetActiveIdentityID()
	return err == nil && activeID != identity.MainIdentityID
}

func EnableIncognito() string {
	if IsIncognitoEnabled() {
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
	return pubKey
}

func DisableIncognito() string {
	if !IsIncognitoEnabled() {
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
	if activeID != identity.MainIdentityID && !cache.HasPendingModerationForIdentity(activeID) {
		_ = identity.DeleteIdentity(activeID)
	}
	return pubKey
}

// ── Discovery server helpers ──────────────────────────────────────────────────

// GetRelayAddresses fetches relay multiaddresses as a JSON array of strings.
// Returns an error string prefixed with "error:" on failure.
func GetRelayAddresses() string {
	addrs, err := util.GetRelayAddr()
	if err != nil {
		return fmt.Sprintf("error:%v", err)
	}
	b, _ := json.Marshal(addrs)
	return string(b)
}

// GetOnlineMods returns a JSON array of {peer_id, public_key} objects.
// Returns an error string prefixed with "error:" on failure.
func GetOnlineMods() string {
	mods, err := util.GetOnlineMods()
	if err != nil {
		return fmt.Sprintf("error:%v", err)
	}
	b, _ := json.Marshal(mods)
	return string(b)
}

// AmIMod returns true if the current key is in the mod allowlist.
func AmIMod() bool {
	pubB64 := base64.StdEncoding.EncodeToString(keycache.PubKey)
	ok, _ := util.AmIMod(pubB64)
	return ok
}

// ── Messaging ─────────────────────────────────────────────────────────────────

// SendTextMessage sends a plain-text message through the moderation pipeline.
// If an image is detected, it automatically switches to the "manual_mod" route.
// Returns a JSON-encoded types.SendResult, or an error string prefixed "error:".
func SendTextMessage(content string) string {
	ts := time.Now().Unix()

	// 0. Defensive Size Check: Enforce ~1MB limit (with small buffer for protocol/text overhead)
	if len(content) > 1100000 {
		return "error:Message too large (max 1MB for media content)"
	}

	// 1. Detection: Is this an image message?
	// Mobile encodes images as <img src="data:image/jpeg;base64,..."/> inside <BODY>.
	// Use strings.Contains so the check is reliable regardless of where in the
	// content the img tag appears (the byte-loop equivalent missed edge cases).
	isImage := strings.Contains(content, "data:image") || strings.Contains(content, "<img ")

	if isImage {
		// ── MANUAL ROUTE (Images) — mirrors desktop SendImageInput ──
		msgCert := types.MsgCert{
			PublicKey: base64.StdEncoding.EncodeToString(keycache.PubKey),
			Msg: types.Msg{
				Content: content,
				Ts:      ts,
			},
			Reason: "Image attached",
			Type:   "manual_mod",
		}

		// Sign it so it follows the protocol
		dataToSign := types.DataToSign{
			Content:   content,
			Timestamp: ts,
			ModCerts:  []types.ModCert{},
		}
		jsonBytes, _ := json.Marshal(dataToSign)
		_, sign, err := cryptoutils.SignMessage(keycache.PrivKey, string(jsonBytes))
		if err != nil {
			b, _ := json.Marshal(types.SendResult{Status: "error:signing"})
			return string(b)
		}
		msgCert.Sign = sign

		// Get online mods — treat fetch error the same as no mods available
		mods, err := util.GetOnlineMods()
		if err != nil || len(mods) == 0 {
			b, _ := json.Marshal(types.SendResult{Status: "error:no_mods"})
			return string(b)
		}

		// Wrap ManualSendToMods in a 4-second timeout, matching desktop SendImageInput
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		modChan := make(chan []types.ModCert, 1)
		go func() {
			modChan <- core.ManualSendToMods(msgCert, mods, "Image attached", true)
		}()

		var modcertlist []types.ModCert
		select {
		case modcertlist = <-modChan:
		case <-ctx.Done():
			b, _ := json.Marshal(types.SendResult{Status: "timeout"})
			return string(b)
		}

		// nil means ManualSendToMods could not reach any mod
		if modcertlist == nil {
			b, _ := json.Marshal(types.SendResult{Status: "error:no_mods"})
			return string(b)
		}

		// Empty list means all mods acknowledged — pending async review
		if len(modcertlist) == 0 {
			b, _ := json.Marshal(types.SendResult{
				Status: "pending_manual",
				Sign:   msgCert.Sign,
				Ts:     ts,
			})
			return string(b)
		}

		b, _ := json.Marshal(types.SendResult{
			Status:   "sent",
			ModCerts: modcertlist,
			Sign:     msgCert.Sign,
			Ts:       ts,
		})
		return string(b)
	}

	// ── AUTO ROUTE (Text) ──
	modcerts, err := core.AutoSendToMods(content, ts)
	if err != nil || len(modcerts) == 0 {
		status := "rejected"
		if err != nil {
			status = fmt.Sprintf("error:%v", err)
		}
		b, _ := json.Marshal(types.SendResult{Status: status})
		return string(b)
	}

	msgCert := core.CreateMsgCert(content, ts, modcerts)
	tsMin := msgCert.Msg.Ts - (msgCert.Msg.Ts % 60)
	key := util.GenerateNodeID(strconv.FormatInt(tsMin, 10))

	dbErr := core.SendToDb(key, msgCert, "/route=store")
	dbStatus := "stored"
	if dbErr != nil {
		dbStatus = fmt.Sprintf("db_error:%v", dbErr)
	} else {
		// Add to in-memory cache for immediate UI refresh
		cache.AddToSentCache(msgCert)
	}

	result := types.SendResult{
		Status:   fmt.Sprintf("sent:%s", dbStatus),
		ModCerts: modcerts,
		Sign:     msgCert.Sign,
		Ts:       msgCert.Msg.Ts,
	}
	b, _ := json.Marshal(result)
	return string(b)
}

// ResendToDb re-attempts to store an already approved MsgCert to the DHT.
func ResendToDb(msgCertJSON string) string {
	var msgCert types.MsgCert
	if err := json.Unmarshal([]byte(msgCertJSON), &msgCert); err != nil {
		b, _ := json.Marshal(types.SendResult{Status: fmt.Sprintf("error:invalid_json:%v", err)})
		return string(b)
	}

	tsMin := msgCert.Msg.Ts - (msgCert.Msg.Ts % 60)
	key := util.GenerateNodeID(strconv.FormatInt(tsMin, 10))

	dbErr := core.SendToDb(key, msgCert, "/route=store")
	dbStatus := "stored"
	if dbErr != nil {
		dbStatus = fmt.Sprintf("db_error:%v", dbErr)
	} else {
		// Add to in-memory cache for immediate UI refresh
		cache.AddToSentCache(msgCert)
	}

	result := types.SendResult{
		Status:   fmt.Sprintf("sent:%s", dbStatus),
		ModCerts: msgCert.ModCerts,
		Sign:     msgCert.Sign,
		Ts:       msgCert.Msg.Ts,
	}
	b, _ := json.Marshal(result)
	return string(b)
}

// FetchMessages retrieves messages from the last hour.
// Returns a JSON array of RetMsgCert objects, or "error:..." on failure.
func FetchMessages() string {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	certs := core.FetchRecent(ctx)
	b, err := json.Marshal(certs)
	if err != nil {
		return fmt.Sprintf("error:%v", err)
	}
	return string(b)
}

// ReportMessage sends a report for the message identified by msgSign.
// msgCertJSON should be a JSON-encoded MsgCert for the message.
// Returns "ok" on success, or an error string prefixed "error:".
func ReportMessage(msgCertJSON string, reason string) string {
	var msgCert types.MsgCert
	if err := json.Unmarshal([]byte(msgCertJSON), &msgCert); err != nil {
		return fmt.Sprintf("error:unmarshal msgcert: %v", err)
	}

	mods, err := util.GetOnlineMods()
	if err != nil || len(mods) == 0 {
		return "error:no online mods"
	}

	msgCert.Type = "report"
	if reason != "" {
		msgCert.Reason = reason
	}

	// firstTry = true to allow caching the new request
	repModCerts := core.ManualSendToMods(msgCert, mods, reason, true)
	if repModCerts == nil {
		return "error:no moderators reached"
	}

	return "ok"
}

// TickCron performs a moderation cron tick.
func TickCron() string {
	if peer.Peer == nil || peer.Peer.Host == nil {
		return "skipped:node_not_initialized"
	}
	core.MaybeStartCron()
	return "ok"
}

// StartCron initializes the moderation cron job to process pending reports.
// Call this when the app comes to the foreground.
func StartCron() string {
	if peer.Peer == nil || peer.Peer.Host == nil {
		return "skipped:node_not_initialized"
	}
	core.MaybeStartCron()
	return "ok"
}

// StopCron gracefully halts the moderation cron job.
// Call this when the app goes to the background.
func StopCron() string {
	core.StopModerationCron()
	return "ok"
}

// PendingReportStatus exposes the current moderation status to React Native
type PendingReportStatus struct {
	Total       int    `json:"total"`
	Approved    int    `json:"approved"`
	Rejected    int    `json:"rejected"`
	FinalStatus string `json:"final_status,omitempty"`
	Type        string `json:"type,omitempty"`
}

// GetPendingReports retrieves a map of msgSign -> status counts for all pending reports.
func GetPendingReports() string {
	pendings, err := cache.GetAllPendingModerations()
	if err != nil {
		return fmt.Sprintf("error:%v", err)
	}

	statusMap := make(map[string]PendingReportStatus)
	for _, p := range pendings {
		// Tally current partial certs
		var approved, rejected int
		for _, cert := range p.PartialCerts {
			if cert.Status == "1" {
				approved++
			} else if cert.Status == "0" {
				rejected++
			}
		}

		itemType := "report"
		if p.MsgCert.Type == "manual_mod" || p.MsgCert.Reason == "Image attached" {
			itemType = "manual_mod"
		}

		statusMap[p.MsgSign] = PendingReportStatus{
			Total:    p.AckCount, // Total expected responses based on Acks
			Approved: approved,
			Rejected: rejected,
			Type:     itemType,
		}
	}

	// Merge finalized moderations so React Native gets final decisions
	finalized, err := cache.GetAllFinalizedModerations()
	if err == nil {
		for _, f := range finalized {
			statusMap[f.MsgSign] = PendingReportStatus{
				Total:       f.Total,
				Approved:    f.Approved,
				Rejected:    f.Rejected,
				FinalStatus: f.Status,
				Type:        f.Type,
			}
		}
	}

	b, err := json.Marshal(statusMap)
	if err != nil {
		return fmt.Sprintf("error:%v", err)
	}
	return string(b)
}

// DismissReport deletes a finalized moderation from disk.
func DismissReport(msgSign string) string {
	if err := cache.DeleteFinalizedModeration(msgSign); err != nil {
		return fmt.Sprintf("error:%v", err)
	}
	return "ok"
}

// FetchReports retrieves recent reported messages for moderation.
func FetchReports() string {
	reports := core.FetchRecentReports(context.Background())
	b, err := json.Marshal(reports)
	if err != nil {
		return fmt.Sprintf("error:%v", err)
	}
	return string(b)
}

// ModerateMessage allows a moderator to approve or reject a reported message.
func ModerateMessage(msgCertJSON string, action string) string {
	var msgCert types.MsgCert
	if err := json.Unmarshal([]byte(msgCertJSON), &msgCert); err != nil {
		return fmt.Sprintf("error:unmarshal msgcert: %v", err)
	}

	status := "0" // reject
	if action == "approve" {
		status = "1"
	}

	mods, err := util.GetOnlineMods()
	if err != nil || len(mods) == 0 {
		return "error:no online mods"
	}

	// Sign the moderation action
	payload := msgCert.Sign + status
	_, sign, err := cryptoutils.SignMessage(keycache.PrivKey, payload)
	if err != nil {
		return fmt.Sprintf("error:signing: %v", err)
	}

	modCert := types.ModCert{
		Sign:      sign,
		PublicKey: base64.StdEncoding.EncodeToString(keycache.PubKey),
		Status:    status,
	}

	repCert := core.CreateRepCert(msgCert, []types.ModCert{modCert}, "moderation")

	ts := msgCert.Msg.Ts - (msgCert.Msg.Ts % 60)
	key := util.GenerateNodeID(strconv.FormatInt(ts, 10))
	if err := core.SendToDb(key, repCert, "/route=moderation"); err != nil {
		return fmt.Sprintf("error:send to db: %v", err)
	}

	return "ok"
}

// DeleteMessage sends a delete request for the message identified by msgCertJSON.
// Returns "ok" on success, or an error string prefixed "error:".
func DeleteMessage(msgCertJSON string) string {
	var msgcert types.MsgCert
	if err := json.Unmarshal([]byte(msgCertJSON), &msgcert); err != nil {
		return fmt.Sprintf("error:unmarshal msgcert: %v", err)
	}

	payload := msgcert.Sign
	pubkey, sign, err := cryptoutils.SignMessage(keycache.PrivKey, payload)
	if err != nil {
		return fmt.Sprintf("error:signing delete cert: %v", err)
	}
	delcert := []types.ModCert{{
		Sign:      sign,
		PublicKey: string(pubkey),
		Status:    "",
	}}
	repCert := core.CreateRepCert(msgcert, delcert, "delete")
	tsmin := msgcert.Msg.Ts - (msgcert.Msg.Ts % 60)
	key := util.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	if err := core.SendToDb(key, repCert, "/route=delete"); err != nil {
		return fmt.Sprintf("error:send to db: %v", err)
	}

	return "ok"
}

// ── Identity helpers ──────────────────────────────────────────────────────────

// GenerateAlias returns a human-readable alias for a base64-encoded public key.
func GenerateAlias(key string) string {
	if len(key) != 44 {
		return "…"
	}

	// 1. Check cache
	if entry, err := cache.GetFromCache(key); err == nil && entry != nil && entry.Alias != "" {
		return entry.Alias
	}

	// 2. Generate new
	res := alias.GenerateAlias(key)

	// 3. Save to cache (preserving existing avatar if any)
	var existingAvatar string
	if entry, err := cache.GetFromCache(key); err == nil && entry != nil {
		existingAvatar = entry.AvatarSVG
	}
	_ = cache.AddToCache(key, existingAvatar, res)

	return res
}

// GenerateAvatar returns a base64-encoded SVG avatar for a base64-encoded public key.
func GenerateAvatar(key string) string {
	if len(key) != 44 {
		return ""
	}

	// 1. Check cache
	if entry, err := cache.GetFromCache(key); err == nil && entry != nil && entry.AvatarSVG != "" {
		return base64.StdEncoding.EncodeToString([]byte(entry.AvatarSVG))
	}

	// 2. Generate new
	svg := avatar.GenerateAvatar(key)

	// 3. Save to cache (preserving existing alias if any)
	var existingAlias string
	if entry, err := cache.GetFromCache(key); err == nil && entry != nil {
		existingAlias = entry.Alias
	}
	_ = cache.AddToCache(key, svg, existingAlias)

	return base64.StdEncoding.EncodeToString([]byte(svg))
}
