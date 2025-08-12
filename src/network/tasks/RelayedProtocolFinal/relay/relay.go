package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"math/big"
	"sort"
	"strings"
	"sync"

	//Peers "chatprotocol/peer"

	"context"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	relay "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	ma "github.com/multiformats/go-multiaddr"
)

type RelayDist struct {
	relayID string
	dist    *big.Int
}

const ChatProtocol = protocol.ID("/chat/1.0.0")

//var RelayMultiAddrList = []string{"/dns4/0.tcp.in.ngrok.io/tcp/14395/p2p/12D3KooWLBVV1ty7MwJQos34jy1WqGrfkb3bMAfxUJzCgwTBQ2pn",}

type reqFormat struct {
	Type      string          `json:"type,omitempty"`
	//PubIP     string          `json:"pubip,omitempty"`
	PeerID    string			`json:"peerid"`
	ReqParams json.RawMessage `json:"reqparams,omitempty"`
	Body      json.RawMessage `json:"body,omitempty"`
}

// var (
// 	IDmap = make(map[string]string)
// 	mu    sync.RWMutex
// )

var (
	ConnectedPeers []string 
	mu sync.RWMutex
)

var RelayHost host.Host

type respFormat struct {
	Type string `json:"type"`
	Resp []byte `json:"resp"`
}

type RelayEvents struct{}

func (re *RelayEvents) Listen(net network.Network, addr ma.Multiaddr)      {}
func (re *RelayEvents) ListenClose(net network.Network, addr ma.Multiaddr) {}
func (re *RelayEvents) Connected(net network.Network, conn network.Conn) {
	fmt.Printf("[INFO] Peer connected: %s\n", conn.RemotePeer())
}
func (re *RelayEvents) Disconnected(net network.Network, conn network.Conn) {
	fmt.Printf("[INFO] Peer disconnected: %s\n", conn.RemotePeer())
	// Remove peer from IDmap if needed
	mu.Lock()
	// for pubip, pid := range IDmap {
	// 	if pid == conn.RemotePeer().String() {
	// 		delete(IDmap, pubip)
	// 		break
	// 	}
	// }
	if contains(ConnectedPeers,conn.RemotePeer().String()){
		remove(ConnectedPeers, conn.RemotePeer().String())
	}
	mu.Unlock()
}

const sheetWebAppURL = "https://script.google.com/macros/s/AKfycbzQSQ1rKykcp-HVC0qEO4-C8GhEtKVZ3S5u2iR91-nZR9jOOWkvhb7K73QSmDmjSdmN/exec"

