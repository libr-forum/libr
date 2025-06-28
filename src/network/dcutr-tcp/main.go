package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

const (
	RelayProtocol = "/dcutr-relay/1.0.0"
	PeerProtocol  = "/dcutr-peer/1.0.0"
	ChatProtocol  = "/chat/1.0.0"
)

// Message types for DCUtR protocol
type MessageType string

const (
	MsgTypeConnect         MessageType = "CONNECT"
	MsgTypeSync            MessageType = "SYNC"
	MsgTypeSyncAck         MessageType = "SYNC_ACK"
	MsgTypeRelay           MessageType = "RELAY"
	MsgTypeHolePunch       MessageType = "HOLE_PUNCH"
	MsgTypeDirectTest      MessageType = "DIRECT_TEST"
	MsgTypeChat            MessageType = "CHAT"
	MsgTypePeerInfo        MessageType = "PEER_INFO"
	MsgTypeAddressExchange MessageType = "ADDRESS_EXCHANGE"
	MsgTypeKeepalive       MessageType = "KEEPALIVE"
)

type DCUtRMessage struct {
	Type      MessageType `json:"type"`
	PeerID    string      `json:"peer_id,omitempty"`
	TargetID  string      `json:"target_id,omitempty"`
	Addresses []string    `json:"addresses,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Data      string      `json:"data,omitempty"`
	RTT       int64       `json:"rtt,omitempty"`
}

type RelayServer struct {
	host     host.Host
	peers    map[peer.ID]*PeerInfo
	peersMux sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

type PeerInfo struct {
	ID        peer.ID
	Addresses []multiaddr.Multiaddr
	Stream    network.Stream
	LastSeen  time.Time
	// Keep connection alive
	keepaliveTimer *time.Timer
}

type PeerClient struct {
	host           host.Host
	relayAddr      multiaddr.Multiaddr
	relayStream    network.Stream
	connectedPeers map[peer.ID]network.Stream
	peersMux       sync.RWMutex
	rttMap         map[peer.ID]time.Duration
	rttMux         sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	// Address discovery
	knownPeers    map[peer.ID][]multiaddr.Multiaddr
	knownPeersMux sync.RWMutex
}

// RelayServer implementation
func NewRelayServer(port int) (*RelayServer, error) {
	// Generate a key pair for this host
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create libp2p host with better configuration
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
		libp2p.DisableRelay(),
		// Enable connection manager to handle more connections
		libp2p.ResourceManager(&network.NullResourceManager{}),
	)
	if err != nil {
		cancel()
		return nil, err
	}

	relay := &RelayServer{
		host:   h,
		peers:  make(map[peer.ID]*PeerInfo),
		ctx:    ctx,
		cancel: cancel,
	}

	// Set stream handler for relay protocol
	h.SetStreamHandler(protocol.ID(RelayProtocol), relay.handleRelayStream)

	// Start cleanup routine
	go relay.cleanupRoutine()

	fmt.Printf("Relay server started with ID: %s\n", h.ID().String())
	fmt.Printf("Listening on: %v\n", h.Addrs())

	return relay, nil
}

func (r *RelayServer) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.cleanupStaleConnections()
		}
	}
}

func (r *RelayServer) cleanupStaleConnections() {
	r.peersMux.Lock()
	defer r.peersMux.Unlock()

	now := time.Now()
	for peerID, info := range r.peers {
		if now.Sub(info.LastSeen) > 2*time.Minute {
			fmt.Printf("[RELAY] Cleaning up stale connection: %s\n", peerID.String()[:16])
			if info.Stream != nil {
				info.Stream.Close()
			}
			if info.keepaliveTimer != nil {
				info.keepaliveTimer.Stop()
			}
			delete(r.peers, peerID)
		}
	}
}

func (r *RelayServer) handleRelayStream(s network.Stream) {
	defer s.Close()

	peerID := s.Conn().RemotePeer()
	fmt.Printf("[RELAY] New connection from peer: %s\n", peerID.String())
	fmt.Printf("[RELAY] Remote address: %s\n", s.Conn().RemoteMultiaddr().String())

	// Store peer info with all available addresses
	r.peersMux.Lock()
	var addresses []multiaddr.Multiaddr

	// Get remote multiaddr
	remoteAddr := s.Conn().RemoteMultiaddr()
	peerAddr := remoteAddr.Encapsulate(multiaddr.StringCast("/p2p/" + peerID.String()))
	addresses = append(addresses, peerAddr)

	// Get all known addresses for this peer
	knownAddrs := r.host.Peerstore().Addrs(peerID)
	addresses = append(addresses, knownAddrs...)

	// Setup keepalive timer
	keepaliveTimer := time.AfterFunc(60*time.Second, func() {
		r.sendKeepalive(s)
	})

	r.peers[peerID] = &PeerInfo{
		ID:             peerID,
		Addresses:      addresses,
		Stream:         s,
		LastSeen:       time.Now(),
		keepaliveTimer: keepaliveTimer,
	}

	fmt.Printf("[RELAY] Stored peer info for %s with %d addresses\n", peerID.String(), len(addresses))
	fmt.Printf("[RELAY] Total connected peers: %d\n", len(r.peers))
	r.peersMux.Unlock()

	// Handle messages from this peer
	scanner := bufio.NewScanner(s)
	for scanner.Scan() {
		msgBytes := scanner.Bytes()
		if len(msgBytes) == 0 {
			continue
		}

		fmt.Printf("[RELAY] Received raw message from %s: %s\n", peerID.String()[:16], string(msgBytes))

		var msg DCUtRMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			fmt.Printf("[RELAY] Error unmarshaling message from %s: %v\n", peerID.String()[:16], err)
			continue
		}

		// Update last seen
		r.peersMux.Lock()
		if info, exists := r.peers[peerID]; exists {
			info.LastSeen = time.Now()
			// Reset keepalive timer
			if info.keepaliveTimer != nil {
				info.keepaliveTimer.Stop()
			}
			info.keepaliveTimer = time.AfterFunc(60*time.Second, func() {
				r.sendKeepalive(s)
			})
		}
		r.peersMux.Unlock()

		fmt.Printf("[RELAY] Parsed message from %s: Type=%s, TargetID=%s, Data=%s\n",
			peerID.String()[:16], msg.Type, msg.TargetID, msg.Data)

		r.handleMessage(peerID, &msg, s)
	}

	// Clean up on disconnect
	r.peersMux.Lock()
	if info, exists := r.peers[peerID]; exists {
		if info.keepaliveTimer != nil {
			info.keepaliveTimer.Stop()
		}
		delete(r.peers, peerID)
	}
	fmt.Printf("[RELAY] Peer disconnected: %s (remaining: %d)\n", peerID.String()[:16], len(r.peers))
	r.peersMux.Unlock()
}

func (r *RelayServer) sendKeepalive(s network.Stream) {
	keepaliveMsg := DCUtRMessage{
		Type:      MsgTypeKeepalive,
		Timestamp: time.Now().UnixNano(),
	}
	r.sendMessage(s, &keepaliveMsg)
}

func (r *RelayServer) handleMessage(from peer.ID, msg *DCUtRMessage, s network.Stream) {
	fmt.Printf("[RELAY] Handling message type %s from %s\n", msg.Type, from.String()[:16])

	switch msg.Type {
	case MsgTypeConnect:
		fmt.Printf("[RELAY] Processing CONNECT request from %s\n", from.String()[:16])
		r.handleConnectRequest(from, msg, s)
	case MsgTypeRelay:
		fmt.Printf("[RELAY] Processing RELAY message from %s to %s\n", from.String()[:16], msg.TargetID)
		r.handleRelayMessage(from, msg)
	case MsgTypeSync:
		fmt.Printf("[RELAY] Processing SYNC request from %s\n", from.String()[:16])
		r.handleSync(from, msg, s)
	case MsgTypePeerInfo:
		fmt.Printf("[RELAY] Processing PEER_INFO request from %s\n", from.String()[:16])
		r.handlePeerInfoRequest(from, msg, s)
	case MsgTypeKeepalive:
		fmt.Printf("[RELAY] Received keepalive from %s\n", from.String()[:16])
		// Just update last seen (already done above)
	default:
		fmt.Printf("[RELAY] Unknown message type %s from %s\n", msg.Type, from.String()[:16])
	}
}

func (r *RelayServer) handleConnectRequest(from peer.ID, msg *DCUtRMessage, s network.Stream) {
	// Store peer-reported public addresses if provided
	if len(msg.Addresses) > 0 {
		r.peersMux.Lock()
		if info, exists := r.peers[from]; exists {
			// Update with peer-reported addresses (these should be public)
			var newAddrs []multiaddr.Multiaddr
			for _, addrStr := range msg.Addresses {
				if addr, err := multiaddr.NewMultiaddr(addrStr); err == nil {
					newAddrs = append(newAddrs, addr)
					fmt.Printf("[RELAY] Updated public address for %s: %s\n", from.String()[:16], addrStr)
				}
			}
			if len(newAddrs) > 0 {
				info.Addresses = newAddrs // Replace with peer-reported public addresses
			}
		}
		r.peersMux.Unlock()
	}

	// Rest of the function remains the same...
	r.peersMux.RLock()
	var peerList []string
	var addressList []string

	for pID, info := range r.peers {
		if pID != from {
			peerList = append(peerList, pID.String())
			// Use the stored addresses (now public)
			for _, addr := range info.Addresses {
				addressList = append(addressList, addr.String())
			}
		}
	}
	r.peersMux.RUnlock()

	fmt.Printf("[RELAY] Responding to CONNECT from %s with %d peers\n",
		from.String()[:16], len(peerList))

	response := DCUtRMessage{
		Type:      MsgTypeConnect,
		Addresses: append(peerList, addressList...),
		Timestamp: time.Now().UnixNano(),
	}

	r.sendMessage(s, &response)
}

func (r *RelayServer) handlePeerInfoRequest(from peer.ID, msg *DCUtRMessage, s network.Stream) {
	if msg.TargetID == "" {
		return
	}

	targetID, err := peer.Decode(msg.TargetID)
	if err != nil {
		fmt.Printf("[RELAY] Invalid target peer ID in peer info request: %v\n", err)
		return
	}

	r.peersMux.RLock()
	targetPeer, exists := r.peers[targetID]
	r.peersMux.RUnlock()

	if !exists {
		fmt.Printf("[RELAY] Target peer %s not found for peer info request\n", targetID.String()[:16])
		return
	}

	// Send target peer's addresses back to requester
	var addressStrings []string
	for _, addr := range targetPeer.Addresses {
		addressStrings = append(addressStrings, addr.String())
	}

	response := DCUtRMessage{
		Type:      MsgTypePeerInfo,
		PeerID:    targetID.String(),
		Addresses: addressStrings,
		Timestamp: time.Now().UnixNano(),
	}

	r.sendMessage(s, &response)
}

func (r *RelayServer) handleRelayMessage(from peer.ID, msg *DCUtRMessage) {
	fmt.Printf("[RELAY] Relaying message from %s to target %s\n", from.String()[:16], msg.TargetID)

	targetID, err := peer.Decode(msg.TargetID)
	if err != nil {
		fmt.Printf("[RELAY] Invalid target peer ID: %v\n", err)
		return
	}

	r.peersMux.RLock()
	targetPeer, exists := r.peers[targetID]
	r.peersMux.RUnlock()

	if !exists {
		fmt.Printf("[RELAY] Target peer %s not connected (available: %d peers)\n",
			targetID.String()[:16], len(r.peers))
		r.peersMux.RLock()
		for pID := range r.peers {
			fmt.Printf("[RELAY]   - %s\n", pID.String()[:16])
		}
		r.peersMux.RUnlock()
		return
	}

	// Forward message to target peer
	msg.PeerID = from.String()
	fmt.Printf("[RELAY] Forwarding message to %s: Type=%s, Data=%s\n",
		targetID.String()[:16], msg.Type, msg.Data)
	r.sendMessage(targetPeer.Stream, msg)
}

func (r *RelayServer) handleSync(from peer.ID, msg *DCUtRMessage, s network.Stream) {
	// Calculate RTT and send SYNC_ACK
	response := DCUtRMessage{
		Type:      MsgTypeSyncAck,
		Timestamp: time.Now().UnixNano(),
		RTT:       time.Now().UnixNano() - msg.Timestamp,
	}
	r.sendMessage(s, &response)
}

func (r *RelayServer) sendMessage(s network.Stream, msg *DCUtRMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("[RELAY] Error marshaling message: %v\n", err)
		return
	}

	fmt.Printf("[RELAY] Sending message: %s\n", string(data))
	_, err = s.Write(append(data, '\n'))
	if err != nil {
		fmt.Printf("[RELAY] Error sending message: %v\n", err)
	}
}

// PeerClient implementation
func NewPeerClient(relayAddr string) (*PeerClient, error) {
	// Generate a key pair for this host
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create libp2p host with better NAT traversal support
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/0",
			"/ip6/::/tcp/0",
		),
		libp2p.DisableRelay(),
		libp2p.ResourceManager(&network.NullResourceManager{}),
		// Enable NAT port mapping
		libp2p.NATPortMap(),
	)
	if err != nil {
		cancel()
		return nil, err
	}

	relayMA, err := multiaddr.NewMultiaddr(relayAddr)
	if err != nil {
		cancel()
		return nil, err
	}

	client := &PeerClient{
		host:           h,
		relayAddr:      relayMA,
		connectedPeers: make(map[peer.ID]network.Stream),
		rttMap:         make(map[peer.ID]time.Duration),
		ctx:            ctx,
		cancel:         cancel,
		knownPeers:     make(map[peer.ID][]multiaddr.Multiaddr),
	}

	// Set stream handlers
	h.SetStreamHandler(protocol.ID(PeerProtocol), client.handlePeerStream)
	h.SetStreamHandler(protocol.ID(ChatProtocol), client.handleChatStream)

	fmt.Printf("Peer client started with ID: %s\n", h.ID().String())
	fmt.Printf("Listening on: %v\n", h.Addrs())

	return client, nil
}

// Add this method to PeerClient
func (p *PeerClient) discoverPublicIP() (string, error) {
	// Use a simple HTTP service to discover public IP
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (p *PeerClient) ConnectToRelay() error {
	ctx := context.Background()

	fmt.Printf("[PEER] Attempting to connect to relay: %s\n", p.relayAddr.String())

	// Connect to relay
	info, err := peer.AddrInfoFromP2pAddr(p.relayAddr)
	if err != nil {
		fmt.Printf("[PEER] Error parsing relay address: %v\n", err)
		return err
	}

	fmt.Printf("[PEER] Parsed relay info: ID=%s, Addrs=%v\n", info.ID.String(), info.Addrs)

	if err := p.host.Connect(ctx, *info); err != nil {
		fmt.Printf("[PEER] Error connecting to relay: %v\n", err)
		return err
	}

	// Open stream to relay
	s, err := p.host.NewStream(ctx, info.ID, protocol.ID(RelayProtocol))
	if err != nil {
		fmt.Printf("[PEER] Error opening stream to relay: %v\n", err)
		return err
	}

	p.relayStream = s

	fmt.Printf("[PEER] Connected to relay: %s\n", info.ID.String())

	// Discover public IP
	publicIP, err := p.discoverPublicIP()
	if err != nil {
		fmt.Printf("[PEER] Could not discover public IP: %v\n", err)
		publicIP = ""
	}

	// Get our listening port
	var listenPort string
	for _, addr := range p.host.Addrs() {
		if strings.Contains(addr.String(), "/tcp/") {
			if port, err := addr.ValueForProtocol(multiaddr.P_TCP); err == nil {
				listenPort = port
				break
			}
		}
	}

	// Create our public address
	var publicAddresses []string
	if publicIP != "" && listenPort != "" {
		publicAddr := fmt.Sprintf("/ip4/%s/tcp/%s/p2p/%s", publicIP, listenPort, p.host.ID().String())
		publicAddresses = append(publicAddresses, publicAddr)
		fmt.Printf("[PEER] Reporting public address: %s\n", publicAddr)
	}

	// Send connect message with public address
	connectMsg := DCUtRMessage{
		Type:      MsgTypeConnect,
		PeerID:    p.host.ID().String(),
		Addresses: publicAddresses, // Include our public address
		Timestamp: time.Now().UnixNano(),
	}

	fmt.Printf("[PEER] Sending CONNECT message to relay\n")
	p.sendMessage(s, &connectMsg)

	// Handle relay stream
	go p.handleRelayStream(s)

	// Start keepalive routine
	go p.keepaliveRoutine()

	return nil
}

func (p *PeerClient) keepaliveRoutine() {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			if p.relayStream != nil {
				keepaliveMsg := DCUtRMessage{
					Type:      MsgTypeKeepalive,
					Timestamp: time.Now().UnixNano(),
				}
				p.sendMessage(p.relayStream, &keepaliveMsg)
			}
		}
	}
}

func (p *PeerClient) handleRelayStream(s network.Stream) {
	fmt.Printf("[PEER] Starting relay stream handler\n")
	scanner := bufio.NewScanner(s)
	for scanner.Scan() {
		msgBytes := scanner.Bytes()
		if len(msgBytes) == 0 {
			continue
		}

		fmt.Printf("[PEER] Received message from relay: %s\n", string(msgBytes))

		var msg DCUtRMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			fmt.Printf("[PEER] Error unmarshaling relay message: %v\n", err)
			continue
		}

		fmt.Printf("[PEER] Parsed relay message: Type=%s, PeerID=%s, Data=%s\n",
			msg.Type, msg.PeerID, msg.Data)

		p.handleRelayMessage(&msg, s)
	}
	fmt.Printf("[PEER] Relay stream handler ended\n")
}

func (p *PeerClient) handleRelayMessage(msg *DCUtRMessage, relayStream network.Stream) {
	fmt.Printf("[PEER] Handling relay message type: %s\n", msg.Type)

	switch msg.Type {
	case MsgTypeConnect:
		fmt.Printf("[PEER] Available peers from relay: %v\n", msg.Addresses)
		// Parse and store peer addresses
		p.parseAndStorePeerAddresses(msg.Addresses)

	case MsgTypePeerInfo:
		fmt.Printf("[PEER] Received peer info for %s with addresses: %v\n", msg.PeerID, msg.Addresses)
		p.storePeerAddresses(msg.PeerID, msg.Addresses)

	case MsgTypeHolePunch:
		fmt.Printf("[PEER] Received HOLE_PUNCH message from %s\n", msg.PeerID)
		go p.handleHolePunch(msg)

	case MsgTypeRelay:
		fmt.Printf("[PEER] Received relayed message from %s: %s\n", msg.PeerID, msg.Data)
		// Handle different types of relayed messages
		if msg.Data == "HOLE_PUNCH_INIT" {
			fmt.Printf("[PEER] Processing HOLE_PUNCH_INIT from %s\n", msg.PeerID)
			holePunchMsg := DCUtRMessage{
				Type:      MsgTypeHolePunch,
				PeerID:    msg.PeerID,
				Data:      "HOLE_PUNCH_INIT",
				Timestamp: msg.Timestamp,
			}
			p.handleHolePunch(&holePunchMsg)
		} else if msg.Data == "ADDRESS_EXCHANGE" {
			fmt.Printf("[PEER] Processing ADDRESS_EXCHANGE from %s\n", msg.PeerID)
			p.requestPeerInfo(msg.PeerID)
		}

	case MsgTypeKeepalive:
		fmt.Printf("[PEER] Received keepalive from relay\n")

	default:
		fmt.Printf("[PEER] Unknown relay message type: %s\n", msg.Type)
	}
}

func (p *PeerClient) parseAndStorePeerAddresses(addresses []string) {
	p.knownPeersMux.Lock()
	defer p.knownPeersMux.Unlock()

	for _, addrStr := range addresses {
		// Try to parse as peer ID first
		if peerID, err := peer.Decode(addrStr); err == nil {
			if _, exists := p.knownPeers[peerID]; !exists {
				p.knownPeers[peerID] = []multiaddr.Multiaddr{}
			}
		} else if addr, err := multiaddr.NewMultiaddr(addrStr); err == nil {
			// Parse multiaddr and extract peer ID
			if peerID, err := addr.ValueForProtocol(multiaddr.P_P2P); err == nil {
				if pid, err := peer.Decode(peerID); err == nil {
					p.knownPeers[pid] = append(p.knownPeers[pid], addr)
					// Also add to peerstore
					p.host.Peerstore().AddAddr(pid, addr, time.Hour)
				}
			}
		}
	}
}

func (p *PeerClient) storePeerAddresses(peerIDStr string, addresses []string) {
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		return
	}

	p.knownPeersMux.Lock()
	defer p.knownPeersMux.Unlock()

	var addrs []multiaddr.Multiaddr
	for _, addrStr := range addresses {
		if addr, err := multiaddr.NewMultiaddr(addrStr); err == nil {
			addrs = append(addrs, addr)
			// Also add to peerstore with longer TTL
			p.host.Peerstore().AddAddr(peerID, addr, 2*time.Hour)
		}
	}

	p.knownPeers[peerID] = addrs
	fmt.Printf("[PEER] Stored %d addresses for peer %s\n", len(addrs), peerID.String()[:16])
}

func (p *PeerClient) requestPeerInfo(peerIDStr string) {
	if p.relayStream == nil {
		return
	}

	peerInfoMsg := DCUtRMessage{
		Type:      MsgTypePeerInfo,
		TargetID:  peerIDStr,
		Timestamp: time.Now().UnixNano(),
	}

	p.sendMessage(p.relayStream, &peerInfoMsg)
}

func (p *PeerClient) InitiateHolePunch(targetPeerID string) error {
	fmt.Printf("[PEER] Initiating hole punch to: %s\n", targetPeerID)

	targetID, err := peer.Decode(targetPeerID)
	if err != nil {
		fmt.Printf("[PEER] Error decoding target peer ID: %v\n", err)
		return err
	}

	// First, request peer info to get addresses
	p.requestPeerInfo(targetPeerID)

	// Wait a bit for peer info response
	time.Sleep(500 * time.Millisecond)

	// Send address exchange request through relay
	addressExchangeMsg := DCUtRMessage{
		Type:      MsgTypeRelay,
		TargetID:  targetPeerID,
		Data:      "ADDRESS_EXCHANGE",
		Timestamp: time.Now().UnixNano(),
	}

	fmt.Printf("[PEER] Sending ADDRESS_EXCHANGE to %s via relay\n", targetPeerID)
	p.sendMessage(p.relayStream, &addressExchangeMsg)

	// Wait a bit more for address exchange
	time.Sleep(500 * time.Millisecond)

	// Send hole punch initiation through relay
	relayMsg := DCUtRMessage{
		Type:      MsgTypeRelay,
		TargetID:  targetPeerID,
		Data:      "HOLE_PUNCH_INIT",
		Timestamp: time.Now().UnixNano(),
	}

	fmt.Printf("[PEER] Sending HOLE_PUNCH_INIT to %s via relay\n", targetPeerID)
	p.sendMessage(p.relayStream, &relayMsg)

	// Start simultaneous connect attempt
	fmt.Printf("[PEER] Starting simultaneous connect attempt to %s\n", targetID.String())
	go p.attemptDirectConnect(targetID)

	return nil
}

func (p *PeerClient) handleHolePunch(msg *DCUtRMessage) {
	fmt.Printf("[PEER] Handling hole punch message: %s from %s\n", msg.Data, msg.PeerID)

	if msg.Data == "HOLE_PUNCH_INIT" {
		peerID, err := peer.Decode(msg.PeerID)
		if err != nil {
			fmt.Printf("[PEER] Invalid peer ID in hole punch: %v\n", err)
			return
		}

		fmt.Printf("[PEER] Starting simultaneous connect to %s in response to hole punch\n", peerID.String())
		// Start simultaneous connect attempt
		go p.attemptDirectConnect(peerID)
	}
}

func (p *PeerClient) attemptDirectConnect(targetID peer.ID) {
	fmt.Printf("[PEER] Starting direct connection attempts to: %s\n", targetID.String())

	// Get all known addresses for the target peer
	p.knownPeersMux.RLock()
	knownAddrs := p.knownPeers[targetID]
	p.knownPeersMux.RUnlock()

	// Also get addresses from peerstore
	peerstoreAddrs := p.host.Peerstore().Addrs(targetID)

	// Combine all addresses
	allAddrs := append(knownAddrs, peerstoreAddrs...)

	fmt.Printf("[PEER] Known addresses for %s: %v\n", targetID.String()[:16], allAddrs)

	if len(allAddrs) == 0 {
		fmt.Printf("[PEER] No addresses known for %s, requesting info\n", targetID.String()[:16])
		p.requestPeerInfo(targetID.String())
		time.Sleep(1 * time.Second)

		// Try again with updated addresses
		p.knownPeersMux.RLock()
		knownAddrs = p.knownPeers[targetID]
		p.knownPeersMux.RUnlock()
		peerstoreAddrs = p.host.Peerstore().Addrs(targetID)
		allAddrs = append(knownAddrs, peerstoreAddrs...)
	}

	// Add addresses to peerstore for connection attempts
	for _, addr := range allAddrs {
		p.host.Peerstore().AddAddr(targetID, addr, time.Minute)
	}

	// Try multiple connection attempts with different timing
	for i := 0; i < 10; i++ {
		fmt.Printf("[PEER] Connection attempt %d/10 to %s\n", i+1, targetID.String()[:16])

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// Try to connect directly
		err := p.host.Connect(ctx, peer.AddrInfo{
			ID:    targetID,
			Addrs: allAddrs,
		})

		if err == nil {
			cancel()
			fmt.Printf("[PEER] ✓ Direct connection established with: %s\n", targetID.String())

			// Test the connection
			p.testDirectConnection(targetID)
			return
		}

		fmt.Printf("[PEER] Connection attempt %d failed: %v\n", i+1, err)
		cancel()

		// Progressive backoff
		backoff := time.Duration(100+i*100) * time.Millisecond
		time.Sleep(backoff)
	}

	fmt.Printf("[PEER] ✗ Failed to establish direct connection with: %s after 10 attempts\n", targetID.String())
}

func (p *PeerClient) testDirectConnection(peerID peer.ID) {
	fmt.Printf("[PEER] Testing direct connection to %s\n", peerID.String()[:16])
	start := time.Now()

	// Open stream for direct communication
	s, err := p.host.NewStream(context.Background(), peerID, protocol.ID(PeerProtocol))
	if err != nil {
		fmt.Printf("[PEER] Error opening direct stream to %s: %v\n", peerID.String()[:16], err)
		return
	}

	fmt.Printf("[PEER] Direct stream opened to %s\n", peerID.String()[:16])

	// Measure RTT
	testMsg := DCUtRMessage{
		Type:      MsgTypeDirectTest,
		Timestamp: time.Now().UnixNano(),
	}

	fmt.Printf("[PEER] Sending direct test message to %s\n", peerID.String()[:16])
	p.sendMessage(s, &testMsg)

	// Wait for response with timeout
	s.SetReadDeadline(time.Now().Add(5 * time.Second))
	scanner := bufio.NewScanner(s)
	if scanner.Scan() {
		rtt := time.Since(start)
		p.rttMux.Lock()
		p.rttMap[peerID] = rtt
		p.rttMux.Unlock()

		fmt.Printf("[PEER] ✓ Direct connection RTT to %s: %v\n", peerID.String()[:16], rtt)

		// Store the stream for chat
		p.peersMux.Lock()
		p.connectedPeers[peerID] = s
		p.peersMux.Unlock()

		fmt.Printf("[PEER] Direct connection to %s ready for chat\n", peerID.String()[:16])
	} else {
		fmt.Printf("[PEER] No response to direct test from %s\n", peerID.String()[:16])
		s.Close()
	}
}

func (p *PeerClient) handlePeerStream(s network.Stream) {
	defer s.Close()

	peerID := s.Conn().RemotePeer()
	fmt.Printf("[PEER] Handling peer stream from %s\n", peerID.String()[:16])

	scanner := bufio.NewScanner(s)
	for scanner.Scan() {
		var msg DCUtRMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			fmt.Printf("[PEER] Error unmarshaling peer message: %v\n", err)
			continue
		}

		if msg.Type == MsgTypeDirectTest {
			fmt.Printf("[PEER] Received direct test from %s, echoing back\n", peerID.String()[:16])
			// Echo back for RTT measurement
			response := DCUtRMessage{
				Type:      MsgTypeDirectTest,
				Timestamp: time.Now().UnixNano(),
			}
			p.sendMessage(s, &response)
		}
	}
}

func (p *PeerClient) handleChatStream(s network.Stream) {
	peerID := s.Conn().RemotePeer()
	fmt.Printf("\n=== Direct chat stream from %s ===\n", peerID.String())

	scanner := bufio.NewScanner(s)
	for scanner.Scan() {
		var msg DCUtRMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			fmt.Printf("[CHAT] Error unmarshaling chat message: %v\n", err)
			continue
		}

		if msg.Type == MsgTypeChat {
			fmt.Printf("[%s]: %s\n", peerID.String()[:16], msg.Data)
		}
	}
}

func (p *PeerClient) sendMessage(s network.Stream, msg *DCUtRMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("[PEER] Error marshaling message: %v\n", err)
		return
	}

	_, err = s.Write(append(data, '\n'))
	if err != nil {
		fmt.Printf("[PEER] Error sending message: %v\n", err)
	}
}

func (p *PeerClient) StartChat(targetPeerID string) error {
	peerID, err := peer.Decode(targetPeerID)
	if err != nil {
		return err
	}

	// Check if we have a direct connection
	p.peersMux.RLock()
	_, exists := p.connectedPeers[peerID]
	p.peersMux.RUnlock()

	var chatStream network.Stream

	if !exists {
		// Try to establish direct connection first
		fmt.Printf("[PEER] No existing connection, initiating hole punch...\n")
		if err := p.InitiateHolePunch(targetPeerID); err != nil {
			return err
		}

		// Wait longer for connection to establish
		fmt.Printf("[PEER] Waiting for direct connection to establish...\n")
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			p.peersMux.RLock()
			_, exists = p.connectedPeers[peerID]
			p.peersMux.RUnlock()
			if exists {
				break
			}
			fmt.Printf("[PEER] Still waiting... (%d/10)\n", i+1)
		}

		if !exists {
			return fmt.Errorf("could not establish direct connection after 10 seconds")
		}
	}

	// Open chat stream
	chatStream, err = p.host.NewStream(context.Background(), peerID, protocol.ID(ChatProtocol))
	if err != nil {
		return fmt.Errorf("error opening chat stream: %v", err)
	}

	fmt.Printf("Chat started with %s (RTT: %v)\n", peerID.String()[:16], p.getRTT(peerID))
	fmt.Printf("Type messages (press Enter to send, 'quit' to exit):\n")

	// Read user input and send messages
	go func() {
		defer chatStream.Close()
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)

			if text == "quit" {
				fmt.Printf("Chat ended with %s\n", peerID.String()[:16])
				return
			}

			if text == "" {
				continue
			}

			chatMsg := DCUtRMessage{
				Type:      MsgTypeChat,
				Data:      text,
				Timestamp: time.Now().UnixNano(),
			}

			p.sendMessage(chatStream, &chatMsg)
		}
	}()

	return nil
}

func (p *PeerClient) getRTT(peerID peer.ID) time.Duration {
	p.rttMux.RLock()
	defer p.rttMux.RUnlock()
	return p.rttMap[peerID]
}

func (p *PeerClient) ListConnectedPeers() {
	fmt.Printf("\n=== Connected Peers ===\n")

	p.peersMux.RLock()
	if len(p.connectedPeers) == 0 {
		fmt.Printf("No direct connections established\n")
	} else {
		for peerID := range p.connectedPeers {
			rtt := p.getRTT(peerID)
			fmt.Printf("- %s (RTT: %v)\n", peerID.String(), rtt)
		}
	}
	p.peersMux.RUnlock()

	fmt.Printf("\n=== Known Peers ===\n")
	p.knownPeersMux.RLock()
	if len(p.knownPeers) == 0 {
		fmt.Printf("No known peers\n")
	} else {
		for peerID, addrs := range p.knownPeers {
			fmt.Printf("- %s (%d addresses)\n", peerID.String(), len(addrs))
		}
	}
	p.knownPeersMux.RUnlock()
	fmt.Printf("\n")
}

func (p *PeerClient) Close() {
	p.cancel()
	if p.relayStream != nil {
		p.relayStream.Close()
	}

	p.peersMux.Lock()
	for _, stream := range p.connectedPeers {
		stream.Close()
	}
	p.peersMux.Unlock()

	p.host.Close()
}

func (r *RelayServer) Close() {
	r.cancel()

	r.peersMux.Lock()
	for _, info := range r.peers {
		if info.Stream != nil {
			info.Stream.Close()
		}
		if info.keepaliveTimer != nil {
			info.keepaliveTimer.Stop()
		}
	}
	r.peersMux.Unlock()

	r.host.Close()
}

// Main function demonstrating usage
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [relay|peer] [options]")
		fmt.Println("  relay <port>")
		fmt.Println("  peer <relay_multiaddr>")
		return
	}

	switch os.Args[1] {
	case "relay":
		port := 9000
		if len(os.Args) > 2 {
			fmt.Sscanf(os.Args[2], "%d", &port)
		}

		relay, err := NewRelayServer(port)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Relay server running on port %d. Press Ctrl+C to stop.\n", port)

		// Graceful shutdown
		defer relay.Close()
		select {} // Keep running

	case "peer":
		if len(os.Args) < 3 {
			fmt.Println("Please provide relay multiaddr")
			fmt.Println("Example: /ip4/127.0.0.1/tcp/9000/p2p/12D3KooW...")
			return
		}

		peerClient, err := NewPeerClient(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		defer peerClient.Close()

		if err := peerClient.ConnectToRelay(); err != nil {
			log.Fatal(err)
		}

		fmt.Println("\nCommands:")
		fmt.Println("  list                - list connected and known peers")
		fmt.Println("  punch <peer_id>     - initiate hole punching")
		fmt.Println("  chat <peer_id>      - start direct chat")
		fmt.Println("  info <peer_id>      - request peer info")
		fmt.Println("  quit                - exit")
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			input, _ := reader.ReadString('\n')
			parts := strings.Fields(strings.TrimSpace(input))

			if len(parts) == 0 {
				continue
			}

			switch parts[0] {
			case "list":
				peerClient.ListConnectedPeers()

			case "punch":
				if len(parts) > 1 {
					fmt.Printf("Initiating hole punch to %s...\n", parts[1])
					if err := peerClient.InitiateHolePunch(parts[1]); err != nil {
						fmt.Printf("Error: %v\n", err)
					}
				} else {
					fmt.Println("Usage: punch <peer_id>")
				}

			case "chat":
				if len(parts) > 1 {
					fmt.Printf("Starting chat with %s...\n", parts[1])
					if err := peerClient.StartChat(parts[1]); err != nil {
						fmt.Printf("Error: %v\n", err)
					}
				} else {
					fmt.Println("Usage: chat <peer_id>")
				}

			case "info":
				if len(parts) > 1 {
					fmt.Printf("Requesting info for %s...\n", parts[1])
					peerClient.requestPeerInfo(parts[1])
				} else {
					fmt.Println("Usage: info <peer_id>")
				}

			case "quit":
				fmt.Println("Goodbye!")
				return

			default:
				fmt.Printf("Unknown command: %s\n", parts[0])
			}
		}

	default:
		fmt.Println("Invalid command. Use 'relay' or 'peer'")
	}
}
