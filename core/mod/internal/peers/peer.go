package Peers

import (
	"bufio"
	"bytes"

	"context"
	"encoding/json"
	"fmt"

	//"io"
	"strconv"
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
	"github.com/pion/stun"
)

const ChatProtocol = protocol.ID("/chat/1.0.0")

type ChatPeer struct {
	Host      host.Host
	relayAddr multiaddr.Multiaddr
	relayID   peer.ID
	peers     map[peer.ID]string // peer ID to nickname mapping
}

type reqFormat struct {
	Type      string          `json:"type,omitempty"`
	PubIP     string          `json:"pubip,omitempty"`
	ReqParams json.RawMessage `json:"reqparams,omitempty"`
	Body      json.RawMessage `json:"body,omitempty"`
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
		//libp2p.EnableHolePunching(),
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

	// Start a goroutine to periodically refresh reservations
	go cp.refreshReservations(ctx, *relayInfo)

	var reqSent reqFormat
	reqSent.Type = "register"
	reqSent.PubIP = GetPublicIP() // have too use a stun server to get public ip first and then send register command
	fmt.Println(reqSent.PubIP)
	stream, err := cp.Host.NewStream(context.Background(), relayInfo.ID, ChatProtocol)

	if err != nil {
		fmt.Println("[DEBUG]Error Opening stream to relay")
	}
	fmt.Println("[DEBUG]Opened atream to relay successsfully")
	reqJson, err := json.Marshal(reqSent)
	if err != nil {
		fmt.Println("[DEBUG]Error marshalling the req to be sent")
	}
	stream.Write([]byte(reqJson))

	time.Sleep(1 * time.Second)

	stream.Close()
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

	reader := bufio.NewReader(s)
	for {

		line, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println("[DEBUG]Error reading the bytes from the stream")
		}
		line = bytes.TrimRight(line, "\n")
		line = bytes.TrimRight(line, "\x00")
		var reqStruct reqFormat
		err = json.Unmarshal(line, &reqStruct)

		fmt.Println("[DEBUG] Raw input:", string(line))

		if err != nil {
			fmt.Println("[DEBUG]Error unmarshalling to reqStruc")
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
			resp := ServePostReq(reqStruct.ReqParams, reqStruct.Body)
			resp = bytes.TrimRight(resp, "\x00")
			_, err = s.Write(resp)
			if err != nil {
				fmt.Println("[DEBUG]Error writing resp bytes to relay")
				return
			}
		}

	}
}

func (cp *ChatPeer) Send(ctx context.Context, TargetIP string, targetPort string, jsonReq []byte, body []byte) ([]byte, error) {
	completeIP := TargetIP + ":" + targetPort

	var req reqFormat
	req.Type = "SendMsg"
	req.PubIP = completeIP
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

	reader := bufio.NewReader(stream)
	ack, err := reader.ReadString('\n')

	if err != nil {
		fmt.Println("[DEBUG]Error getting the acknowledgement")
		return nil, err
	}
	_ = ack //can be used if required

	var resp = make([]byte, 1024*4)
	reader.Read(resp)
	resp = bytes.TrimRight(resp, "\x00")
	defer stream.Close()

	return resp, err
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

func GetPublicIP() string {
	c, err := stun.Dial("udp4", "stun.l.google.com:19302")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	var xorAddr stun.XORMappedAddress
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	if err := c.Do(message, func(res stun.Event) {
		if res.Error != nil {
			panic(res.Error)
		}
		if err := xorAddr.GetFrom(res.Message); err != nil {
			panic(err)
		}
	}); err != nil {
		panic(err)
	}

	if xorAddr.IP.To4() == nil {
		panic("STUN returned an IPv6 address; IPv4 not available")
	}

	peerAd := xorAddr.IP.String() + ":" + strconv.Itoa(xorAddr.Port)
	return peerAd
}