func main() {
	// fmt.Println("123")
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatalf("Error loading .env file")
	// }

	// // Fetch values

	// sheetURL := os.Getenv("sheetWebAppURL")
	// //fmt.Println(sheetURL)
	// sheetWebAppURL = sheetURL

	// Create connection manager
	fmt.Println("[DEBUG] Creating connection manager...")
	connMgr, err := connmgr.NewConnManager(100, 400)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create connection manager: %v", err)
	}

	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		// handle error
		panic(err)
	}
	fmt.Println("[DEBUG] Creating relay host...")

	RelayHost, err = libp2p.New(
		libp2p.Identity(privKey),

		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/443/ws"), // Changed from /tcp/4567 ?
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.ConnectionManager(connMgr),
		libp2p.EnableNATService(),
		libp2p.EnableRelayService(),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(websocket.New), // Add the websocket transport
	)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create relay host: %v", err)
	}
	RelayHost.Network().Notify(&RelayEvents{})
	relayMultiaddrFull := fmt.Sprintf("/dns4/libr-relay.onrender.com/tcp/443/wss/p2p/%s", RelayHost.ID().String())

	defer func() {
		fmt.Println("[DEBUG] Closing relay host...")
		deleteRelayAddrFromSheet(relayMultiaddrFull)
		RelayHost.Close()
	}()
	customRelayResources := relay.Resources{
		Limit: &relay.RelayLimit{
			Duration: 30 * time.Minute,
			Data:     1 << 20, // 1MB data limit per stream
		},
		ReservationTTL:         time.Hour,
		MaxReservations:        512,
		MaxCircuits:            64,
		BufferSize:             4096,
		MaxReservationsPerPeer: 10,
		MaxReservationsPerIP:   100, // Increased from the default of 8
		MaxReservationsPerASN:  64,
	}

	// Enable circuit relay service
	fmt.Println("[DEBUG] Enabling circuit relay service...")
	_, err = relay.New(RelayHost, relay.WithResources(customRelayResources))
	if err != nil {
		log.Fatalf("[ERROR] Failed to enable relay service: %v", err)
	}

	fmt.Printf("[INFO] Relay started!\n")
	fmt.Printf("[INFO] Peer ID: %s\n", RelayHost.ID())

	// Print all addresses
	for _, addr := range RelayHost.Addrs() {
		fmt.Printf("[INFO] Relay Address: %s/p2p/%s\n", addr, RelayHost.ID())
	}

	//relayMultiaddrFull :=  fmt.Sprintf("/dns4/0.tcp.in.ngrok.io/tcp/%s/p2p/%s","port_number", RelayHost.ID().String())

	go uploadRelayAddrToSheet(relayMultiaddrFull)

	RelayHost.SetStreamHandler("/chat/1.0.0", handleChatStream)
	go func() {
		for {
			fmt.Println(ConnectedPeers)
			time.Sleep(30 * time.Second)
		}
	}()

	// go func() {
	// 	log.Println("ENTERING GO ROUTINE FOR HEALTH CHECK SERVER")

	// 	check := func(w http.ResponseWriter, r *http.Request) {
	// 		log.Println("[DEBUG] /check endpoint hit")
	// 		if r.Method == http.MethodGet {
	// 			w.WriteHeader(http.StatusOK)
	// 			w.Write([]byte("running"))
	// 		} else {
	// 			w.WriteHeader(http.StatusMethodNotAllowed)
	// 		}
	// 	}

	// 	http.HandleFunc("/check", check)

	// 	port := os.Getenv("PORT")
	// 	if port == "" {
	// 		port = "8080"
	// 	}

	// 	addr := fmt.Sprintf(":%s", port)
	// 	log.Printf("[INFO] Starting health check server on %s", addr)

	// 	if err := http.ListenAndServe(addr, nil); err != nil {
	// 		log.Fatalf("[ERROR] Failed to start health check server: %v", err)
	// 	}
	// }()

	addr, _ := fetchRelayAddrsFromSheet()
	go PingTargets(addr, 5*time.Minute)

	fmt.Println("[DEBUG] Waiting for interrupt signal...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("[INFO] Shutting down relay...")
}

func remove(slice []string, val string) []string {
    for i, item := range slice {
        if item == val {
            // Remove element at index i
            return append(slice[:i], slice[i+1:]...)
        }
    }
    // If not found, return original slice
    return slice
}

func PingTargets(addresses []string, interval time.Duration) {
	go func() {
		for {
			for _, multiAddrStr := range addresses {
				// Parse the multiaddress string
				maddr, err := ma.NewMultiaddr(multiAddrStr)
				if err != nil {
					log.Printf("[WARN] Could not parse multiaddress %s: %v\n", multiAddrStr, err)
					continue
				}

				// Extract the domain name
				host, err := maddr.ValueForProtocol(ma.P_DNS4)
				if err != nil {
					// Fallback for P_DNS6 or other domain protocols if needed
					host, err = maddr.ValueForProtocol(ma.P_DNS6)
					if err != nil {
						log.Printf("[WARN] Could not extract host from multiaddress %s: %v\n", multiAddrStr, err)
						continue
					}
				}

				// Construct the final HTTP URL for the health check
				pingURL := fmt.Sprintf("https://%s/check", host)

				// Ping the valid URL
				resp, err := http.Get(pingURL)
				if err != nil {
					log.Printf("[WARN] Failed to ping %s: %v\n", pingURL, err)
					continue
				}
				resp.Body.Close()
				log.Printf("[INFO] Pinged %s — Status: %s\n", pingURL, resp.Status)
			}
			time.Sleep(interval)
		}
	}()
}

