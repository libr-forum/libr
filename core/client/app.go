package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/devlup-labs/Libr/core/client/core"
	"github.com/devlup-labs/Libr/core/client/keycache"
	Peers "github.com/devlup-labs/Libr/core/client/peers"
	"github.com/devlup-labs/Libr/core/client/types"
	util "github.com/devlup-labs/Libr/core/client/util"
)

type App struct {
	ctx         context.Context
	relayStatus string
}

func NewApp() *App {
	keycache.InitKeys()
	return &App{relayStatus: "offline"}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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
	key := util.GenerateNodeID(strconv.FormatInt(msgCert.Msg.Ts, 10))
	core.SendToDb(key, msgCert)

	return fmt.Sprintf("✅ Sent to DB. Time: %d", msgCert.Msg.Ts)
}

func (a *App) FetchAll() []string {
	messages := core.FetchRecent(context.Background())
	myKey := base64.StdEncoding.EncodeToString(keycache.PubKey)

	var out []string
	for _, cert := range messages {
		sender := cert.PublicKey
		if sender == myKey {
			sender = "You"
		}
		out = append(out, fmt.Sprintf("Sender: %s | Msg: %s | Time: %d", sender, cert.Msg.Content, cert.Msg.Ts))
	}
	return out
}

func (a *App) FetchTimestamp(tsStr string) []string {
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return []string{"❌ Invalid timestamp format"}
	}

	messages := core.Fetch(ts)
	myKey := base64.StdEncoding.EncodeToString(keycache.PubKey)

	var out []string
	for _, cert := range messages {
		sender := cert.PublicKey
		if sender == myKey {
			sender = "You"
		}
		out = append(out, fmt.Sprintf("Sender: %s | Msg: %s | Time: %d", sender, cert.Msg.Content, cert.Msg.Ts))
	}
	return out
}
