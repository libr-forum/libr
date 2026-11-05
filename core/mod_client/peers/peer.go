package peer

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"log"
	"math/big"
	"sort"

	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	//"io"

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
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/libr-forum/Libr/core/mod_client/logger"
	"github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"

	//webtransport "github.com/libp2p/go-libp2p/p2p/transport/webtransport"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
)

const ChatProtocol = protocol.ID("/chat/1.0.0")

var OwnPubIP string

type ChatPeer struct {
	Host           host.Host
	relayAddr      multiaddr.Multiaddr
	relayID        peer.ID
	peers          map[peer.ID]string // peer ID to nickname mapping
	OnStatusChange func(string)
}

type reqFormat struct {
	Type      string          `json:"type,omitempty"`
	PeerID    string          `json:"peer_id,omitempty"`
	ReqParams json.RawMessage `json:"reqparams,omitempty"`
	Body      json.RawMessage `json:"body,omitempty"`
}

// type RelayDist struct {
// 	relayID string
// 	dist    *big.Int
// }

func NewChatPeer(relayMultiAddrList []string) (*ChatPeer, error) {

	var relayList []string
	for _, multiaddr := range relayMultiAddrList {
		parts := strings.Split(multiaddr, "/")
		relayList = append(relayList, parts[len(parts)-1])
	}

	caCertPool := x509.NewCertPool()

	fmt.Println("[DEBUG] Creating connection manager")
	connMgr, err := connmgr.NewConnManager(100, 400)
	if err != nil {
		fmt.Println("[DEBUG] Failed to create connection manager:", err)
		return nil, err
	}

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
		// Other TLS configurations like ClientAuth, InsecureSkipVerify, etc.
	}

	fmt.Println("[DEBUG] Creating libp2p Host")
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0/ws"), // WebSocket
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.ConnectionManager(connMgr),
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(websocket.New, websocket.WithTLSConfig(tlsConfig)),
		// libp2p.Transport(websocket.NewWithTLSConfig(tlsConfig)),
		// libp2p.Transport(websocket.New),
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

	var distmap []RelayDist
	//OwnPubIP = GetPublicIP()
	h1 := sha256.New()
	h1.Write([]byte(h.ID().String()))
	peerIDhash := hex.EncodeToString(h1.Sum(nil))

	for _, relay := range relayList {

		h_R := sha256.New()
		h_R.Write([]byte(relay))
		RelayIDhash := hex.EncodeToString(h_R.Sum(nil))

		dist := XorHexToBigInt(peerIDhash, RelayIDhash)

		distmap = append(distmap, RelayDist{dist: dist, relayID: relay})
	}

	sort.Slice(distmap, func(i, j int) bool {
		return distmap[i].dist.Cmp(distmap[j].dist) < 0
	})

	relayIDused := distmap[0].relayID
	fmt.Println(relayIDused)
	var relayAddr string

	for _, multiaddr := range relayMultiAddrList {
		parts := strings.Split(multiaddr, "/")
		if parts[len(parts)-1] == relayIDused {
			relayAddr = multiaddr
			break
		}
	}

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
	// Create circuit relay client
	fmt.Println("[DEBUG] Creating circuit relay client")
	// _ = client // Import for reservation function

	cp := &ChatPeer{
		Host:      h,
		relayAddr: relayMA,
		relayID:   relayInfo.ID,
		peers:     make(map[peer.ID]string),
	}
	// Protect relay connection from connection manager pruning
	h.ConnManager().Protect(relayInfo.ID, "relay")

	fmt.Println(h.ID().String())

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

// why????

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

	fmt.Printf("[INFO] Circuit Address (share this with other peers): %s\n", circuitAddr)

	// Protect connection to relay from connection manager
	cp.Host.ConnManager().Protect(relayInfo.ID, "relay")

	// Start a goroutine to periodically refresh reservations
	go cp.refreshReservations(ctx, *relayInfo)
	go cp.keepAlive(ctx, *relayInfo)

	// Register with relay with immediate retries
	for attempt := 1; attempt <= 3; attempt++ {
	err = cp.registerWithRelay(relayInfo.ID)
	if err == nil {
		break
	}
	fmt.Printf("[DEBUG] Register with relay attempt %d failed: %v\n", attempt, err)
	time.Sleep(1 * time.Second)
	}

	if cp.OnStatusChange != nil {
		cp.OnStatusChange("online")
	}

	return nil
}