func contains(arr []string, target string) bool {
	for _, vals := range arr {
		if vals == target {
			return true
		}
	}
	return false
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
			peerID2 := req.PeerID

			if(peerID2 != peerID.String()){
				fmt.Println("PEER ID MISMATCH")
				return 
			}


			fmt.Printf("[INFO]Given peerID is %s \n", req.PeerID)
			fmt.Println("[INFO]Registering the peer into relay")
			mu.Lock()
			//IDmap[req.PubIP] = peerID.String()
			ConnectedPeers = append(ConnectedPeers, peerID.String())
			mu.Unlock()
		}

		if req.Type == "SendMsg" {
			mu.RLock()
			var targetPeerID string

			if contains(ConnectedPeers, req.PeerID) {
				targetPeerID = req.PeerID
			}
			mu.RUnlock()

			// checks if the target peer is connected to the relay or some other relay
			// have to handle some logic here but later

			if targetPeerID == "" {
				fmt.Println("[DEBUG]This peer is not on this relay, contacting other relay")
				targetRelayAddr := GetRelayAddr(req.PeerID)

				var forwardReq reqFormat
				forwardReq.Body = req.Body
				forwardReq.ReqParams = req.ReqParams
				forwardReq.PeerID = req.PeerID
				forwardReq.Type = "forward"

				relayMA, err := ma.NewMultiaddr(targetRelayAddr)
				if err != nil {
					fmt.Println("[DEBUG] Failed to parse relay multiaddr:", err)
					return
				}

				TargetRelayInfo, err := peer.AddrInfoFromP2pAddr(relayMA)
				if err != nil {
					fmt.Println("[DEBUG] Failed to parse target relay info:", err)
					return
				}

				err = RelayHost.Connect(context.Background(), *TargetRelayInfo)
				if err != nil {
					fmt.Println("[DEBUG] Failed to connect to target relay:", err)
					return
				}

				forwardStream, err := RelayHost.NewStream(context.Background(), TargetRelayInfo.ID, ChatProtocol)
				if err != nil {
					fmt.Println("[DEBUG] Failed to open stream to target relay:", err)
					return
				}
				defer forwardStream.Close()

				jsonForwardReq, err := json.Marshal(forwardReq)
				if err != nil {
					fmt.Println("[DEBUG] Failed to marshal forward request:", err)
					return
				}

				_, err = forwardStream.Write(append(jsonForwardReq, '\n'))
				if err != nil {
					fmt.Println("[DEBUG] Failed to write forward request to stream:", err)
					return
				}

				buf := make([]byte, 4096)
				respReader := bufio.NewReader(forwardStream)
				_, err = respReader.Read(buf)
				buf = bytes.TrimRight(buf, "\x00")
				var resp respFormat
				resp.Type = "GET"
				resp.Resp = buf
				fmt.Printf("[Debug]Frowarded Resp from relay : %s : %+v \n", TargetRelayInfo.ID.String(), resp)

				if err != nil {
					fmt.Println("[DEBUG] Error reading response from target relay:", err)
					return
				}

				_, err = s.Write(resp.Resp)
				defer s.Close()
				if err != nil {
					fmt.Println("[DEBUG] Error sending back to original sender:", err)
					return
				}

			} else {
				fmt.Println("Target peer ID: ", targetPeerID)
				if RelayHost == nil {
					fmt.Println("[FATAL] RelayHost is nil!")
					return
				}

				relayID := RelayHost.ID()

				fmt.Println("1")
				targetID, err := peer.Decode(targetPeerID)
				fmt.Println("2")


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
				fmt.Println("[DEBUG]", fullAddr.String())
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
				_ = jsonResp // if required whole jsonResp can be sent but it makes unmarhsalling the response harder for the client
				fmt.Println("[DEBUG]Raw Resp :", string(resp.Resp))
				_, err = s.Write(resp.Resp)
				if err != nil {
					fmt.Println("[DEBUG]Error sending response back")
				}
				defer s.Close()
				defer sendStream.Close()
			}
		}

		if req.Type == "forward" {
			mu.RLock()
			var targetPeerID string
			if(contains(ConnectedPeers, req.PeerID)){

				targetPeerID = req.PeerID
			}
			mu.RUnlock()

			if targetPeerID == "" {
				fmt.Println("[DEBUG] Target peer not found in this relay")
				s.Write([]byte("Target peer not found"))
				return
			}

			targetID, err := peer.Decode(targetPeerID)
			if err != nil {
				fmt.Println("[DEBUG] Invalid target peer ID")
				return
			}

			// Build relayed addr
			relayID := RelayHost.ID()
			relayBaseAddr, _ := ma.NewMultiaddr("/p2p/" + relayID.String())
			circuitAddr, _ := ma.NewMultiaddr("/p2p-circuit")
			targetAddr, _ := ma.NewMultiaddr("/p2p/" + targetID.String())
			fullAddr := relayBaseAddr.Encapsulate(circuitAddr).Encapsulate(targetAddr)

			addrInfo, err := peer.AddrInfoFromP2pAddr(fullAddr)
			if err != nil {
				fmt.Println("[DEBUG] Invalid relayed address")
				return
			}

			RelayHost.Peerstore().AddAddrs(addrInfo.ID, addrInfo.Addrs, peerstore.PermanentAddrTTL)

			err = RelayHost.Connect(context.Background(), *addrInfo)
			if err != nil {
				fmt.Println("[DEBUG] Failed to connect to target peer at this relay")
				return
			}

			sendStream, err := RelayHost.NewStream(context.Background(), targetID, ChatProtocol)
			if err != nil {
				fmt.Println("[DEBUG] Failed to open stream to target peer")
				return
			}
			defer sendStream.Close()

			jsonReqServer, err := json.Marshal(req)
			if err != nil {
				fmt.Println("[DEBUG]Error marshalling the req for server ")
			}
			_, err = sendStream.Write(append(jsonReqServer, '\n'))

			if err != nil {
				fmt.Println("[DEBUG]Error sending message despite stream opened")
				return
			}
			//s.Write([]byte("Success\n"))

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
			_ = jsonResp // if required whole jsonResp can be sent but it makes unmarhsalling the response harder for the client
			fmt.Println("[DEBUG]Raw Resp :", string(resp.Resp))
			_, err = s.Write(resp.Resp)
			if err != nil {
				fmt.Println("[DEBUG]Error sending response back")
			}
			defer s.Close()
			defer sendStream.Close()
		}
	}
}

