package handler

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/devlup-labs/Libr/core/client/core"
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

		if strings.HasPrefix(msg, "fetch") {
			if msg == "fetch all" {
				core.FetchMsgAll()
				continue
			} else {
				ts, _ := strconv.Atoi(msg[6:])
				core.Fetch(int64(ts))
				continue
			}
		}

		if !core.IsValidMessage(msg) {
			fmt.Println("[!] Invalid message. Must be non-empty, valid string.")
			continue
		}

		fmt.Println("⏳ Sending to moderators...")
		ts := time.Now().Unix()

		modcertlist := core.SendToMods(msg, ts)
		if modcertlist == nil {
			fmt.Printf("Message rejected by mods\n")
		} else {
			fmt.Printf("✅ Received %d accepted moderator responses.\n", len(modcertlist))

			msgCert := core.CreateMsgCert(msg, ts, modcertlist)

			response := core.SendToDb(msgCert)

			fmt.Printf("Response by DB: %v\n", response)
		}
	}
}
