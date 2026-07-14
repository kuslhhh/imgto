package mcp

import (
	"crypto/subtle"
)

// Authenticate checks if the provided API key is valid.
// Returns true if authentication passes or if no API key is configured.
func (s *Server) Authenticate(providedKey string) bool {
	if s.cfg.APIKey == "" {
		// No API key configured — auth is disabled
		return true
	}
	if providedKey == "" {
		return false
	}
	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(s.cfg.APIKey), []byte(providedKey)) == 1
}
