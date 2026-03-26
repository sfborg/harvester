package clb

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	tokenDir  = "sfborg/harvester"
	tokenFile = "clb_token.json"
	// Treat the token as expired this far before actual expiry
	// so requests don't fail mid-flight.
	expiryMargin = 60 * time.Minute
)

// cachedToken is the structure persisted to disk.
type cachedToken struct {
	API   string `json:"api"`
	User  string `json:"user"`
	Token string `json:"token"`
}

// tokenPath returns the platform-appropriate path for the token file.
func tokenPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("finding config dir: %w", err)
	}
	return filepath.Join(cfgDir, tokenDir, tokenFile), nil
}

// HasValidToken returns true if a non-expired cached token exists
// for the given API endpoint.
func HasValidToken(api string) bool {
	path, err := tokenPath()
	if err != nil {
		return false
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var ct cachedToken
	if err := json.Unmarshal(raw, &ct); err != nil {
		return false
	}

	if ct.API != api || ct.Token == "" {
		return false
	}

	exp, err := jwtExpiry(ct.Token)
	if err != nil {
		return false
	}

	return time.Now().Before(exp.Add(-expiryMargin))
}

// loadCachedToken reads a cached token from disk and returns it if
// it matches the current API/user and has not expired.
func loadCachedToken(api, user string) (string, bool) {
	path, err := tokenPath()
	if err != nil {
		return "", false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	var ct cachedToken
	if err := json.Unmarshal(data, &ct); err != nil {
		return "", false
	}

	// Token must be for the same API. If user is empty (no
	// credentials provided), accept any cached user.
	if ct.API != api || ct.Token == "" {
		return "", false
	}
	if user != "" && ct.User != user {
		return "", false
	}

	// Check JWT expiry.
	exp, err := jwtExpiry(ct.Token)
	if err != nil {
		slog.Debug("cannot parse cached token expiry", "error", err)
		return "", false
	}

	if time.Now().After(exp.Add(-expiryMargin)) {
		slog.Info("cached CLB token has expired")
		return "", false
	}

	slog.Info(
		"using cached CLB token",
		"expires", exp.Format(time.RFC3339),
	)
	return ct.Token, true
}

// saveCachedToken writes the token to disk with 0600 permissions.
func saveCachedToken(api, user, token string) {
	path, err := tokenPath()
	if err != nil {
		slog.Warn("cannot determine token path", "error", err)
		return
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		slog.Warn("cannot create token dir", "error", err)
		return
	}

	ct := cachedToken{API: api, User: user, Token: token}
	data, err := json.Marshal(ct)
	if err != nil {
		slog.Warn("cannot marshal token", "error", err)
		return
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		slog.Warn("cannot write token file", "error", err)
		return
	}

	slog.Info("cached CLB token", "path", path)
}

// jwtExpiry extracts the "exp" claim from a JWT without verifying
// the signature. Returns the expiry time.
func jwtExpiry(token string) (time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("not a valid JWT")
	}

	// Decode the payload (second part). JWT uses unpadded base64url.
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("decoding JWT payload: %w", err)
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return time.Time{}, fmt.Errorf("parsing JWT claims: %w", err)
	}

	if claims.Exp == 0 {
		return time.Time{}, fmt.Errorf("JWT has no exp claim")
	}

	return time.Unix(claims.Exp, 0), nil
}
