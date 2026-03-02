package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	wsWriteWait      = 10 * time.Second
	wsPongWait       = 60 * time.Second
	wsPingPeriod     = (wsPongWait * 9) / 10
	wsReadLimit      = int64(1 << 20)
	relayRetryPeriod = 5 * time.Second
	ledgerSyncPeriod = 30 * time.Second
	sendQueueSize    = 256
)

type Envelope struct {
	Type          string          `json:"type"`
	RequestID     string          `json:"request_id"`
	OriginRelayID string          `json:"origin_relay_id"`
	TargetDBID    string          `json:"target_db_id"`
	ClientPeerID  string          `json:"client_peer_id"`
	Payload       json.RawMessage `json:"payload"`
}

type DBRegisterMessage struct {
	Type     string `json:"type"`
	DBPeerID string `json:"db_peer_id"`
}

type RelayRegisterMessage struct {
	Type        string `json:"type"`
	RelayPeerID string `json:"relay_peer_id"`
}

type HTTPRequest struct {
	ClientPeerID string          `json:"client_peer_id"`
	TargetDBID   string          `json:"target_db_id"`
	Payload      json.RawMessage `json:"payload"`
}

type LedgerRelay struct {
	PeerID    string `json:"peer_id"`
	WSAddress string `json:"ws_address"`
}

type wsPeer struct {
	conn      *websocket.Conn
	send      chan []byte
	done      chan struct{}
	server    *RelayServer
	peerID    string
	onMessage func([]byte)
	onClose   func()
	closed    atomic.Bool
}

type DBConn struct {
	peer *wsPeer
}

type RelayConn struct {
	peer      *wsPeer
	outbound  bool
	wsAddress string
}

type RelayServer struct {
	peerID        string
	port          int
	ledgerBaseURL string
	wsAddress     string

	httpServer *http.Server
	httpClient *http.Client

	upgrader websocket.Upgrader

	connectedDBs map[string]*DBConn
	dbMu         sync.RWMutex

	connectedRelays map[string]*RelayConn
	relayMu         sync.RWMutex

	pendingRequests map[string]chan Envelope
	pendingMu       sync.RWMutex

	knownRelays map[string]string
	knownMu     sync.RWMutex

	connectorRunning map[string]bool
	connectorMu      sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func main() {
	relayPeerID := strings.TrimSpace(os.Getenv("RELAY_PEER_ID"))
	if relayPeerID == "" {
		relayPeerID = GeneratePeerID()
	}

	port := 8080
	if raw := strings.TrimSpace(os.Getenv("PORT")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 65535 {
			log.Fatalf("invalid PORT %q", raw)
		}
		port = parsed
	}

	fmt.Print("Starting relay server with peer ID: ", relayPeerID, " on port: ", port, "\n")

	ledgerBaseURL := strings.TrimSpace(os.Getenv("LEDGER_BASE_URL"))
	if ledgerBaseURL == "" {
		ledgerBaseURL = "http://127.0.0.1:9000"
		log.Printf("LEDGER_BASE_URL not set, defaulting to %s", ledgerBaseURL)
	}

	server, err := NewRelayServer(relayPeerID, port, ledgerBaseURL)
	if err != nil {
		log.Fatalf("failed to create relay server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("failed to start relay server: %v", err)
	}
}

func NewRelayServer(peerID string, port int, ledgerBaseURL string) (*RelayServer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	wsAddress := fmt.Sprintf("ws://%s:%d/ws/relay", localAdvertiseHost(), port)

	srv := &RelayServer{
		peerID:           peerID,
		port:             port,
		ledgerBaseURL:    strings.TrimRight(ledgerBaseURL, "/"),
		wsAddress:        wsAddress,
		httpClient:       &http.Client{Timeout: 10 * time.Second},
		connectedDBs:     make(map[string]*DBConn),
		connectedRelays:  make(map[string]*RelayConn),
		pendingRequests:  make(map[string]chan Envelope),
		knownRelays:      make(map[string]string),
		connectorRunning: make(map[string]bool),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}

	srv.knownRelays[peerID] = wsAddress

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/db", srv.handleDBWebSocket)
	mux.HandleFunc("/ws/relay", srv.handleRelayWebSocket)
	mux.HandleFunc("/request", srv.handleHTTPRequest)
	mux.HandleFunc("/healthz", srv.handleHealth)

	srv.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return srv, nil
}

func (s *RelayServer) Start() error {
	log.Printf("relay starting peer_id=%s port=%d ws_address=%s", s.peerID, s.port, s.wsAddress)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("http server error: %v", err)
			s.cancel()
		}
	}()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.ledgerLoop()
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-sigCtx.Done():
		log.Printf("shutdown signal received")
	case <-s.ctx.Done():
		log.Printf("relay context canceled")
	}

	s.Stop()
	return nil
}

