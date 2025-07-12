package handler

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/devlup-labs/Libr/core/client/core"
	util "github.com/devlup-labs/Libr/core/client/util"
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
			fields := strings.Fields(msg)
			if len(fields) == 2 {
				if fields[1] == "all" {
					core.FetchRecent()
				} else {
					ts, err := strconv.ParseInt(fields[1], 10, 64)
					if err != nil {
						fmt.Println("[!] Invalid timestamp format.")
					} else {
						core.Fetch(ts)
					}
				}
			} else {
				fmt.Println("[!] Usage: fetch <timestamp> or fetch all")
			}
			continue // ✅ this prevents further processing
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

			key := util.GenerateNodeID(strconv.FormatInt(msgCert.Msg.Ts, 10))
			response := core.SendToDb(key, msgCert)

			fmt.Printf("Response by DB: %v\n", response)
		}
	}
}
