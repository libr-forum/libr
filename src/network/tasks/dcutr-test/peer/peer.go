// peer2.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	libp2p "github.com/libp2p/go-libp2p"
	holepunch "github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	identify "github.com/libp2p/go-libp2p/p2p/protocol/identify"

	host "github.com/libp2p/go-libp2p/core/host"
	network "github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	connmgr "github.com/libp2p/go-libp2p/p2p/net/connmgr"
	ma "github.com/multiformats/go-multiaddr"
)

const ChatProtocol = "/chat/1.0.0"

type ChatPeer struct {
	host        host.Host
	relayAddr   ma.Multiaddr
	relayID     peer.ID
	circuitAddr ma.Multiaddr
	peers       map[peer.ID]string
}

func NewChatPeer(relayAddr string) (*ChatPeer, error) {
	relayMA, err := ma.NewMultiaddr(relayAddr)
	if err != nil {
		return nil, err
	}

	relayInfo, err := peer.AddrInfoFromP2pAddr(relayMA)
	if err != nil {
		return nil, err
	}

	connMgr, err := connmgr.NewConnManager(100, 400)
	if err != nil {
		return nil, err
	}

	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.ConnectionManager(connMgr),
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		return nil, err
	}

	idSvc, _ := identify.NewIDService(h)
	getListenAddrs := func() []ma.Multiaddr { return h.Addrs() }
	hps, _ := holepunch.NewService(h, idSvc, getListenAddrs)
	_ = hps

	circuitAddr := relayMA.Encapsulate(ma.StringCast("/p2p-circuit"))
	circuitAddr = circuitAddr.Encapsulate(ma.StringCast("/p2p/" + h.ID().String()))

	cp := &ChatPeer{
		host:        h,
		relayAddr:   relayMA,
		relayID:     relayInfo.ID,
		circuitAddr: circuitAddr,
		peers:       make(map[peer.ID]string),
	}

	h.SetStreamHandler(ChatProtocol, cp.handleChatStream)

	return cp, nil
}

func (cp *ChatPeer) Start(ctx context.Context) error {
	relayInfo, _ := peer.AddrInfoFromP2pAddr(cp.relayAddr)
	if err := cp.host.Connect(ctx, *relayInfo); err != nil {
		return err
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println("[INFO] Your Listen Addresses:")
	for _, addr := range cp.host.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, cp.host.ID())
	}
	fmt.Println("[INFO] Relay Circuit Address:")
	fmt.Println(" ", cp.circuitAddr.String())
	fmt.Println("--------------------------------------------------")

	return nil
}

func (cp *ChatPeer) RequestPeerAddressesFromRelay(targetPeerID peer.ID) ([]ma.Multiaddr, error) {
	stream, err := cp.host.NewStream(context.Background(), cp.relayID, "/address-exchange/1.0.0")
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	_, err = stream.Write([]byte(targetPeerID.String() + "\n"))
	if err != nil {
		return nil, err
	}

	var addrs []ma.Multiaddr
	buf := bufio.NewScanner(stream)
	for buf.Scan() {
		line := buf.Text()
		if line == "END" {
			break
		}
		addr, err := ma.NewMultiaddr(line)
		if err == nil {
			addrs = append(addrs, addr)
		}
	}
	if err := buf.Err(); err != nil {
		return nil, err
	}

	return addrs, nil
}

func (cp *ChatPeer) ConnectToPeer(peerAddrStr, nickname string) error {
	peerMA, err := ma.NewMultiaddr(peerAddrStr)
	if err != nil {
		return err
	}
	peerInfo, err := peer.AddrInfoFromP2pAddr(peerMA)
	if err != nil {
		return err
	}

	fmt.Println("[INFO] Requesting peer addresses from relay...")
	targetAddrs, err := cp.RequestPeerAddressesFromRelay(peerInfo.ID)
	if err != nil {
		fmt.Println("[WARN] Failed to get from relay:", err)
	}

	fmt.Println("[DEBUG] Received addresses:")
	for _, addr := range targetAddrs {
		fmt.Println(" ", addr.String())
	}

	targetInfo := peer.AddrInfo{
		ID:    peerInfo.ID,
		Addrs: targetAddrs,
	}

	fmt.Println("[INFO] Trying direct connection...")
	err = cp.host.Connect(context.Background(), targetInfo)
	if err == nil {
		fmt.Println("[SUCCESS] Direct connection successful.")
		cp.peers[peerInfo.ID] = nickname
		return nil
	}

	fmt.Println("[WARN] Direct connection failed:", err)

	fmt.Println("[INFO] Falling back to relay circuit...")
	err = cp.host.Connect(context.Background(), *peerInfo)
	if err != nil {
		return err
	}
	fmt.Println("[SUCCESS] Connected via relay circuit.")
	cp.peers[peerInfo.ID] = nickname
	return nil
}

func (cp *ChatPeer) handleChatStream(s network.Stream) {
	remotePeer := s.Conn().RemotePeer()

	fmt.Println("[INFO] Incoming chat stream from:", remotePeer)

	// 🔥 Auto-register the incoming peer if not already in peers
	if _, exists := cp.peers[remotePeer]; !exists {
		nickname := remotePeer.String()[:8] // First 8 chars of peerID as nickname
		cp.peers[remotePeer] = nickname
		fmt.Println("[INFO] Auto-added peer:", nickname, "→", remotePeer)
	}

	defer s.Close()

	reader := bufio.NewReader(s)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("[INFO] Stream closed from:", remotePeer)
			return
		}
		fmt.Printf("[MESSAGE] From %s: %s", remotePeer, msg)
	}
}

func (cp *ChatPeer) SendMessage(peerID peer.ID, msg string) error {
	s, err := cp.host.NewStream(context.Background(), peerID, ChatProtocol)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = s.Write([]byte(msg + "\n"))
	return err
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run peer2.go <relay-address>")
		return
	}

	relayAddr := os.Args[1]
	ctx := context.Background()

	cp, err := NewChatPeer(relayAddr)
	if err != nil {
		panic(err)
	}

	if err := cp.Start(ctx); err != nil {
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\n[COMMAND] > ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		parts := strings.Split(line, " ")
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]

		switch cmd {
		case "connect":
			if len(parts) != 3 {
				fmt.Println("Usage: connect <peer-address> <nickname>")
				continue
			}
			err := cp.ConnectToPeer(parts[1], parts[2])
			if err != nil {
				fmt.Println("[ERROR] Failed to connect:", err)
			}
		case "send":
			if len(parts) < 3 {
				fmt.Println("Usage: send <nickname> <message>")
				continue
			}
			var target peer.ID
			found := false
			for pid, name := range cp.peers {
				if name == parts[1] {
					target = pid
					found = true
					break
				}
			}
			if !found {
				fmt.Println("[ERROR] No peer with nickname", parts[1])
				continue
			}
			msg := strings.Join(parts[2:], " ")
			err := cp.SendMessage(target, msg)
			if err != nil {
				fmt.Println("[ERROR] Failed to send:", err)
			}
		case "peers":
			for pid, name := range cp.peers {
				fmt.Println(name, "→", pid)
			}
		case "exit":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Commands: connect, send, peers, exit")
		}
	}
}
