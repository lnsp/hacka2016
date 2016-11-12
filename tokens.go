package main

import (
	"crypto/sha1"
	"encoding/hex"
	"time"
)

func validateHotspot(token string) *Hotspot {
	var hotspot Hotspot
	database.First(&hotspot, "token = ?", token)

	if hotspot.Token != token {
		return nil
	}

	return &hotspot
}

// Validate an access token.
func validate(accessToken string) *Account {
	var account Account
	database.First(&account, "Token = ?", accessToken)
	if account.ID == 0 {
		return nil
	}
	return &account
}

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
	return "honeypot" + str[:24]
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
