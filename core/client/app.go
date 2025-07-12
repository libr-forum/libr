package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/devlup-labs/Libr/core/client/core"
	"github.com/devlup-labs/Libr/core/client/keycache"
	util "github.com/devlup-labs/Libr/core/client/util"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	keycache.InitKeys()
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SendInput(input string) string {
	fmt.Println("Received input:", input)

	ts := time.Now().Unix()
	modcertlist := core.SendToMods(input, ts)

	if len(modcertlist) == 0 {
		fmt.Println("Rejected by mods")
		return "❌ Message rejected by moderators."
	}

	fmt.Printf("Approved by %d mods\n", len(modcertlist))
	msgCert := core.CreateMsgCert(input, ts, modcertlist)

	key := util.GenerateNodeID(strconv.FormatInt(msgCert.Msg.Ts, 10))
	response := core.SendToDb(key, msgCert)

	fmt.Println("Sent to DB, response:", response)
	return fmt.Sprintf("✅ Sent to DB. Time: %d", msgCert.Msg.Ts)
}

func (a *App) FetchAll() []string {
	messages := core.FetchRecent()
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