func (cp *ChatPeer) registerWithRelay(relayID peer.ID) error {
	var reqSent reqFormat
	reqSent.Type = "register"
	reqSent.PeerID = cp.Host.ID().String()

	fmt.Println(reqSent.PeerID)
	logger.LogToFile("PeerID: " + reqSent.PeerID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := cp.Host.NewStream(ctx, relayID, ChatProtocol)
	if err != nil {
		return fmt.Errorf("error opening register stream: %w", err)
	}
	defer stream.Close()

	reqJson, err := json.Marshal(reqSent)
	if err != nil {
		return fmt.Errorf("error marshalling register req: %w", err)
	}
	_, err = stream.Write([]byte(reqJson))
	if err != nil {
		return fmt.Errorf("error writing to register stream: %w", err)
	}
	// Crucial: signal EOF immediately so the relay's io.ReadAll receives it without delay
	_ = stream.CloseWrite()

	fmt.Println("[DEBUG] Registered peer ID with relay successfully:", reqSent.PeerID)
	return nil
}

func (cp *ChatPeer) refreshReservations(ctx context.Context, relayInfo peer.AddrInfo) {
	ticker := time.NewTicker(4 * time.Minute) // Refresh every 4 minutes (slightly faster than 5m)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("[DEBUG] Refreshing relay reservation...")
			// Ensuring we are registered just in case. It's idempotent and cheap.
			// It handles the edge case where an implicit reconnection didn't trigger registration.
			cp.registerWithRelay(relayInfo.ID)

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

func (cp *ChatPeer) keepAlive(ctx context.Context, relayInfo peer.AddrInfo) {
	// Fast initial check after 2 seconds to ensure registration is cemented
	select {
	case <-time.After(2 * time.Second):
		if cp.Host.Network().Connectedness(relayInfo.ID) != network.Connected {
			fmt.Println("[DEBUG] Initial relay connection dropped! Reconnecting...")
			_ = cp.Host.Connect(ctx, relayInfo)
			_ = cp.registerWithRelay(relayInfo.ID)
			_, _ = client.Reserve(ctx, cp.Host, relayInfo)
		} else {
			_ = cp.registerWithRelay(relayInfo.ID)
		}
		if cp.OnStatusChange != nil {
			cp.OnStatusChange("online")
		}
	case <-ctx.Done():
		return
	}

	ticker := time.NewTicker(20 * time.Second) // Ping every 20 seconds to prevent idle timeout
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wasConnected := cp.Host.Network().Connectedness(relayInfo.ID) == network.Connected

			pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			ch := ping.Ping(pingCtx, cp.Host, relayInfo.ID)
			res := <-ch
			cancel()
			if res.Error != nil {
				fmt.Println("[DEBUG] Relay ping failed:", res.Error)
				// Close the existing, broken connection aggressively to force a new dial
				cp.Host.Network().ClosePeer(relayInfo.ID)

				for i := 0; i < 3; i++ {
					connCtx, connCancel := context.WithTimeout(ctx, 10*time.Second)
					err := cp.Host.Connect(connCtx, relayInfo)
					connCancel()
					if err == nil {
						fmt.Println("[DEBUG] Reconnected transport to relay!")
						cp.registerWithRelay(relayInfo.ID)
						client.Reserve(ctx, cp.Host, relayInfo)
						if cp.OnStatusChange != nil {
							cp.OnStatusChange("online")
						}
						break
					}
					fmt.Println("[DEBUG] Failed to reconnect to relay, retrying in 2s...")
					if i == 2 && cp.OnStatusChange != nil {
						cp.OnStatusChange("offline")
					}
					time.Sleep(2 * time.Second)
				}
			} else {
				// Ping succeeded! But what if the peer wasn't connected at the beginning of this tick?
				// This means ping.Ping() automatically dialed and implicitly reconnected!
				if !wasConnected {
					fmt.Println("[DEBUG] Implicitly reconnected transport to relay!")
					cp.registerWithRelay(relayInfo.ID)
					client.Reserve(ctx, cp.Host, relayInfo)
					if cp.OnStatusChange != nil {
						cp.OnStatusChange("online")
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (cp *ChatPeer) handleChatStream(s network.Stream) {
	fmt.Println("[DEBUG] Incoming chat stream from", s.Conn().RemotePeer())
	defer s.Close()

	reader := bufio.NewReader(s)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		fmt.Println("[DEBUG]Error reading the bytes from the stream:", err)
		return
	}
	line = bytes.TrimRight(line, "\n")
	line = bytes.TrimRight(line, "\x00")

	fmt.Println("[DEBUG] Raw input:", string(line))

	var reqStruct reqFormat
	err = json.Unmarshal(line, &reqStruct)
	if err != nil {
		fmt.Println("[DEBUG]Error unmarshalling to reqStruc:", err)
		return
	}

	var reqData map[string]interface{}
	reqStruct.ReqParams = bytes.TrimRight(reqStruct.ReqParams, "\x00")

	if err := json.Unmarshal(reqStruct.ReqParams, &reqData); err != nil {
		fmt.Printf("[ERROR] Failed to unmarshal incoming request: %v\n", err)
		return
	}

	fmt.Printf("[DEBUG]ReqData is : %+v \n", reqData)

	if reqData["Method"] == "GET" {
		resp := ServeGetReq(reqStruct.ReqParams)
		resp = bytes.TrimRight(resp, "\x00")
		_, err = s.Write(resp)
		if err != nil {
			fmt.Println("[DEBUG]Error writing resp bytes to relay")
			return
		}
	}

	if reqData["Method"] == "POST" {
		resp := ServePostReq(reqStruct.PeerID, reqStruct.ReqParams, reqStruct.Body) // have to set the new logic in serve post req now
		resp = bytes.TrimRight(resp, "\x00")
		fmt.Println("Response from ServePostReq:", string(resp))
		_, err = s.Write(resp)
		if err != nil {
			fmt.Println("[DEBUG]Error writing resp bytes to relay")
			return
		}
	}
}

func (cp *ChatPeer) Send(ctx context.Context, targetPeerID string, jsonReq []byte, body []byte) ([]byte, error) {
	if cp == nil {
		return nil, fmt.Errorf("chat peer is nil")
	}
	if cp.Host == nil {
		return nil, fmt.Errorf("chat peer host is nil")
	}

	//completeIP := TargetIP + ":" + targetPort

	var req reqFormat
	req.Type = "SendMsg"
	//req.PeerID = completeIP
	req.PeerID = targetPeerID
	req.ReqParams = jsonReq
	req.Body = body
	stream, err := cp.Host.NewStream(ctx, cp.relayID, ChatProtocol)
	if err != nil {
		fmt.Println("[DEBUG]Error opneing a fetch ID stream to relay")
		return nil, err
	}

	jsonReqRelay, err := json.Marshal(req)

	if err != nil {
		fmt.Println("[DEBUG]Error marshalling get req to be sent to relay")
		return nil, err
	}

	stream.Write([]byte(jsonReqRelay))

	fmt.Println("[DEBUG]Msg req sent to relay, waiting for ack")

	// Signal that we've finished writing the request
	// (lets the relay/mod detect EOF on their reader)
	stream.CloseWrite()

	fmt.Println("[DEBUG]Msg req sent to relay, waiting for ack")

	// Use json.NewDecoder so it reads exactly one JSON object
	// and doesn't wait for EOF which hangs the stream.
	var rawMsg json.RawMessage
	err = json.NewDecoder(stream).Decode(&rawMsg)
	if err != nil {
		fmt.Println("[DEBUG] Error reading JSON response from relay:", err)
		defer stream.Close()
		return nil, err
	}

	defer stream.Close()

	return []byte(rawMsg), nil
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

// func GetPublicIP() string {
// 	c, err := stun.Dial("udp4", "stun.l.google.com:19302")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer c.Close()

// 	var xorAddr stun.XORMappedAddress
// 	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
// 	if err := c.Do(message, func(res stun.Event) {
// 		if res.Error != nil {
// 			panic(res.Error)
// 		}
// 		if err := xorAddr.GetFrom(res.Message); err != nil {
// 			panic(err)
// 		}
// 	}); err != nil {
// 		panic(err)
// 	}

// 	if xorAddr.IP.To4() == nil {
// 		panic("STUN returned an IPv6 address; IPv4 not available")
// 	}

// 	peerAd := xorAddr.IP.String() + ":" + strconv.Itoa(xorAddr.Port)
// 	return peerAd
// }

func XorHexToBigInt(hex1, hex2 string) *big.Int {

	bytes1, err1 := hex.DecodeString(hex1)
	bytes2, err2 := hex.DecodeString(hex2)

	if err1 != nil || err2 != nil {
		log.Fatalf("Error decoding hex: %v %v", err1, err2)
	}

	if len(bytes1) != len(bytes2) {
		log.Fatalf("Hex strings must be the same length")
	}

	xorBytes := make([]byte, len(bytes1))
	for i := 0; i < len(bytes1); i++ {
		xorBytes[i] = bytes1[i] ^ bytes2[i]
	}

	result := new(big.Int).SetBytes(xorBytes)
	return result
}