func (s *RelayServer) Stop() {
	s.cancel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}

	s.closeAllConnections()
	s.wg.Wait()
	log.Printf("relay stopped")
}

func (s *RelayServer) closeAllConnections() {
	s.dbMu.Lock()
	dbs := make([]*DBConn, 0, len(s.connectedDBs))
	for _, conn := range s.connectedDBs {
		dbs = append(dbs, conn)
	}
	s.dbMu.Unlock()

	s.relayMu.Lock()
	relays := make([]*RelayConn, 0, len(s.connectedRelays))
	for _, conn := range s.connectedRelays {
		relays = append(relays, conn)
	}
	s.relayMu.Unlock()

	for _, db := range dbs {
		db.peer.close()
	}
	for _, relay := range relays {
		relay.peer.close()
	}
}

func (s *RelayServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *RelayServer) handleDBWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("db upgrade failed: %v", err)
		return
	}

	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		_ = conn.Close()
		return
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Printf("db register read failed: %v", err)
		_ = conn.Close()
		return
	}

	var reg DBRegisterMessage
	if err := json.Unmarshal(msg, &reg); err != nil || reg.Type != "REGISTER_DB" || reg.DBPeerID == "" {
		log.Printf("invalid db register message: %s", string(msg))
		_ = conn.Close()
		return
	}

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		_ = conn.Close()
		return
	}

	peer := &wsPeer{
		conn:   conn,
		send:   make(chan []byte, sendQueueSize),
		done:   make(chan struct{}),
		server: s,
		peerID: reg.DBPeerID,
	}

	dbConn := &DBConn{peer: peer}

	peer.onMessage = func(raw []byte) {
		s.handleDBMessage(reg.DBPeerID, raw)
	}
	peer.onClose = func() {
		s.dbMu.Lock()
		existing, ok := s.connectedDBs[reg.DBPeerID]
		if ok && existing == dbConn {
			delete(s.connectedDBs, reg.DBPeerID)
		}
		s.dbMu.Unlock()
		log.Printf("db disconnected db_peer_id=%s", reg.DBPeerID)
	}

	s.dbMu.Lock()
	old := s.connectedDBs[reg.DBPeerID]
	s.connectedDBs[reg.DBPeerID] = dbConn
	s.dbMu.Unlock()
	if old != nil {
		old.peer.close()
	}

	log.Printf("db connected db_peer_id=%s", reg.DBPeerID)
	s.startPeerLoops(peer)
}

func (s *RelayServer) handleRelayWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("relay upgrade failed: %v", err)
		return
	}

	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		_ = conn.Close()
		return
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Printf("relay register read failed: %v", err)
		_ = conn.Close()
		return
	}

	var reg RelayRegisterMessage
	if err := json.Unmarshal(msg, &reg); err != nil || reg.Type != "REGISTER_RELAY" || reg.RelayPeerID == "" {
		log.Printf("invalid relay register message: %s", string(msg))
		_ = conn.Close()
		return
	}

	if reg.RelayPeerID == s.peerID {
		log.Printf("self relay connection rejected peer_id=%s", reg.RelayPeerID)
		_ = conn.Close()
		return
	}

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		_ = conn.Close()
		return
	}

	relayConn := &RelayConn{
		peer: &wsPeer{
			conn:   conn,
			send:   make(chan []byte, sendQueueSize),
			done:   make(chan struct{}),
			server: s,
			peerID: reg.RelayPeerID,
		},
		outbound:  false,
		wsAddress: "",
	}

	if !s.registerRelayConn(reg.RelayPeerID, relayConn) {
		_ = conn.Close()
		return
	}

	relayConn.peer.onMessage = func(raw []byte) {
		s.handleRelayMessage(reg.RelayPeerID, raw)
	}
	relayConn.peer.onClose = func() {
		s.unregisterRelayConn(reg.RelayPeerID, relayConn)
		log.Printf("relay disconnected relay_peer_id=%s inbound=true", reg.RelayPeerID)
	}

	s.noteKnownRelay(reg.RelayPeerID, "")

	s.startPeerLoops(relayConn.peer)

	register := RelayRegisterMessage{Type: "REGISTER_RELAY", RelayPeerID: s.peerID}
	_ = s.enqueueJSON(relayConn.peer, register)

	log.Printf("relay connected relay_peer_id=%s inbound=true", reg.RelayPeerID)
}

