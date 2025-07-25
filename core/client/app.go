package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/devlup-labs/Libr/core/client/alias"
	"github.com/devlup-labs/Libr/core/client/avatar"
	"github.com/devlup-labs/Libr/core/client/cache"
	"github.com/devlup-labs/Libr/core/client/core"
	"github.com/devlup-labs/Libr/core/client/keycache"
	Peers "github.com/devlup-labs/Libr/core/client/peers"
	"github.com/devlup-labs/Libr/core/client/types"
	util "github.com/devlup-labs/Libr/core/client/util"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	relayStatus string
}

func NewApp() *App {
	cache.InitCacheFile()
	keycache.InitKeys()
	return &App{relayStatus: "offline"}
}
func (a *App) FetchPubKey() string {
	pubStr := keycache.LoadPubKey()
	return pubStr
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
}

func (a *App) Connect(relayAdd string) string {
	err := Peers.StartNode(relayAdd)
	if err != nil {
		a.relayStatus = "offline"
		return err.Error()
	}
	a.relayStatus = "online"
	return "Online"
}

func (a *App) GetRelayStatus() string {
	return a.relayStatus
}

func (a *App) SendInput(input string) string {
	if a.relayStatus != "online" {
		return "Offline"
	}

	ts := time.Now().Unix()

	// Optional: Add timeout for whole SendToMods process (not just per mod)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// Run SendToMods with timeout
	modChan := make(chan []types.ModCert, 1)

	go func() {
		modcerts := core.SendToMods(input, ts)
		modChan <- modcerts
	}()

	var modcertlist []types.ModCert
	select {
	case modcertlist = <-modChan:
	case <-ctx.Done():
		return "❌ Moderator timeout"
	}

	if len(modcertlist) == 0 {
		return "❌ Message rejected by moderators."
	}

	msgCert := core.CreateMsgCert(input, ts, modcertlist)
	tsmin := msgCert.Msg.Ts - (msgCert.Msg.Ts % 60)
	key := util.GenerateNodeID(strconv.FormatInt(tsmin, 10))
	core.SendToDb(key, msgCert)

	return fmt.Sprintf("✅ Sent to DB. Time: %d", msgCert.Msg.Ts)
}

func (a *App) FetchAll() []string {
	messages := core.FetchRecent(context.Background())
	var out []string
	for _, cert := range messages {
		out = append(out, fmt.Sprintf("Sender: %s | Msg: %s | Time: %d", cert.PublicKey, cert.Msg.Content, cert.Msg.Ts))
	}
	return out
}

func (a *App) StreamMessages() {
	ctx := context.Background()

	msgChan := core.FetchRecentStreamed(ctx)

	go func() {
		for cert := range msgChan {
			msg := map[string]interface{}{
				"sender":    cert.PublicKey,
				"content":   cert.Msg.Content,
				"timestamp": cert.Msg.Ts,
			}
			runtime.EventsEmit(a.ctx, "newMessage", msg)

		}
	}()
}

func (a *App) FetchMessagesByDate(ts int64) {
	ctx := context.Background()

	msgChan := core.FetchRecentStreamed(ctx)

	go func() {
		for cert := range msgChan {
			msg := map[string]interface{}{
				"sender":    cert.PublicKey,
				"content":   cert.Msg.Content,
				"timestamp": cert.Msg.Ts,
			}
			runtime.EventsEmit(a.ctx, "newMessage", msg)
		}
	}()
}

func (a *App) FetchTimestamp(tsStr string) []string {
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return []string{"❌ Invalid timestamp format"}
	}

	messages := core.Fetch(ts)

	var out []string
	for _, cert := range messages {
		out = append(out, fmt.Sprintf("Sender: %s | Msg: %s | Time: %d", cert.PublicKey, cert.Msg.Content, cert.Msg.Ts))
	}
	return out
}
