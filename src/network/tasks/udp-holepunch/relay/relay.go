package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/JaineelVora08/udpholepunchtest/addressutils"
)

type PeerInfo struct {
	peerID       string
	publicAddr   string
	privateAddr  string
	connectedVia string
}

var peers = make(map[string]PeerInfo)
var mu sync.Mutex

func main() {
	addr := net.UDPAddr{
		Port: 8000,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	privateIP := addressutils.GetPrivateIP()
	publicIP := addressutils.GetPublicIP()

	fmt.Println("Server listening on:")
	fmt.Println("Private IP:", privateIP+":"+strconv.Itoa(addr.Port))
	fmt.Println("Public IP:", publicIP)

	buf := make([]byte, 1024)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading:", err)
			continue
		}

		data := make([]byte, n)
		copy(data, buf[:n])

		go handleMessage(conn, remoteAddr, data)
	}
}

func handleMessage(conn *net.UDPConn, remoteAddr *net.UDPAddr, data []byte) {
	message := string(data)
	fmt.Printf("Received '%s' from %s\n", message, remoteAddr.String())

	parts := strings.SplitN(message, " ", 3)

	cmd := parts[0]

	switch cmd {
	case "register":
		if len(parts) != 3 {
			conn.WriteToUDP([]byte("Usage: register <privateAddr> <publicAddr>"), remoteAddr)
			return
		}

		privateAddr := parts[1]
		publicAddr := parts[2]
		peerID := PeerIDGenerate()

		mu.Lock()
		peers[peerID] = PeerInfo{
			peerID:       peerID,
			publicAddr:   publicAddr,
			privateAddr:  privateAddr,
			connectedVia: remoteAddr.String(),
		}
		mu.Unlock()

		conn.WriteToUDP([]byte(fmt.Sprintf("You have been registered. Your peerID is: %s", peerID)), remoteAddr)
		fmt.Printf("Registered peer %s: Public %s, Private %s, ConnectedVia %s\n", peerID, publicAddr, privateAddr, remoteAddr.String())

	case "connectpeer":
		if len(parts) != 2 {
			conn.WriteToUDP([]byte("Usage: connectpeer <peerID>"), remoteAddr)
			return
		}
		peerID := parts[1]

		mu.Lock()
		peerB, existsB := peers[peerID]

		var myPeerID string
		var myPeer PeerInfo
		for id, p := range peers {
			if p.publicAddr == remoteAddr.String() || p.privateAddr == remoteAddr.String() {
				myPeerID = id
				myPeer = p
				break
			}
		}
		mu.Unlock()

		if !existsB {
			conn.WriteToUDP([]byte("peer not found"), remoteAddr)
			return
		}
		if myPeerID == "" {
			conn.WriteToUDP([]byte("you are not registered"), remoteAddr)
			return
		}

		respToA := fmt.Sprintf("peerinfo %s %s %s", peerB.peerID, peerB.publicAddr, peerB.privateAddr)
		conn.WriteToUDP([]byte(respToA), remoteAddr)

		respToB := fmt.Sprintf("peerinfo %s %s %s", myPeer.peerID, myPeer.publicAddr, myPeer.privateAddr)
		targetAddrPublic, err1 := net.ResolveUDPAddr("udp", peerB.publicAddr)
		if err1 == nil {
			conn.WriteToUDP([]byte(respToB), targetAddrPublic)
		}
		targetAddrPrivate, err2 := net.ResolveUDPAddr("udp", peerB.privateAddr)
		if err2 == nil {
			conn.WriteToUDP([]byte(respToB), targetAddrPrivate)
		}

		fmt.Printf("Exchanged info between %s and %s\n", myPeerID, peerID)

	case "relaymsg":
		if len(parts) < 3 {
			conn.WriteToUDP([]byte("Usage: relaymsg <peerID> <message>"), remoteAddr)
			return
		}

		targetID := parts[1]
		msgToSend := strings.Join(parts[2:], " ")

		mu.Lock()
		targetPeer, exists := peers[targetID]
		mu.Unlock()

		if !exists {
			conn.WriteToUDP([]byte("Target peer not found"), remoteAddr)
			return
		}

		targetAddr, err := net.ResolveUDPAddr("udp", targetPeer.connectedVia)
		if err != nil {
			conn.WriteToUDP([]byte("Invalid target peer address"), remoteAddr)
			return
		}

		conn.WriteToUDP([]byte("[Relayed] "+msgToSend), targetAddr)
		fmt.Printf("Relayed message from %s to %s\n", remoteAddr.String(), targetAddr.String())

	default:
		conn.WriteToUDP([]byte("Unknown command"), remoteAddr)
	}
}

func PeerIDGenerate() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := make([]byte, 6)
	for i := range id {
		id[i] = charset[rand.Intn(len(charset))]
	}
	return string(id)
}
