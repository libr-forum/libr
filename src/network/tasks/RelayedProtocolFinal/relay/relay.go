package main

import (
	"bufio"
	"bytes"
	//Peers "chatprotocol/peer"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	relay "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	ma "github.com/multiformats/go-multiaddr"
)

const ChatProtocol = protocol.ID("/chat/1.0.0")


type reqFormat struct {
	Type  string `json:"type,omitempty"`
	PubIP string `json:"pubip,omitempty"`
	ReqParams json.RawMessage `json:"reqparams,omitempty"`
	Body json.RawMessage `json:"body,omitempty"`
}

var (
	IDmap = make(map[string]string)
	mu    sync.RWMutex
)

var RelayHost host.Host

type respFormat struct {
	Type string `json:"type"`
	Resp []byte `json:"resp"`
}

type RelayEvents struct{}

func (re *RelayEvents) Listen(net network.Network, addr ma.Multiaddr) {}
func (re *RelayEvents) ListenClose(net network.Network, addr ma.Multiaddr) {}
func (re *RelayEvents) Connected(net network.Network, conn network.Conn) {
	fmt.Printf("[INFO] Peer connected: %s\n", conn.RemotePeer())
}
func (re *RelayEvents) Disconnected(net network.Network, conn network.Conn) {
	fmt.Printf("[INFO] Peer disconnected: %s\n", conn.RemotePeer())
	// Remove peer from IDmap if needed
	mu.Lock()
	for pubip, pid := range IDmap {
		if pid == conn.RemotePeer().String() {
			delete(IDmap, pubip)
			break
		}
	}
	mu.Unlock()
}


func main() {
	fmt.Println("[DEBUG] Starting relay node...")

	// Create connection manager
	fmt.Println("[DEBUG] Creating connection manager...")
	connMgr, err := connmgr.NewConnManager(100, 400)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create connection manager: %v", err)
	}

	// Create the relay host
	fmt.Println("[DEBUG] Creating relay host...")
	RelayHost, err = libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/4567"),
		libp2p.ConnectionManager(connMgr),
		libp2p.EnableNATService(),
		libp2p.EnableRelayService(),
	)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create relay host: %v", err)
	}
	RelayHost.Network().Notify(&RelayEvents{})

	defer func() {
		fmt.Println("[DEBUG] Closing relay host...")
		RelayHost.Close()
	}()

	// Enable circuit relay service
	fmt.Println("[DEBUG] Enabling circuit relay service...")
	_, err = relay.New(RelayHost)
	if err != nil {
		log.Fatalf("[ERROR] Failed to enable relay service: %v", err)
	}

	fmt.Printf("[INFO] Relay started!\n")
	fmt.Printf("[INFO] Peer ID: %s\n", RelayHost.ID())

	// Print all addresses
	for _, addr := range RelayHost.Addrs() {
		fmt.Printf("[INFO] Relay Address: %s/p2p/%s\n", addr, RelayHost.ID())
	}

	RelayHost.SetStreamHandler("/chat/1.0.0", handleChatStream)
	go func() {
		for {
			fmt.Println(IDmap)
			time.Sleep(30 * time.Second)
		}
	}()

	// Wait for interrupt signal
	fmt.Println("[DEBUG] Waiting for interrupt signal...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("[INFO] Shutting down relay...")
}

