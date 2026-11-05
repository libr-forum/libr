package util

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/libr-forum/Libr/core/crypto/cryptoutils"
	cache "github.com/libr-forum/Libr/core/mod_client/cache_handler"
	"github.com/libr-forum/Libr/core/mod_client/types"
)

var (
	serverURL  string
	httpClient = &http.Client{Timeout: 10 * time.Second}
)

// InitServerClient sets the base URL of the discovery server.
// Call this once during app startup in place of SetupMongo.
func InitServerClient(baseURL string) {
	serverURL = strings.TrimRight(baseURL, "/")
}

// ── Read endpoints (no auth) ──────────────────────────────────────────────────

// GetOnlineMods returns a list of currently live moderators from the server.
// Only entries whose lastSeen is within the server-side TTL are returned.
// Fallback: If the server is unreachable, it returns the last known good list from cache.
func GetOnlineMods() ([]types.Mod, error) {
	resp, err := httpClient.Get(serverURL + "/mods")
	if err != nil {
		fmt.Printf("⚠️ GetOnlineMods: request failed: %v. Falling back to cache.\n", err)
		return cache.GetModsCache()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("⚠️ GetOnlineMods: server returned %d. Falling back to cache.\n", resp.StatusCode)
		return cache.GetModsCache()
	}

	var docs []struct {
		PeerId    string `json:"peerId"`
		PublicKey string `json:"publicKey"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		fmt.Printf("⚠️ GetOnlineMods: decode failed: %v. Falling back to cache.\n", err)
		return cache.GetModsCache()
	}

	mods := make([]types.Mod, 0, len(docs))
	for _, d := range docs {
		mods = append(mods, types.Mod{PeerId: d.PeerId, PublicKey: d.PublicKey})
	}

	// Update cache with fresh data
	if len(mods) > 0 {
		if err := cache.SaveModsCache(mods); err != nil {
			fmt.Printf("⚠️ Failed to update mods cache: %v\n", err)
		}
	}

	return mods, nil
}

// IsModAllowed checks whether the given public key is in the mod allowlist
// (the server-side 'mods' collection managed by admins).
// This is used at startup to decide if this node is a moderator, before
// it has registered its online presence.
func IsModAllowed(pubKeyB64 string) (bool, error) {
	resp, err := httpClient.Get(serverURL + "/mods/check?publicKey=" + url.QueryEscape(pubKeyB64))
	if err != nil {
		return false, fmt.Errorf("IsModAllowed: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("IsModAllowed: server returned %d", resp.StatusCode)
	}

	var result struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("IsModAllowed: decode failed: %w", err)
	}
	return result.Allowed, nil
}

// GetRelayAddr returns live relay multiaddresses from the server.
func GetRelayAddr() ([]string, error) {
	resp, err := httpClient.Get(serverURL + "/relays")
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

// GetStartNodes returns the list of bootstrap DB nodes from the server.
func GetStartNodes() ([]*types.Node, error) {
	resp, err := httpClient.Get(serverURL + "/nodes")
	if err != nil {
		return nil, fmt.Errorf("GetStartNodes: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetStartNodes: server returned %d", resp.StatusCode)
	}

	var docs []struct {
		NodeId string `json:"node_id"`
		PeerId string `json:"peer_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&docs); err != nil {
		return nil, fmt.Errorf("GetStartNodes: decode failed: %w", err)
	}

	nodes := make([]*types.Node, 0, len(docs))
	for _, d := range docs {
		nodeId, err := DecodeNodeID(d.NodeId)
		if err != nil {
			continue
		}
		nodes = append(nodes, &types.Node{NodeId: nodeId, PeerId: d.PeerId})
	}
	return nodes, nil
}

// ── Authenticated write helpers ───────────────────────────────────────────────

// getChallenge fetches a one-time nonce from the server for the given public key.
func getChallenge(pubKeyB64 string) (string, error) {
	endpoint := fmt.Sprintf("%s/auth/challenge?publicKey=%s", serverURL, url.QueryEscape(pubKeyB64))
	resp, err := httpClient.Get(endpoint)
	if err != nil {
		return "", fmt.Errorf("getChallenge: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("getChallenge: server returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Nonce string `json:"nonce"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("getChallenge: decode failed: %w", err)
	}
	return result.Nonce, nil
}

// signedPost performs get-challenge → sign → POST for authenticated write endpoints.
func signedPost(endpoint string, pubKeyB64 string, privKey ed25519.PrivateKey, extra map[string]string) error {
	nonce, err := getChallenge(pubKeyB64)
	if err != nil {
		return err
	}

	_, sigB64, err := cryptoutils.SignMessage(privKey, nonce)
	if err != nil {
		return fmt.Errorf("signedPost: signing failed: %w", err)
	}

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
		return fmt.Errorf("signedPost: marshal failed: %w", err)
	}

	resp, err := httpClient.Post(serverURL+endpoint, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("signedPost %s: request failed: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("signedPost %s: server returned %d: %s", endpoint, resp.StatusCode, string(b))
	}
	return nil
}

// ── Mod write operations ──────────────────────────────────────────────────────

// RegisterAsMod registers this node as an online moderator in the discovery server.
func RegisterAsMod(peerId, pubKeyB64 string, privKey ed25519.PrivateKey) error {
	return signedPost("/mods/register", pubKeyB64, privKey, map[string]string{"peerId": peerId})
}

// RefreshModPresence updates lastSeen for this mod entry, keeping it alive.
// Call periodically (e.g. every TTL/2 seconds) from the refresh cron.
func RefreshModPresence(pubKeyB64 string, privKey ed25519.PrivateKey) error {
	return signedPost("/mods/refresh", pubKeyB64, privKey, nil)
}

// DeregisterAsMod removes this mod entry immediately (best-effort on shutdown).
// The entry will expire naturally after the TTL anyway.
func DeregisterAsMod(pubKeyB64 string, privKey ed25519.PrivateKey) error {
	return signedPost("/mods/deregister", pubKeyB64, privKey, nil)
}
