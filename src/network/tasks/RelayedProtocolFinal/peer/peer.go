package Node

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/multiformats/go-multiaddr"
)

const ChatProtocol = protocol.ID("/chat/1.0.0")

type ChatPeer struct {
	Host      host.Host
	relayAddr multiaddr.Multiaddr
	relayID   peer.ID
	peers     map[peer.ID]string // peer ID to nickname mapping
}

type reqFormat struct {
	Type string `json:"type"`
	Addr string
}

func NewChatPeer(relayAddr string) (*ChatPeer, error) {
	fmt.Println("[DEBUG] Parsing relay address:", relayAddr)
	relayMA, err := multiaddr.NewMultiaddr(relayAddr)
	if err != nil {
		fmt.Println("[DEBUG] Failed to parse relay multiaddr:", err)
		return nil, err
	}

	relayInfo, err := peer.AddrInfoFromP2pAddr(relayMA)
	if err != nil {
		fmt.Println("[DEBUG] Failed to extract relay peer info:", err)
		return nil, err
	}

	fmt.Println("[DEBUG] Creating connection manager")
	connMgr, err := connmgr.NewConnManager(100, 400)
	if err != nil {
		fmt.Println("[DEBUG] Failed to create connection manager:", err)
		return nil, err
	}

	fmt.Println("[DEBUG] Creating libp2p Host")
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.ConnectionManager(connMgr),
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		fmt.Println("[DEBUG] Failed to create Host:", err)
		return nil, err
	}

	fmt.Println("[DEBUG] Creating identify service")
	idSvc, err := identify.NewIDService(h)
	if err != nil {
		fmt.Println("[DEBUG] Failed to create identify service:", err)
		h.Close()
		return nil, err
	}

	getListenAddrs := func() []multiaddr.Multiaddr {
		var publicAddrs []multiaddr.Multiaddr
		for _, addr := range h.Addrs() {
			if !isPrivateAddr(addr) {
				publicAddrs = append(publicAddrs, addr)
			}
		}
		return publicAddrs
	}

	fmt.Println("[DEBUG] Creating hole punching service")
	hps, err := holepunch.NewService(h, idSvc, getListenAddrs)
	if err != nil {
		fmt.Println("[DEBUG] Failed to create hole punching service:", err)
		h.Close()
		return nil, err
	}
	_ = hps

	// Create circuit relay client
	fmt.Println("[DEBUG] Creating circuit relay client")
	// _ = client // Import for reservation function

	cp := &ChatPeer{
		Host:      h,
		relayAddr: relayMA,
		relayID:   relayInfo.ID,
		peers:     make(map[peer.ID]string),
	}

	fmt.Println("[DEBUG] Setting stream handler for chat protocol")
	h.SetStreamHandler(ChatProtocol, cp.handleChatStream)

	return cp, nil
}

func isPrivateAddr(addr multiaddr.Multiaddr) bool {
	addrStr := addr.String()
	return strings.Contains(addrStr, "127.0.0.1") ||
		strings.Contains(addrStr, "192.168.") ||
		strings.Contains(addrStr, "10.") ||
		strings.Contains(addrStr, "172.16.") ||
		strings.Contains(addrStr, "172.17.") ||
		strings.Contains(addrStr, "172.18.") ||
		strings.Contains(addrStr, "172.19.") ||
		strings.Contains(addrStr, "172.2") ||
		strings.Contains(addrStr, "172.30.") ||
		strings.Contains(addrStr, "172.31.")
}

func (cp *ChatPeer) Start(ctx context.Context) error {
	fmt.Println("[DEBUG] Connecting to relay:", cp.relayAddr)
	relayInfo, _ := peer.AddrInfoFromP2pAddr(cp.relayAddr)
	if err := cp.Host.Connect(ctx, *relayInfo); err != nil {
		fmt.Println("[DEBUG] Failed to connect to relay:", err)
		return fmt.Errorf("failed to connect to relay: %w", err)
	}

	// Make reservation with the relay
	fmt.Println("[DEBUG] Making reservation with relay...")
	reservation, err := client.Reserve(ctx, cp.Host, *relayInfo)
	if err != nil {
		fmt.Printf("[DEBUG] Failed to make reservation: %v\n", err)
		return fmt.Errorf("failed to make reservation: %w", err)
	}
	fmt.Printf("[DEBUG] Reservation successful! Expiry: %v\n", reservation.Expiration)

	fmt.Printf("[DEBUG] Peer started!\n")
	fmt.Printf("[DEBUG] Peer ID: %s\n", cp.Host.ID())

	for _, addr := range cp.Host.Addrs() {
		fmt.Printf("[DEBUG] Address: %s/p2p/%s\n", addr, cp.Host.ID())
	}

	circuitAddr := cp.relayAddr.Encapsulate(
		multiaddr.StringCast(fmt.Sprintf("/p2p-circuit/p2p/%s", cp.Host.ID())))

	listener, _ := net.Listen("tcp", "relayAddr") //signalling server address to be added
	conn, err := listener.Accept()

	if err != nil {
		fmt.Println("[DEBUG]Error connecting with signalling server")
	}

	var req reqFormat
	req.Addr = circuitAddr.String()
	req.Type = "register"

	reqJson, err := json.Marshal(req)
	if err != nil {
		fmt.Println("[DEBUG]Error converting to json signalling request")
	}
	conn.Write([]byte(reqJson))
	conn.Close()
	fmt.Printf("[INFO] Circuit Address (share this with other peers): %s\n", circuitAddr)

	// Start a goroutine to periodically refresh reservations
	go cp.refreshReservations(ctx, *relayInfo)

	return nil
}

