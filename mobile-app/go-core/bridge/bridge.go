// Package bridge is the gomobile API of the Android app: a thin layer over the shared
// v2 node (core/libr/app). Every parameter is a string (Kotlin calls by reflection
// with boxed types, so bool/int parameters would never resolve), and results are JSON
// or "ok"; failures are "error:<message>".
package bridge

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/libr-forum/Libr/core/libr/alias"
	libr "github.com/libr-forum/Libr/core/libr/app"
	"github.com/libr-forum/Libr/core/libr/avatar"
	cache "github.com/libr-forum/Libr/core/libr/cache"
	cryptoconfig "github.com/libr-forum/Libr/core/libr/crypto/config"
	"github.com/libr-forum/Libr/core/libr/crypto/cryptoutils"
	"github.com/libr-forum/Libr/core/libr/identity"
	"github.com/libr-forum/Libr/core/libr/keycache"
	"github.com/libr-forum/Libr/core/libr/proto"
)

var (
	mu       sync.Mutex
	node     *libr.App
	keyDir   string
	lastConf string // StartNode config, reused after an identity change
	notify   NotificationHandler
)

func errStr(err error) string { return "error:" + err.Error() }

func jsonStr(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return errStr(err)
	}
	return string(b)
}

// ── Keys and identity ─────────────────────────────────────────────────────────

// InitApp loads (or creates) the user's keys under dir and returns the public key.
// Call once at startup.
func InitApp(dir string) string {
	if dir != "" {
		keyDir = dir
		cryptoconfig.SetKeyDir(dir)
		cache.SetCacheDir(dir) // mobile keeps caches next to the keys
	}
	keycache.InitKeys()
	cache.InitCacheFile()
	_ = identity.EnsureInitialized()
	return base64.StdEncoding.EncodeToString(keycache.PubKey)
}

// GetPublicKey returns the active identity's public key (base64).
func GetPublicKey() string { return keycache.LoadPubKey() }

// RegenKeys replaces the key pair and reconnects under it.
func RegenKeys() string {
	pub, _, _ := cryptoutils.GenerateKeyPair()
	keycache.InitKeys()
	_ = identity.SyncActiveIdentity()
	restart()
	return base64.StdEncoding.EncodeToString(pub)
}

func IsIncognitoEnabled() bool {
	activeID, err := identity.GetActiveIdentityID()
	return err == nil && activeID != identity.MainIdentityID
}

func EnableIncognito() string {
	if IsIncognitoEnabled() {
		return base64.StdEncoding.EncodeToString(keycache.PubKey)
	}
	id, _, err := identity.CreateIncognitoIdentity()
	if err != nil {
		return fmt.Sprintf("error:create incognito identity: %v", err)
	}
	pub, err := identity.ActivateIdentity(id)
	if err != nil {
		_ = identity.DeleteIdentity(id)
		return fmt.Sprintf("error:activate incognito identity: %v", err)
	}
	restart()
	return pub
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
	pub, err := identity.ActivateIdentity(identity.MainIdentityID)
	if err != nil {
		return fmt.Sprintf("error:restore main identity: %v", err)
	}
	restart()
	if pending, _ := filepath.Glob(filepath.Join(dataDir(activeID), "pending", "*.json")); len(pending) == 0 {
		_ = identity.DeleteIdentity(activeID)
		_ = os.RemoveAll(dataDir(activeID))
	}
	return pub
}

// dataDir is per identity, so an incognito identity's pending reports are never
// resumed (signed) by the main one.
func dataDir(identityID string) string { return filepath.Join(keyDir, "v2", identityID) }

func activeDataDir() string {
	id, err := identity.GetActiveIdentityID()
	if err != nil {
		id = identity.MainIdentityID
	}
	return dataDir(id)
}

// ── Network ───────────────────────────────────────────────────────────────────

// nodeConfig is StartNode's JSON argument.
type nodeConfig struct {
	Relays    []string `json:"relays"`
	Bootstrap []string `json:"bootstrap"`
	SeedURLs  []string `json:"seed_urls"`
	// Communities are genesis files (proto.CommunityFile) to install before starting,
	// e.g. a development network's; shipped ones are built in.
	Communities []json.RawMessage `json:"communities"`
}