func (s *RelayServer) startPeerLoops(p *wsPeer) {
	s.wg.Add(2)
	go func() {
		defer s.wg.Done()
		p.readPump()
	}()
	go func() {
		defer s.wg.Done()
		p.writePump()
	}()
}

func (p *wsPeer) readPump() {
	defer p.close()

	p.conn.SetReadLimit(wsReadLimit)
	_ = p.conn.SetReadDeadline(time.Now().Add(wsPongWait))
	p.conn.SetPongHandler(func(_ string) error {
		return p.conn.SetReadDeadline(time.Now().Add(wsPongWait))
	})

	for {
		_, message, err := p.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws read error peer_id=%s err=%v", p.peerID, err)
			}
			return
		}
		if p.onMessage != nil {
			p.onMessage(message)
		}
	}
}

func (p *wsPeer) writePump() {
	ticker := time.NewTicker(wsPingPeriod)
	defer func() {
		ticker.Stop()
		p.close()
	}()

	for {
		select {
		case msg := <-p.send:
			if err := p.conn.SetWriteDeadline(time.Now().Add(wsWriteWait)); err != nil {
				return
			}
			if err := p.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			if err := p.conn.SetWriteDeadline(time.Now().Add(wsWriteWait)); err != nil {
				return
			}
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-p.server.ctx.Done():
			return
		case <-p.done:
			return
		}
	}
}

func (p *wsPeer) close() {
	if !p.closed.CompareAndSwap(false, true) {
		return
	}

	if p.onClose != nil {
		p.onClose()
	}

	close(p.done)
	_ = p.conn.Close()
}

func (s *RelayServer) handleDBMessage(dbPeerID string, raw []byte) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		log.Printf("invalid db message db_peer_id=%s err=%v", dbPeerID, err)
		return
	}

	if env.Type != "RESPONSE" {
		log.Printf("ignoring non-response from db db_peer_id=%s type=%s", dbPeerID, env.Type)
		return
	}

	s.processResponseEnvelope(env)
}

func (s *RelayServer) handleRelayMessage(fromRelayID string, raw []byte) {
	var kind struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &kind); err != nil {
		log.Printf("invalid relay message from=%s err=%v", fromRelayID, err)
		return
	}

	if kind.Type == "REGISTER_RELAY" {
		return
	}

	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		log.Printf("invalid envelope from relay=%s err=%v", fromRelayID, err)
		return
	}

	switch env.Type {
	case "REQUEST":
		s.processIncomingRequest(env)
	case "RESPONSE":
		s.processResponseEnvelope(env)
	default:
		log.Printf("unknown envelope type from relay=%s type=%s", fromRelayID, env.Type)
	}
}

func (s *RelayServer) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, wsReadLimit))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	var req HTTPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.ClientPeerID == "" || req.TargetDBID == "" {
		http.Error(w, "client_peer_id and target_db_id are required", http.StatusBadRequest)
		return
	}

	requestID := uuid.NewString()
	responseCh := make(chan Envelope, 1)

	s.pendingMu.Lock()
	s.pendingRequests[requestID] = responseCh
	s.pendingMu.Unlock()

	defer func() {
		s.pendingMu.Lock()
		delete(s.pendingRequests, requestID)
		s.pendingMu.Unlock()
	}()

	env := Envelope{
		Type:          "REQUEST",
		RequestID:     requestID,
		OriginRelayID: s.peerID,
		TargetDBID:    req.TargetDBID,
		ClientPeerID:  req.ClientPeerID,
		Payload:       req.Payload,
	}

	if err := s.routeRequest(env); err != nil {
		log.Printf("request route failed request_id=%s err=%v", requestID, err)
		http.Error(w, "failed to route request", http.StatusBadGateway)
		return
	}

	select {
	case resp := <-responseCh:
		writeJSON(w, http.StatusOK, resp)
	case <-time.After(5 * time.Second):
		http.Error(w, "request timeout", http.StatusGatewayTimeout)
	case <-s.ctx.Done():
		http.Error(w, "relay shutting down", http.StatusServiceUnavailable)
	}
}