func (cp *ChatPeer) refreshReservations(ctx context.Context, relayInfo peer.AddrInfo) {
	ticker := time.NewTicker(5 * time.Minute) // Refresh every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("[DEBUG] Refreshing relay reservation...")
			if reservation, err := client.Reserve(ctx, cp.Host, relayInfo); err != nil {
				fmt.Printf("[DEBUG] Failed to refresh reservation: %v\n", err)
			} else {
				fmt.Printf("[DEBUG] Reservation refreshed! Expiry: %v\n", reservation.Expiration)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (cp *ChatPeer) handleChatStream(s network.Stream) {
	fmt.Println("[DEBUG] Incoming chat stream from", s.Conn().RemotePeer())
	defer s.Close()

	remotePeer := s.Conn().RemotePeer()
	nickname := cp.peers[remotePeer]
	if nickname == "" {
		nickname = remotePeer.String()[:8] + "..."
	}

	reader := bufio.NewReader(s)
	for {
		message, err := reader.ReadString('\n')

		if err != nil {
			// io.EOF means the end of the input stream was reached (e.g., connection closed gracefully).
			if err == io.EOF {
				fmt.Printf("[INFO] Reader finished or connection closed for %s.\n", nickname)
			} else {
				// Other errors (e.g., network issues)
				fmt.Printf("[ERROR] Failed to read message from %s: %v\n", nickname, err)
			}
			return
		}

		message = strings.TrimSpace(message)

		if message == "" {
			//fmt.Printf("[DEBUG] Skipped empty message from %s.\n", nickname)
			continue
		}

		fmt.Printf("[%s] Received Raw: %s\n", nickname, message)

		// 5. Unmarshal the JSON string into a Go data structure.
		//    We use map[string]interface{} for flexibility, allowing any valid JSON object.
		var jsonData map[string]interface{}
		err = json.Unmarshal([]byte(message), &jsonData) // Convert string to []byte for unmarshaling

		if err != nil {
			fmt.Printf("[ERROR] Failed to unmarshal JSON for %s: %v. Message: \"%s\"\n", nickname, err, message)

			continue
		}

		//fmt.Printf("[%s] Successfully Unmarshaled JSON: %+v\n", nickname, jsonData)

		if jsonData["Method"] == "GET" {
			s.Write([]byte("Serving your GET request"))
			//actual logic to be added
		}

	}
}

func (cp *ChatPeer) ConnectToPeer(ctx context.Context, TargetIP string, targetPort string, nickname string) (peer.ID, error) {
	completeIP := TargetIP + ":" + targetPort
	var req reqFormat
	req.Addr = completeIP
	req.Type = "getID"
	listener, err := net.Listen("tcp", "signallingServerIP")
	conn, err := listener.Accept()
	jsonReq, _ := json.Marshal(req)
	_, err = conn.Write([]byte(jsonReq))
	buf := new([]byte)
	conn.Read(*buf)
	targetPeerID := string(*buf)
	peerAddr := fmt.Sprintf("/%s/<RELAY_IP>/tcp/<RELAY_PORT>/p2p/<RELAY_PEER_ID>/p2p-circuit/p2p/%s", completeIP, targetPeerID)

	fmt.Printf("[DEBUG] Parsing peer address: %s\n", peerAddr)
	peerMA, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		fmt.Println("[DEBUG] Failed to parse peer multiaddr:", err)
		return "", err
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(peerMA)
	if err != nil {
		fmt.Println("[DEBUG] Failed to extract peer info:", err)
		return "", err
	}

	cp.peers[peerInfo.ID] = nickname

	fmt.Printf("[DEBUG] Connecting to peer %s (%s)...\n", nickname, peerInfo.ID)

	// Create a context with timeout for connection
	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := cp.Host.Connect(connectCtx, *peerInfo); err != nil {
		fmt.Println("[DEBUG] Failed to connect to peer:", err)
		return "", fmt.Errorf("failed to connect to peer: %w", err)
	}

	fmt.Printf("[DEBUG] Connected to %s!\n", nickname)

	conns := cp.Host.Network().ConnsToPeer(peerInfo.ID)
	if len(conns) > 0 {
		fmt.Printf("[DEBUG] Connection type: %s\n", conns[0].RemoteMultiaddr())
	} else {
		fmt.Println("[DEBUG] No connection found after connect")
	}

	return peerInfo.ID, nil
}

func (cp *ChatPeer) SendMessage(peerID peer.ID, message string) error {
	fmt.Printf("[DEBUG] Sending message to %s: %s\n", peerID, message)
	stream, err := cp.Host.NewStream(context.Background(), peerID, ChatProtocol)
	if err != nil {
		fmt.Println("[DEBUG] Failed to open stream:", err)
		return err
	}
	defer stream.Close()

	_, err = stream.Write([]byte(message + "\n"))
	if err != nil {
		fmt.Println("[DEBUG] Failed to write to stream:", err)
	}
	return err
}

func (cp *ChatPeer) GetConnectedPeers() []peer.ID {
	var peers []peer.ID
	for _, conn := range cp.Host.Network().Conns() {
		remotePeer := conn.RemotePeer()
		if remotePeer != cp.relayID {
			peers = append(peers, remotePeer)
		}
	}
	fmt.Printf("[DEBUG] Connected peers: %v\n", peers)
	return peers
}

func (cp *ChatPeer) Close() error {
	fmt.Println("[DEBUG] Closing Host")
	return cp.Host.Close()
}

// func main() {
// 	if len(os.Args) < 2 {
// 		fmt.Println("Usage: go run peer/main.go <relay_address>")
// 		fmt.Println("Example: go run peer/main.go /ip4/127.0.0.1/tcp/12345/p2p/12D3KooW...")
// 		os.Exit(1)
// 	}

// 	relayAddr := os.Args[1]

// 	ctx := context.Background()

// 	fmt.Println("[DEBUG] Creating ChatPeer")
// 	Peer, err := NewChatPeer(relayAddr)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer Peer.Close()

// 	fmt.Println("[DEBUG] Starting ChatPeer")
// 	if err := Peer.Start(ctx); err != nil {
// 		log.Fatal(err)
// 	}

// 	scanner := bufio.NewScanner(os.Stdin)
// 	fmt.Println("\nCommands:")
// 	fmt.Println("  connect <peer_circuit_address> <nickname> - Connect to a peer")
// 	fmt.Println("  msg <message> - Send message to all connected peers")
// 	fmt.Println("  peers - List connected peers")
// 	fmt.Println("  quit - Exit")
// 	fmt.Println("\nNote: Use the circuit address printed by the other peer to connect")

// 	for {
// 		fmt.Print("> ")
// 		if !scanner.Scan() {
// 			break
// 		}

// 		input := strings.TrimSpace(scanner.Text())
// 		parts := strings.SplitN(input, " ", 3)

// 		switch parts[0] {
// 		case "connect":
// 			if len(parts) < 3 {
// 				fmt.Println("Usage: connect <peer_circuit_address> <nickname>")
// 				continue
// 			}
// 			fmt.Printf("[DEBUG] Command: connect %s %s\n", parts[1], parts[2])
// 			if err := Peer.ConnectToPeer(ctx, parts[1], parts[2]); err != nil {
// 				fmt.Printf("Failed to connect: %v\n", err)
// 			}

// 		case "msg":
// 			if len(parts) < 2 {
// 				fmt.Println("Usage: msg <message>")
// 				continue
// 			}
// 			message := strings.Join(parts[1:], " ")
// 			fmt.Printf("[DEBUG] Command: msg %s\n", message)
// 			connectedPeers := Peer.GetConnectedPeers()

// 			if len(connectedPeers) == 0 {
// 				fmt.Println("No peers connected")
// 				continue
// 			}

// 			for _, peerID := range connectedPeers {
// 				if err := Peer.SendMessage(peerID, message); err != nil {
// 					fmt.Printf("Failed to send to %s: %v\n", peerID, err)
// 				}
// 			}

// 		case "peers":
// 			fmt.Println("[DEBUG] Command: peers")
// 			connectedPeers := Peer.GetConnectedPeers()
// 			if len(connectedPeers) == 0 {
// 				fmt.Println("No peers connected")
// 			} else {
// 				fmt.Printf("Connected peers (%d):\n", len(connectedPeers))
// 				for _, peerID := range connectedPeers {
// 					nickname := Peer.peers[peerID]
// 					if nickname == "" {
// 						nickname = "Unknown"
// 					}
// 					conns := Peer.Host.Network().ConnsToPeer(peerID)
// 					connType := "direct"
// 					if len(conns) > 0 && strings.Contains(conns[0].RemoteMultiaddr().String(), "p2p-circuit") {
// 						connType = "relayed"
// 					}
// 					fmt.Printf("  %s (%s) - %s connection\n", nickname, peerID, connType)
// 				}
// 			}

// 		case "quit":
// 			fmt.Println("[DEBUG] Command: quit")
// 			return

// 		default:
// 			fmt.Println("Unknown command. Available: connect, msg, peers, quit")
// 		}
// 	}
// }
