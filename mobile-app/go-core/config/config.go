package config

import "os"

const K = 4
const Alpha = 4
const DeleteThreshold = 40.0

// ServerURL is the base URL of the librserver discovery server.
// Override with the SERVER_URL environment variable.
const ServerURL = "https://libr-relay007-1.onrender.com"

// RegistryRefreshSeconds is how often (in seconds) a mod refreshes its
// presence in the discovery server.
const RegistryRefreshSeconds = 90

// GetServerURL returns the discovery server URL, preferring the SERVER_URL
// environment variable over the compiled-in default.
func GetServerURL() string {
	if u := os.Getenv("SERVER_URL"); u != "" {
		return u
	}
	return ServerURL
}