func (s *RelayServer) routeRequest(env Envelope) error {
	nearest := s.FindNearestRelay(env.TargetDBID)
	log.Printf("routing request request_id=%s target_db=%s nearest_relay=%s self=%s", env.RequestID, env.TargetDBID, nearest, s.peerID)

	if nearest == s.peerID {
		return s.forwardToLocalDB(env)
	}
	return s.forwardToRelay(nearest, env)
}

func (s *RelayServer) processIncomingRequest(env Envelope) {
	if env.RequestID == "" || env.OriginRelayID == "" || env.TargetDBID == "" {
		log.Printf("dropping invalid request envelope request_id=%s origin=%s target=%s", env.RequestID, env.OriginRelayID, env.TargetDBID)
		return
	}

	if s.hasLocalDB(env.TargetDBID) {
		if err := s.forwardToLocalDB(env); err != nil {
			log.Printf("forward local db failed request_id=%s err=%v", env.RequestID, err)
			s.sendErrorResponse(env, "local-db-unavailable")
		}
		return
	}

	nearest := s.FindNearestRelay(env.TargetDBID)
	log.Printf("forwarding request request_id=%s target_db=%s nearest=%s", env.RequestID, env.TargetDBID, nearest)

	if nearest == s.peerID {
		s.sendErrorResponse(env, "target-db-not-found")
		return
	}

	if err := s.forwardToRelay(nearest, env); err != nil {
		log.Printf("forward relay failed request_id=%s relay=%s err=%v", env.RequestID, nearest, err)
		s.sendErrorResponse(env, "relay-forward-failed")
	}
}

func (s *RelayServer) processResponseEnvelope(env Envelope) {
	if env.RequestID == "" {
		log.Printf("dropping response with empty request_id")
		return
	}

	if env.OriginRelayID == s.peerID {
		s.pendingMu.RLock()
		ch, ok := s.pendingRequests[env.RequestID]
		s.pendingMu.RUnlock()
		if !ok {
			log.Printf("pending request not found request_id=%s", env.RequestID)
			return
		}

		select {
		case ch <- env:
		default:
			log.Printf("pending response channel full request_id=%s", env.RequestID)
		}
		return
	}

	if err := s.forwardToRelay(env.OriginRelayID, env); err != nil {
		log.Printf("response forward failed request_id=%s origin=%s err=%v", env.RequestID, env.OriginRelayID, err)
	}
}

func (s *RelayServer) sendErrorResponse(req Envelope, code string) {
	payload, _ := json.Marshal(map[string]string{"error": code})
	resp := Envelope{
		Type:          "RESPONSE",
		RequestID:     req.RequestID,
		OriginRelayID: req.OriginRelayID,
		TargetDBID:    req.TargetDBID,
		ClientPeerID:  req.ClientPeerID,
		Payload:       payload,
	}
	s.processResponseEnvelope(resp)
}

func (s *RelayServer) hasLocalDB(dbID string) bool {
	s.dbMu.RLock()
	_, ok := s.connectedDBs[dbID]
	s.dbMu.RUnlock()
	return ok
}

func (s *RelayServer) forwardToLocalDB(env Envelope) error {
	s.dbMu.RLock()
	dbConn, ok := s.connectedDBs[env.TargetDBID]
	s.dbMu.RUnlock()
	if !ok {
		return fmt.Errorf("db not connected: %s", env.TargetDBID)
	}

	return s.enqueueJSON(dbConn.peer, env)
}

func (s *RelayServer) forwardToRelay(relayPeerID string, env Envelope) error {
	if relayPeerID == s.peerID {
		return s.forwardToLocalDB(env)
	}

	s.relayMu.RLock()
	relayConn, ok := s.connectedRelays[relayPeerID]
	s.relayMu.RUnlock()
	if !ok {
		return fmt.Errorf("relay not connected: %s", relayPeerID)
	}

	return s.enqueueJSON(relayConn.peer, env)
}

