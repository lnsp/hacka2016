package main

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"time"
)

const (
	TIMESTAMP_FORMAT     = "20060102150405"
	STATUS_INVALID_TOKEN = "Invalid authentication token"
)

// Generate a new access token.
func generateToken(device string) string {
	timestamp := time.Now().Format(TIMESTAMP_FORMAT)
	sha := sha1.New()
	sha.Write([]byte(timestamp))
	sha.Write([]byte(device))
	return hex.EncodeToString(sha.Sum(nil))
}

// Generate a new SSID
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

// Get hotspot by request token
func getHotspot(r *http.Request) (Hotspot, error) {
	var hotspot Hotspot

	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 {
		return hotspot, invalidTokenError
	}
	token := accessTokens[0]

	database.First(&hotspot, "token = ?", token)
	if hotspot.Token != token {
		return hotspot, invalidTokenError
	}

	return hotspot, nil
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