// StartNode joins the LIBR network as the active identity. configJSON is
// {"relays":[...],"bootstrap":[...],"seed_urls":[...],"communities":[...]}; empty
// lists fall back to the built-in seeds. Returns "ok".
func StartNode(configJSON string) string {
	mu.Lock()
	defer mu.Unlock()
	if node != nil {
		return "ok"
	}
	var cfg nodeConfig
	if configJSON != "" {
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return fmt.Sprintf("error:config: %v", err)
		}
	}
	dir := activeDataDir()
	if err := installCommunities(dir, cfg.Communities); err != nil {
		return errStr(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	n, err := libr.Start(ctx, libr.Config{
		DataDir: dir, Key: keycache.PrivKey,
		Relays: cfg.Relays, Bootstrap: cfg.Bootstrap, SeedURLs: cfg.SeedURLs,
		OnFinalized: finalized,
	})
	if err != nil {
		return errStr(err)
	}
	node, lastConf = n, configJSON
	return "ok"
}

func installCommunities(dir string, files []json.RawMessage) error {
	if len(files) == 0 {
		return nil
	}
	cdir := filepath.Join(dir, "communities")
	if err := os.MkdirAll(cdir, 0o700); err != nil {
		return err
	}
	for _, raw := range files {
		var cf proto.CommunityFile
		if err := json.Unmarshal(raw, &cf); err != nil || len(cf.Chain) == 0 {
			return fmt.Errorf("community file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(cdir, cf.Chain[0].Community+".json"), raw, 0o600); err != nil {
			return err
		}
	}
	return nil
}

// StopNode leaves the network.
func StopNode() {
	mu.Lock()
	defer mu.Unlock()
	if node != nil {
		node.Close()
		node = nil
	}
}

func restart() {
	mu.Lock()
	conf := lastConf
	mu.Unlock()
	StopNode()
	if conf != "" {
		go StartNode(conf)
	}
}

var errOffline = errors.New("not connected")

func current() (*libr.App, error) {
	mu.Lock()
	defer mu.Unlock()
	if node == nil {
		return nil, errOffline
	}
	return node, nil
}

func opCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 90*time.Second)
}

// GetPeerID returns this node's peer ID ("" when offline).
func GetPeerID() string {
	n, err := current()
	if err != nil {
		return ""
	}
	return n.PeerID()
}

// NetworkStatus returns {"online","peers","routing_table","reservations"}.
func NetworkStatus() string {
	n, err := current()
	if err != nil {
		return jsonStr(libr.Status{})
	}
	return jsonStr(n.Status())
}

// ── Communities, posts, removal ───────────────────────────────────────────────

// ListCommunities returns [{name,is_mod,mods,threshold,mod_keys}].
func ListCommunities() string {
	n, err := current()
	if err != nil {
		return errStr(err)
	}
	return jsonStr(n.Communities())
}

// IsMod reports whether the active identity moderates community.
func IsMod(community string) bool {
	n, err := current()
	return err == nil && n.IsMod(community)
}

// Post publishes content to community: {"status":"posted|pending|rejected","hash","ts","id"}.
func Post(community, content string) string {
	n, err := current()
	if err != nil {
		return errStr(err)
	}
	ctx, cancel := opCtx()
	defer cancel()
	out, err := n.Post(ctx, community, content)
	if err != nil {
		return errStr(err)
	}
	return jsonStr(out)
}

// Feed returns the last 24 hours of community: [{hash,msg_id,community,author,ts,content,votes,mine}].
func Feed(community string) string {
	n, err := current()
	if err != nil {
		return errStr(err)
	}
	ctx, cancel := opCtx()
	defer cancel()
	posts, err := n.Feed(ctx, community, 24*time.Hour)
	if err != nil {
		return errStr(err)
	}
	return jsonStr(posts)
}

// Report asks moderators to remove the post with hash: {"status":"removed|pending|rejected"}.
func Report(hash, reason string) string {
	n, err := current()
	if err != nil {
		return errStr(err)
	}
	ctx, cancel := opCtx()
	defer cancel()
	out, err := n.Report(ctx, hash, reason)
	if err != nil {
		return errStr(err)
	}
	if out.Status != "pending" { // decided at once: record it like a finished review
		recordFinal(hash, "report", map[string]string{"removed": "approved"}[out.Status])
	}
	return jsonStr(out)
}

// Delete withdraws the user's own post. Returns "ok".
func Delete(hash string) string {
	n, err := current()
	if err != nil {
		return errStr(err)
	}
	ctx, cancel := opCtx()
	defer cancel()
	if err := n.Delete(ctx, hash); err != nil {
		return errStr(err)
	}
	return "ok"
}

// Tick retries every pending post and report now (the node also does so every 30 s).
func Tick() string {
	n, err := current()
	if err != nil {
		return errStr(err)
	}
	ctx, cancel := opCtx()
	defer cancel()
	n.Tick(ctx)
	return "ok"
}

// ── Pending items (the user's posts and reports waiting for moderators) ───────

// status is the per-item shape the app's screens use, keyed by post hash (reports)
// or pending ID (posts).
type status struct {
	Total       int    `json:"total"`
	Approved    int    `json:"approved"`
	Rejected    int    `json:"rejected"`
	FinalStatus string `json:"final_status,omitempty"` // approved | rejected
	Type        string `json:"type,omitempty"`         // report | manual_mod
}

func kindType(kind string) string {
	if kind == "report" {
		return "report"
	}
	return "manual_mod"
}