func GetRelayAddr(peerID string) string {
	RelayMultiAddrList, err := fetchRelayAddrsFromSheet()

	if err != nil {
		fmt.Println("[DEBUG]Error getting addr from the sheet")
	}
	var relayList []string
	for _, multiaddr := range RelayMultiAddrList {
		parts := strings.Split(multiaddr, "/")
		relayList = append(relayList, parts[len(parts)-1])
	}

	var distmap []RelayDist

	h1 := sha256.New() // Use sha256.New() for SHA-256
	h1.Write([]byte(peerID))
	peerIDhash := hex.EncodeToString(h1.Sum(nil))

	for _, relay := range relayList {

		h_R := sha256.New() // Use sha256.New() for SHA-256
		h_R.Write([]byte(relay))
		RelayIDhash := hex.EncodeToString(h_R.Sum(nil))

		dist := XorHexToBigInt(peerIDhash, RelayIDhash)

		distmap = append(distmap, RelayDist{dist: dist, relayID: relay})
	}

	sort.Slice(distmap, func(i, j int) bool {
		return distmap[i].dist.Cmp(distmap[j].dist) < 0
	})

	relayIDused := distmap[0].relayID

	var relayAddr string

	for _, multiaddr := range RelayMultiAddrList {
		parts := strings.Split(multiaddr, "/")
		if parts[len(parts)-1] == relayIDused {
			relayAddr = multiaddr
			break
		}
	}

	return relayAddr
}

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

