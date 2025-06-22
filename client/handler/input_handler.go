package handler

import (
	"bufio"
	"fmt"
	"libr/core"
	"os"
	"strings"
	"time"
)

func RunInputLoop() {
	reader := bufio.NewScanner(os.Stdin)

	fmt.Println("🔐 LIBR CLI — Type your message below:")
	for {
		fmt.Print("> ")

		if !reader.Scan() {
			fmt.Println("\n[!] Input closed.")
			break
		}

		msg := strings.TrimSpace(reader.Text())
		if msg == "exit" {
			fmt.Println("👋 Exiting LIBR client.")
			break
		}

		if !core.IsValidMessage(msg) {
			fmt.Println("[!] Invalid message. Must be non-empty, valid string.")
			continue
		}

		fmt.Println("⏳ Sending to moderators...")
		ts := time.Now().Unix()

		modcertlist := core.SendToMods(msg, ts)
		if modcertlist == nil {
			fmt.Println("Message rejected by mods")
			return
		}
		fmt.Printf("✅ Received %d accepted moderator responses.\n", len(modcertlist))

		msgCert := core.CreateMsgCert(msg, ts, modcertlist)

		response := core.SendToDb(msgCert)

		fmt.Printf("Response by DB: %v", response)
	}
}