func (s *RelayServer) enqueueJSON(peer *wsPeer, v any) error {
	msg, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if peer.closed.Load() {
		return fmt.Errorf("peer closed: %s", peer.peerID)
	}

	select {
	case peer.send <- msg:
		return nil
	default:
		return fmt.Errorf("send queue full for peer=%s", peer.peerID)
	}
}

func (s *RelayServer) registerRelayConn(relayPeerID string, conn *RelayConn) bool {
	var oldConn *RelayConn
	accepted := true

	s.relayMu.Lock()
	if relayPeerID == s.peerID {
		accepted = false
	}

	if accepted {
		if existing, ok := s.connectedRelays[relayPeerID]; ok {
			if existing != conn {
				// Keep outbound when an inbound duplicate arrives.
				if existing.outbound && !conn.outbound {
					accepted = false
				} else {
					s.connectedRelays[relayPeerID] = conn
					oldConn = existing
				}
			}
		} else {
			s.connectedRelays[relayPeerID] = conn
		}
	}
	s.relayMu.Unlock()

	if oldConn != nil {
		oldConn.peer.close()
	}

	return accepted
}

func (s *RelayServer) unregisterRelayConn(relayPeerID string, conn *RelayConn) {
	s.relayMu.Lock()
	existing, ok := s.connectedRelays[relayPeerID]
	if ok && existing == conn {
		delete(s.connectedRelays, relayPeerID)
	}
	s.relayMu.Unlock()
}

