package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/JaineelVora08/udpholepunchtest/addressutils"
)

type PeerInfo struct {
	peerID       string
	publicAddr   string
	privateAddr  string
	connectedVia string
}

var relayPublicIP string
var relayPrivateIP string
var conn *net.UDPConn
var connectedPeer PeerInfo
var myHolePunched bool = false
var peerHolePunched bool = false
var publicPunched bool = false
var privatePunched bool = false

func main() {

	myPort := os.Args[1]

	localAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:"+myPort)
	if err != nil {
		panic(err)
	}

	conn, err = net.ListenUDP("udp", localAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	privateIP := addressutils.GetPrivateIP()
	privateAddr := privateIP + ":" + myPort
	publicAddr := addressutils.GetPublicIP()
	fmt.Println("Peer listening on:")
	fmt.Println("Private Address:", privateAddr)
	fmt.Println("Public Address:", publicAddr)

	go listenLoop()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		parts := strings.Split(input, " ")

		cmd := parts[0]

		switch cmd {

		case "register":
			// Format: register <RelayPublicIP:Port> <RelayPrivateIP:Port>
			// IMP: ie private and public address are send as a message to the relay...they are added internally here
			if len(parts) != 3 {
				fmt.Println("Usage: register <RelayPublicIP:Port> <RelayPrivateIP:Port>")
				continue
			}
			relayPublicIP = parts[1]
			relayPrivateIP = parts[2]

			privateIP := addressutils.GetPrivateIP()
			privateAddr := privateIP + ":" + myPort
			publicAddr := addressutils.GetPublicIP()

			msg := fmt.Sprintf("register %s %s", privateAddr, publicAddr)
			sendToServer(msg)

			fmt.Println("Sent register with private address: ", privateAddr, "\n and public address:", publicAddr)

		case "connectpeer":
			if len(parts) != 2 {
				fmt.Println("Usage: connectpeer <peerID>")
				continue
			}
			peerID := parts[1]
			sendToServer(fmt.Sprintf("connectpeer %s", peerID))

		case "send":
			if connectedPeer.peerID == "" {
				fmt.Println("Not connected to any peer")
				continue
			}
			message := strings.Join(parts[1:], " ")
			sendToPeer(message)

		default:
			fmt.Println("Unknown command")
		}
	}
}

func listenLoop() {
	go func() {
		buf := make([]byte, 1024)

		for {
			n, remoteAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println("Error reading:", err)
				continue
			}

			msg := string(buf[:n])
			parts := strings.Split(msg, " ")

			if parts[0] == "peerinfo" && len(parts) == 4 {
				connectedPeer = PeerInfo{
					peerID:      parts[1],
					publicAddr:  parts[2],
					privateAddr: parts[3],
				}
				fmt.Println("Received peer info:", connectedPeer)

				go punch(connectedPeer.publicAddr, "Public")
				go punch(connectedPeer.privateAddr, "Private")
				go checkHolePunching()

				continue
			}

			switch msg {
			case "punch":
				if !myHolePunched {
					fmt.Println("My NAT hole is punched (received punch from", remoteAddr.String(), ")")
					myHolePunched = true
				}
				conn.WriteToUDP([]byte("ack"), remoteAddr)

			case "ack":
				if remoteAddr.String() == connectedPeer.publicAddr {
					if !publicPunched {
						fmt.Println("Peer NAT hole punched via PUBLIC address:", remoteAddr.String())
						publicPunched = true
					}
				} else if remoteAddr.String() == connectedPeer.privateAddr {
					if !privatePunched {
						fmt.Println("Peer NAT hole punched via PRIVATE address:", remoteAddr.String())
						privatePunched = true
					}
				}

				if !peerHolePunched {
					peerHolePunched = true
				}

			default:
				fmt.Printf("Message from %s: %s\n", remoteAddr.String(), msg)
			}
		}
	}()

}

func punch(addrStr string, punchType string) {
	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		fmt.Println("Invalid address:", addrStr)
		return
	}

	fmt.Println("Attempting punch to", addrStr, "("+punchType+")")

	for i := 0; i < 3; i++ {
		conn.WriteToUDP([]byte("punch"), addr)
		time.Sleep(500 * time.Millisecond)
	}
	//fmt.Println("Sent punching packets to", addrStr, "("+punchType+")")
}

func sendToServer(msg string) {
	relayAddr1, _ := net.ResolveUDPAddr("udp", relayPublicIP)
	relayAddr2, _ := net.ResolveUDPAddr("udp", relayPrivateIP)

	conn.WriteToUDP([]byte(msg), relayAddr1)
	conn.WriteToUDP([]byte(msg), relayAddr2)
}

func sendToPeer(msg string) {
	if myHolePunched && peerHolePunched {
		peerAddr1, _ := net.ResolveUDPAddr("udp", connectedPeer.publicAddr)
		peerAddr2, _ := net.ResolveUDPAddr("udp", connectedPeer.privateAddr)

		conn.WriteToUDP([]byte(msg), peerAddr1)
		conn.WriteToUDP([]byte(msg), peerAddr2)
	} else {
		relayAddr1, _ := net.ResolveUDPAddr("udp", relayPublicIP)
		relayAddr2, _ := net.ResolveUDPAddr("udp", relayPrivateIP)

		conn.WriteToUDP([]byte("relaymsg "+connectedPeer.peerID+" "+msg), relayAddr1)
		conn.WriteToUDP([]byte("relaymsg "+connectedPeer.peerID+" "+msg), relayAddr2)
	}
}

func checkHolePunching() {
	success := make(chan bool)

	go func() {
		for {
			if myHolePunched && peerHolePunched {
				success <- true
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-success:
		if publicPunched {
			fmt.Println("Hole punching successful via PUBLIC IP")
		} else if privatePunched {
			fmt.Println("Hole punching successful via PRIVATE IP")
		} else {
			fmt.Println("Hole punching successful but route unknown")
		}
		fmt.Println("You can now chat using 'send <message>'")

	case <-time.After(10 * time.Second):
		fmt.Println("Hole punching failed. Peer unreachable directly.")
		fmt.Println("Falling back to relay for communication.")
		fmt.Println("Relay fallback activated. You can now chat using 'send <message>' via relay.")
	}
}