func handleChatStream(s network.Stream) {
	fmt.Println("[DEBUG] Incoming chat stream from", s.Conn().RemoteMultiaddr())
	defer s.Close()
	reader := bufio.NewReader(s)
	for {

		var req reqFormat
		buf := make([]byte, 1024*4) // or size based on expected message
		n, err := reader.Read(buf)
		if err != nil {
			fmt.Println("[DEBUG] Error reading from connection at relay:", err)
			return
		}
		buf = bytes.TrimRight(buf, "\x00")
		
		err = json.Unmarshal(buf[:n], &req)
		if err != nil {
			fmt.Printf("[DEBUG] Error parsing JSON at relay: %v\n", err)
			fmt.Printf("[DEBUG] Received Data: %s\n", string(buf[:n]))
			return
		}

		fmt.Printf("req by user is : %+v \n", req)

		if req.Type == "register" {
			peerID := s.Conn().RemotePeer()
			fmt.Printf("[INFO]Given public IP is %s \n", req.PubIP)
			fmt.Println("[INFO]Registering the peer into relay map")
			mu.Lock()
			IDmap[req.PubIP] = peerID.String()
			mu.Unlock()
		}

		if req.Type == "SendMsg" {
			mu.RLock()
			targetPeerID := IDmap[req.PubIP]
			mu.RUnlock()
			fmt.Println(targetPeerID)

			relayID := RelayHost.ID()
			targetID, err := peer.Decode(targetPeerID)
			if err != nil {
				log.Printf("[ERROR] Invalid Peer ID: %v", err)
				s.Write([]byte("invalid peer id"))
				return
			}

			relayBaseAddr, err := ma.NewMultiaddr("/p2p/" + relayID.String())
			if err != nil {
				log.Fatal("relayBaseAddr error:", err)
			}
			circuitAddr, _ := ma.NewMultiaddr("/p2p-circuit")
			targetAddr, _ := ma.NewMultiaddr("/p2p/" + targetID.String())
			fullAddr := relayBaseAddr.Encapsulate(circuitAddr).Encapsulate(targetAddr)
			fmt.Println("[DENUG]", fullAddr.String())
			addrInfo, err := peer.AddrInfoFromP2pAddr(fullAddr)
			if err != nil {
				log.Printf("Invalid relayed multiaddr: %s", fullAddr)
				s.Write([]byte("bad relayed addr"))
				return
			}

			// Add the relayed address to the peerstore. PeerStore is a mapping in which peer ID is mapped to multiaddr for that peer. This is used whenever we want to open a stream. Once added then we should connect to the peer and open a stream to send message to the relay
			RelayHost.Peerstore().AddAddrs(addrInfo.ID, addrInfo.Addrs, peerstore.PermanentAddrTTL)

			err = RelayHost.Connect(context.Background(), *addrInfo)
			if err != nil {
				log.Printf("[ERROR] Failed to connect to relayed peer: %v", err)
			}

			sendStream, err := RelayHost.NewStream(context.Background(), targetID, ChatProtocol)
			if err != nil {
				fmt.Println("[DEBUG]Error opening stream to target peer")
				fmt.Println(err)
				s.Write([]byte("failed"))
				return
			}
			jsonReqServer, err := json.Marshal(req)
			if err != nil {
				fmt.Println("[DEBUG]Error marshalling the req for server ")
			}
			 _, err = sendStream.Write(append(jsonReqServer, '\n'))

			if err != nil {
				fmt.Println("[DEBUG]Error sending messgae despite stream opened")
				return
			}
			s.Write([]byte("Success\n"))

			buf := make([]byte, 1024)
			RespReader := bufio.NewReader(sendStream)
			RespReader.Read(buf)
			buf = bytes.TrimRight(buf, "\x00")
			var resp respFormat
			resp.Type = "GET"
			resp.Resp = buf
			fmt.Printf("[Debug]Resp from %s : %+v \n", targetID.String(), resp)

			jsonResp, err := json.Marshal(resp)
			if err != nil {
				fmt.Println("[DEBUG]Error marshalling the response at relay")
			}
			_=jsonResp // if required whole jsonResp can be sent but it makes unmarhsalling the response harder for the client
			fmt.Println("[DEBUG]Raw Resp :", string(resp.Resp))
			_,err = s.Write(resp.Resp)
			if(err!=nil){
				fmt.Println("[DEBUG]Error sending response back")
			}
			defer s.Close()
			defer sendStream.Close()
		}

	}

}