func (s *RelayServer) ledgerLoop() {
	ticker := time.NewTicker(ledgerSyncPeriod)
	defer ticker.Stop()

	for {
		if err := s.syncLedgerOnce(); err != nil {
			log.Printf("ledger sync error: %v", err)
		}

		select {
		case <-ticker.C:
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *RelayServer) syncLedgerOnce() error {
	errs := make([]string, 0, 2)

	if err := s.putSelfToLedger(); err != nil {
		errs = append(errs, fmt.Sprintf("put self: %v", err))
	}

	relays, err := s.fetchRelaysFromLedger()
	if err != nil {
		errs = append(errs, fmt.Sprintf("get relays: %v", err))
		return errors.New(strings.Join(errs, "; "))
	}

	for _, relay := range relays {
		if relay.PeerID == "" {
			continue
		}

		s.noteKnownRelay(relay.PeerID, relay.WSAddress)

		if relay.PeerID == s.peerID {
			continue
		}

		s.ensureRelayConnector(relay.PeerID)
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func (s *RelayServer) noteKnownRelay(peerID, wsAddress string) {
	s.knownMu.Lock()
	defer s.knownMu.Unlock()

	if peerID == s.peerID {
		s.knownRelays[peerID] = s.wsAddress
		return
	}

	if wsAddress == "" {
		if _, exists := s.knownRelays[peerID]; !exists {
			s.knownRelays[peerID] = ""
		}
		return
	}

	s.knownRelays[peerID] = wsAddress
}

func (s *RelayServer) putSelfToLedger() error {
	payload := LedgerRelay{PeerID: s.peerID, WSAddress: s.wsAddress}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(s.ctx, http.MethodPut, s.ledgerBaseURL+"/relays", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("ledger put status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	return nil
}

func (s *RelayServer) fetchRelaysFromLedger() ([]LedgerRelay, error) {
	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, s.ledgerBaseURL+"/relays", nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("ledger get status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var relays []LedgerRelay
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&relays); err != nil {
		return nil, err
	}

	return relays, nil
}

func (s *RelayServer) ensureRelayConnector(peerID string) {
	s.connectorMu.Lock()
	if s.connectorRunning[peerID] {
		s.connectorMu.Unlock()
		return
	}
	s.connectorRunning[peerID] = true
	s.connectorMu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer func() {
			s.connectorMu.Lock()
			delete(s.connectorRunning, peerID)
			s.connectorMu.Unlock()
		}()

		ticker := time.NewTicker(relayRetryPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			default:
			}

			if s.isRelayConnected(peerID) {
				select {
				case <-ticker.C:
					continue
				case <-s.ctx.Done():
					return
				}
			}

			addr := s.getRelayAddress(peerID)
			if addr == "" {
				select {
				case <-ticker.C:
					continue
				case <-s.ctx.Done():
					return
				}
			}

			if err := s.connectToRelay(peerID, addr); err != nil {
				log.Printf("relay connect failed target=%s addr=%s err=%v", peerID, addr, err)
				select {
				case <-ticker.C:
				case <-s.ctx.Done():
					return
				}
				continue
			}

			select {
			case <-ticker.C:
			case <-s.ctx.Done():
				return
			}
		}
	}()
}

func (s *RelayServer) isRelayConnected(peerID string) bool {
	s.relayMu.RLock()
	_, ok := s.connectedRelays[peerID]
	s.relayMu.RUnlock()
	return ok
}

func (s *RelayServer) getRelayAddress(peerID string) string {
	s.knownMu.RLock()
	addr := s.knownRelays[peerID]
	s.knownMu.RUnlock()
	return addr
}

func (s *RelayServer) connectToRelay(peerID, wsAddress string) error {
	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, _, err := dialer.DialContext(s.ctx, wsAddress, nil)
	if err != nil {
		return err
	}

	register := RelayRegisterMessage{Type: "REGISTER_RELAY", RelayPeerID: s.peerID}
	if err := conn.SetWriteDeadline(time.Now().Add(wsWriteWait)); err != nil {
		_ = conn.Close()
		return err
	}
	if err := conn.WriteJSON(register); err != nil {
		_ = conn.Close()
		return err
	}

	relayConn := &RelayConn{
		peer: &wsPeer{
			conn:   conn,
			send:   make(chan []byte, sendQueueSize),
			done:   make(chan struct{}),
			server: s,
			peerID: peerID,
		},
		outbound:  true,
		wsAddress: wsAddress,
	}

	if !s.registerRelayConn(peerID, relayConn) {
		_ = conn.Close()
		return nil
	}

	relayConn.peer.onMessage = func(raw []byte) {
		s.handleRelayMessage(peerID, raw)
	}
	relayConn.peer.onClose = func() {
		s.unregisterRelayConn(peerID, relayConn)
		log.Printf("relay disconnected relay_peer_id=%s inbound=false", peerID)
	}

	s.startPeerLoops(relayConn.peer)
	log.Printf("relay connected relay_peer_id=%s inbound=false addr=%s", peerID, wsAddress)
	return nil
}

func (s *RelayServer) FindNearestRelay(targetID string) string {
	ids := make([]string, 0, 32)
	ids = append(ids, s.peerID)

	s.knownMu.RLock()
	for peerID := range s.knownRelays {
		if peerID != s.peerID {
			ids = append(ids, peerID)
		}
	}
	s.knownMu.RUnlock()

	nearest := s.peerID
	nearestDistance := XORDistance(s.peerID, targetID)

	for _, peerID := range ids {
		distance := XORDistance(peerID, targetID)
		cmp := distance.Cmp(nearestDistance)
		if cmp < 0 || (cmp == 0 && peerID < nearest) {
			nearest = peerID
			nearestDistance = distance
		}
	}

	return nearest
}

func GeneratePeerID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("failed to generate peer id: %v", err))
	}
	return hex.EncodeToString(buf)
}

func XORDistance(a, b string) *big.Int {
	ai := peerIDToBigInt(a)
	bi := peerIDToBigInt(b)
	return new(big.Int).Xor(ai, bi)
}

func peerIDToBigInt(id string) *big.Int {
	id = strings.TrimSpace(id)
	if id == "" {
		return big.NewInt(0)
	}

	if strings.HasPrefix(id, "0x") || strings.HasPrefix(id, "0X") {
		id = id[2:]
	}

	if isBinaryString(id) {
		if n, ok := new(big.Int).SetString(id, 2); ok {
			return n
		}
	}

	if n, ok := new(big.Int).SetString(id, 16); ok {
		return n
	}

	return new(big.Int).SetBytes([]byte(id))
}

func isBinaryString(v string) bool {
	if v == "" {
		return false
	}
	for i := 0; i < len(v); i++ {
		if v[i] != '0' && v[i] != '1' {
			return false
		}
	}
	return true
}

func localAdvertiseHost() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || localAddr.IP == nil {
		return "127.0.0.1"
	}

	return localAddr.IP.String()
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("http write json failed: %v", err)
	}
}