func AddRelayAddrToCSV(myAddr string, path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(myAddr + "\n")
	return err
}

func uploadRelayAddrToSheet(myAddr string) {
	payload := strings.NewReader(`{"addr":"` + myAddr + `"}`)
	resp, err := http.Post(sheetWebAppURL, "application/json", payload)
	if err != nil {
		fmt.Printf("[ERROR] Failed to upload relay address to sheet: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("[INFO] Uploaded relay address to sheet successfully")
}

// func fetchRelayAddrsFromSheet() []string {
// 	resp, err := http.Get(sheetWebAppURL)
// 	if err != nil {
// 		fmt.Printf("[ERROR] Failed to fetch relay addresses: %v\n", err)
// 		return nil
// 	}
// 	defer resp.Body.Close()
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		fmt.Printf("[ERROR] Failed to read response: %v\n", err)
// 		return nil
// 	}

// 	var addrs []string
// 	err = json.Unmarshal(body, &addrs)
// 	if err != nil {
// 		fmt.Printf("[ERROR] Failed to parse address list: %v\n", err)
// 		return nil
// 	}
// 	fmt.Println("[INFO] Relay address list fetched from sheet")
// 	return addrs
// }

func deleteRelayAddrFromSheet(myAddr string) {
	reqBody := strings.NewReader(`{"delete":"` + myAddr + `"}`)

	client := &http.Client{}
	req, err := http.NewRequest("POST", sheetWebAppURL, reqBody)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create delete request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] Failed to delete relay address from sheet: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("[INFO] Deleted relay address from sheet successfully")
}

// func fetchRelayAddrsFromSheet() ([]string, error) {
// 	csvURL := "https://raw.githubusercontent.com/cherry-aggarwal/LIBR/refs/heads/integration/docs/network.csv"
// 	resp, err := http.Get(csvURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch CSV: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	reader := csv.NewReader(resp.Body)

// 	// Skip header
// 	if _, err := reader.Read(); err != nil {
// 		return nil, fmt.Errorf("failed to read header: %w", err)
// 	}

// 	var relayAddrs []string

// 	for {
// 		row, err := reader.Read()
// 		if err != nil {
// 			if err.Error() == "EOF" {
// 				break
// 			}
// 			log.Printf("skipping bad row: %v", err)
// 			continue
// 		}

// 		if len(row) < 1 {
// 			log.Printf("skipping row with too few columns: %v", row)
// 			continue
// 		}

// 		relayAddrs = append(relayAddrs, row[0])
// 	}

// 	if len(relayAddrs) == 0 {
// 		return nil, fmt.Errorf("no valid address found")
// 	}

// 	return relayAddrs, nil
// }

func fetchRelayAddrsFromSheet() ([]string, error) {
	relayGID := "1789680527"
	rows, err := fetchRawData(relayGID)
	if err != nil {
		return nil, err
	}

	var relayList []string
	for _, row := range rows {
		if len(row) >= 1 {
			addr := strings.TrimSpace(row[0])
			// Only include addresses that start with '/'
			if strings.HasPrefix(addr, "/") {
				relayList = append(relayList, addr)
			}
		}
	}
	return relayList, nil
}

func fetchRawData(gid string) ([][]string, error) {
	url := fmt.Sprintf("https://docs.google.com/spreadsheets/d/e/2PACX-1vRDDE0x6LttdW13zLUwodMcVBsqk8fpnUsv-5SIJifZKWRehFpSKuJZawhswGMHSI2fZJDuENQ8SX1v/pub?output=csv&gid=%s", gid)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invalid CSV: %w", err)
	}

	if len(records) <= 1 {
		return nil, fmt.Errorf("no data rows in sheet")
	}

	return records[1:], nil // :point_left: skip the header row
}