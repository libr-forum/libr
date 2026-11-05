package util

// AmIMod checks whether myKey is in the admin-controlled mod allowlist.
// This intentionally checks the allowlist (mods collection), not the live
// registry (onlinemods), so it works correctly at startup before registration.
func AmIMod(myKey string) (bool, error) {
	return IsModAllowed(myKey)
}
