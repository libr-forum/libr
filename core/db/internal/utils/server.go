package utils

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/libr-forum/Libr/core/db/internal/models"
	"github.com/libr-forum/Libr/core/db/internal/node"
)

var (
	dbServerURL  string
	dbHTTPClient = &http.Client{Timeout: 10 * time.Second}
)

// InitServerClient sets the base URL of the librserver discovery server.
// Call this once at startup in place of SetupMongo.
func InitServerClient(baseURL string) {
	dbServerURL = strings.TrimRight(baseURL, "/")
}

// GetDbAddr fetches bootstrap DB node addresses from the discovery server.
func GetDbAddr() ([]*models.Node, error) {
	resp, err := dbHTTPClient.Get(dbServerURL + "/nodes")
	if err != nil {
		return nil, fmt.Errorf("GetDbAddr: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetDbAddr: server returned %d", resp.StatusCode)
	}

	var docs []struct {
		NodeId string `json:"node_id"`
		PeerId string `json:"peer_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		return nil, fmt.Errorf("GetDbAddr: decode failed: %w", err)
	}

	var nodeList []*models.Node
	for _, d := range docs {
		nodeId, err := node.DecodeNodeID(d.NodeId)
		if err != nil {
			continue
		}
		nodeList = append(nodeList, &models.Node{NodeId: nodeId, PeerId: d.PeerId})
	}
	return nodeList, nil
}

// GetOnlineMods fetches currently live moderators from the discovery server.
// Only their PublicKey is used by the DB node for signature validation.
func GetOnlineMods() ([]*models.Mod, error) {
	resp, err := dbHTTPClient.Get(dbServerURL + "/mods")
	if err != nil {
		return nil, fmt.Errorf("GetOnlineMods: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetOnlineMods: server returned %d", resp.StatusCode)
	}

	var docs []struct {
		PeerId    string `json:"peerId"`
		PublicKey string `json:"publicKey"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		return nil, fmt.Errorf("GetOnlineMods: decode failed: %w", err)
	}

	mods := make([]*models.Mod, 0, len(docs))
	for _, d := range docs {
		mods = append(mods, &models.Mod{
			PublicKey: d.PublicKey,
			// IP and Port are not provided by the server and not used by the DB node
		})
	}
	return mods, nil
}

// GetRelayAddr fetches live relay multiaddresses from the discovery server.
func GetRelayAddr() ([]string, error) {
	resp, err := dbHTTPClient.Get(dbServerURL + "/relays")
	if err != nil {
		return nil, fmt.Errorf("GetRelayAddr: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetRelayAddr: server returned %d", resp.StatusCode)
	}

	var docs []struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		return nil, fmt.Errorf("GetRelayAddr: decode failed: %w", err)
	}

	var addrs []string
	for _, d := range docs {
		if strings.HasPrefix(d.Address, "/") {
			addrs = append(addrs, strings.TrimSpace(d.Address))
		}
	}
	return addrs, nil
}

// ── Authenticated write helpers ───────────────────────────────────────────────

func getChallenge(pubKeyB64 string) (string, error) {
	endpoint := fmt.Sprintf("%s/auth/challenge?publicKey=%s", dbServerURL, url.QueryEscape(pubKeyB64))
	resp, err := dbHTTPClient.Get(endpoint)
	if err != nil {
		return "", fmt.Errorf("getChallenge: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("getChallenge: server returned %d: %s", resp.StatusCode, body)
	}
	var result struct {
		Nonce string `json:"nonce"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("getChallenge: decode: %w", err)
	}
	return result.Nonce, nil
}

func signedPost(endpoint string, pubKeyB64 string, privKey ed25519.PrivateKey, extra map[string]string) error {
	nonce, err := getChallenge(pubKeyB64)
	if err != nil {
		return err
	}
	sig := ed25519.Sign(privKey, []byte(nonce))
	sigB64 := base64.StdEncoding.EncodeToString(sig)

	body := map[string]string{
		"publicKey": pubKeyB64,
		"nonce":     nonce,
		"signature": sigB64,
	}
	for k, v := range extra {
		body[k] = v
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("signedPost: marshal: %w", err)
	}
	resp, err := dbHTTPClient.Post(dbServerURL+endpoint, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("signedPost %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("signedPost %s: server returned %d: %s", endpoint, resp.StatusCode, b)
	}
	return nil
}

// RegisterAsNode registers this DB node with the discovery server.
// nodeId should be base64-encoded. pubKeyB64 is the base64 public key used for auth.
func RegisterAsNode(nodeId, peerId, pubKeyB64 string, privKey ed25519.PrivateKey) error {
	return signedPost("/nodes/register", pubKeyB64, privKey, map[string]string{
		"nodeId": nodeId,
		"peerId": peerId,
	})
}

// RefreshNodePresence refreshes the lastSeen timestamp for this node entry.
func RefreshNodePresence(pubKeyB64 string, privKey ed25519.PrivateKey) error {
	return signedPost("/nodes/refresh", pubKeyB64, privKey, nil)
}

// DeregisterAsNode removes this node entry immediately on shutdown.
func DeregisterAsNode(pubKeyB64 string, privKey ed25519.PrivateKey) error {
	return signedPost("/nodes/deregister", pubKeyB64, privKey, nil)
}
