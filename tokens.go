package main

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"time"
)

const (
	STATUS_INVALID_TOKEN = "Invalid authentication token"
)

// Generate a new access token.
func generateToken(device string) string {
	timestamp := time.Now().Format("20060102150405")
	sha := sha1.New()
	sha.Write([]byte(timestamp))
	sha.Write([]byte(device))
	return hex.EncodeToString(sha.Sum(nil))
}

func generateSSID(active string) string {
	str := generateToken(active)
	return DEFAULT_SSID_PREFIX + str[:MAX_UNIQ_SSID_LEN]
}

// Validate a user request
func validateRequest(r *http.Request) (string, error) {
	var account Account

	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 {
		return "", invalidTokenError
	}

	token := accessTokens[0]
	database.First(&account, SQL_FIND_ACCOUNT_BY_TOKEN, token)

	if token != account.Token && token == "" {
		return "", invalidTokenError
	}

	return account.Token, nil
}

func validateHotspot(token string) *Hotspot {
	var hotspot Hotspot
	database.First(&hotspot, "token = ?", token)

	if hotspot.Token != token {
		return nil
	}

	return &hotspot
}

// Create a new access token or retrieve an existing.
func createAccessToken(device string) string {
	// Look for existing device
	var account Account
	var token string
	database.First(&account, "Device = ?", device)

	if account.Token == "" {
		token = generateToken(device)
	}

	return token
}