var finalMu sync.Mutex

func finalPath() string { return filepath.Join(activeDataDir(), "finalized.json") }

func loadFinal() map[string]status {
	m := map[string]status{}
	if b, err := os.ReadFile(finalPath()); err == nil {
		_ = json.Unmarshal(b, &m)
	}
	return m
}

func recordFinal(key, kind, final string) {
	if final == "" {
		final = "rejected"
	}
	finalMu.Lock()
	defer finalMu.Unlock()
	m := loadFinal()
	m[key] = status{FinalStatus: final, Type: kindType(kind)}
	if b, err := json.Marshal(m); err == nil {
		_ = os.MkdirAll(filepath.Dir(finalPath()), 0o700)
		_ = os.WriteFile(finalPath(), b, 0o600)
	}
}

// PendingStatus returns {key: {total,approved,rejected,final_status?,type}} for
// everything waiting for moderators plus finished items not yet dismissed.
func PendingStatus() string {
	finalMu.Lock()
	out := loadFinal()
	finalMu.Unlock()
	n, err := current()
	if err == nil {
		mods := map[string]int{}
		for _, c := range n.Communities() {
			mods[c.Name] = c.Mods
		}
		for _, p := range n.Pending() {
			out[p.Key] = status{Total: mods[p.Community], Approved: p.Approved, Rejected: p.Rejected, Type: kindType(p.Kind)}
		}
	}
	return jsonStr(out)
}

// DismissReport forgets a finished item.
func DismissReport(key string) string {
	finalMu.Lock()
	defer finalMu.Unlock()
	m := loadFinal()
	delete(m, key)
	b, _ := json.Marshal(m)
	if err := os.WriteFile(finalPath(), b, 0o600); err != nil && !os.IsNotExist(err) {
		return errStr(err)
	}
	return "ok"
}

// NotificationHandler is implemented in Kotlin to show system notifications.
type NotificationHandler interface {
	OnNotification(title string, message string)
}

// SetNotificationHandler receives a notification whenever a pending item finishes.
func SetNotificationHandler(h NotificationHandler) {
	mu.Lock()
	notify = h
	mu.Unlock()
}

func finalized(p libr.PendingInfo) {
	final := "rejected"
	if p.Status == "posted" || p.Status == "removed" {
		final = "approved"
	}
	recordFinal(p.Key, p.Kind, final)
	titles := map[string][2]string{
		"post/posted":     {"Post Approved", "Moderators approved your post; it is now in the feed."},
		"post/rejected":   {"Post Rejected", "Moderators rejected your post."},
		"post/expired":    {"Post Expired", "Moderators did not review your post within a week."},
		"report/removed":  {"Report Upheld", "A reported post was removed."},
		"report/rejected": {"Report Dismissed", "Moderators decided to keep the reported post."},
		"report/expired":  {"Report Expired", "Moderators did not review your report within a week."},
	}
	mu.Lock()
	h := notify
	mu.Unlock()
	if t, ok := titles[p.Kind+"/"+p.Status]; ok && h != nil {
		h.OnNotification(t[0], t[1])
	}
}

// ── Moderation (when the active identity is in a moderator set) ───────────────

// ModQueue returns the human review queue: [{key,kind,cert,epoch,reason,received,vote}].
func ModQueue() string {
	n, err := current()
	if err != nil {
		return errStr(err)
	}
	items := n.ModQueue()
	if items == nil {
		return "[]"
	}
	return jsonStr(items)
}

// ModDecide records a decision on a queued item; decision is "approve" or "reject".
// For a post approve publishes it; for a report approve removes the post.
func ModDecide(key, decision string) string {
	n, err := current()
	if err != nil {
		return errStr(err)
	}
	if decision != "approve" && decision != "reject" {
		return "error:decision must be approve or reject"
	}
	if err := n.ModDecide(key, decision == "approve"); err != nil {
		return errStr(err)
	}
	return "ok"
}

// ── Aliases and avatars ───────────────────────────────────────────────────────

// GenerateAlias returns a human-readable alias for a base64-encoded public key.
func GenerateAlias(key string) string {
	if len(key) != 44 {
		return "…"
	}
	if entry, err := cache.GetFromCache(key); err == nil && entry != nil && entry.Alias != "" {
		return entry.Alias
	}
	res := alias.GenerateAlias(key)
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
	if entry, err := cache.GetFromCache(key); err == nil && entry != nil && entry.AvatarSVG != "" {
		return base64.StdEncoding.EncodeToString([]byte(entry.AvatarSVG))
	}
	svg := avatar.GenerateAvatar(key)
	var existingAlias string
	if entry, err := cache.GetFromCache(key); err == nil && entry != nil {
		existingAlias = entry.Alias
	}
	_ = cache.AddToCache(key, svg, existingAlias)
	return base64.StdEncoding.EncodeToString([]byte(svg))
}
