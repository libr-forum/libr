package cmd

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var listenCmd = &cobra.Command{
	Use:   "listen [your_port]",
	Short: "Start listening for UDP messages and send from same terminal",
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		myPort := args[0]

		myAdd, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+myPort)
		if err != nil {
			fmt.Println("Error resolving local address:", err)
			return
		}

		conn, err := net.ListenUDP("udp", myAdd)
		if err != nil {
			fmt.Println("Error starting listener:", err)
			return
		}
		defer conn.Close()

		fmt.Println("Listening on port", myPort)
		fmt.Println("For sending message: send <target_ip:port> <message>")

		// Goroutine to receive messages
		go func() {
			buffer := make([]byte, 1024)
			for {
				n, addr, err := conn.ReadFromUDP(buffer)
				if err != nil {
					fmt.Println("Read error:", err)
					continue
				}
				fmt.Printf("\nMessage from %s: %s\n", addr.String(), string(buffer[:n]))
				fmt.Print("> ")
			}
		}()

		// Sending messages loop
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			input, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			parts := strings.SplitN(input, " ", 3)
			if len(parts) == 3 && parts[0] == "send" {
				recadd := parts[1]
				msg := parts[2]
				receiverAdd, err := net.ResolveUDPAddr("udp", recadd)
				if err != nil {
					fmt.Println("Error resolving target address:", err)
					continue
				}

				_, err = conn.WriteToUDP([]byte(msg), receiverAdd)
				if err != nil {
					fmt.Println("Send error:", err)
				}
			} else {
				fmt.Println("Unknown command. Use: send <target_ip:port> <message>")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listenCmd)
}
