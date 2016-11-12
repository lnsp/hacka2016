package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func updateHotspot(hotspot *Hotspot) (string, string, string) {
	ssid := generateSSID(hotspot.Session)
	database.Model(hotspot).Update("Session", ssid)

	var name, color string
	if hotspot.Conqueror > 0 {
		var conqueror Profile
		database.First(&conqueror, "ID = ?", hotspot.Conqueror)
		name = conqueror.Name
		color = conqueror.Color
	} else {
		name = "Unknown"
		color = "FF3400"
	}

	return ssid, name, color
}

func captureHotspot(hotspot Hotspot, id uint) bool {
	nextCapture := time.Now().Add(-time.Minute)
	if nextCapture.After(hotspot.LastCapture) {
		database.Model(&hotspot).Update("LastCapture", time.Now())
		database.Model(&hotspot).Update("Conqueror", id)
		return true
	}

	return false
}

func createHotspot() *Hotspot {
	token := generateSSID(ULTIMATE_KEY)
	hotspot := &Hotspot{
		Token:       token,
		Session:     generateSSID(token + ULTIMATE_KEY),
		LastCapture: time.Now(),
		Conqueror:   0,
	}
	database.Create(hotspot)
	return hotspot
}

func setupHotspotHandler(w http.ResponseWriter, r *http.Request) {
	secrets, ok := r.URL.Query()["secret"]
	if !ok || len(secrets) != 1 || secrets[0] != ULTIMATE_KEY {
		http.Error(w, "bad ultimate power", http.StatusUnauthorized)
		return
	}

	hotspot := createHotspot()
	data, err := json.Marshal(struct {
		Token string `json:"token"`
		SSID  string `json:"ssid"`
	}{
		Token: hotspot.Token,
		SSID:  hotspot.Session,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func captureHotspotHandler(w http.ResponseWriter, r *http.Request) {
	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 || validate(accessTokens[0]) == nil {
		http.Error(w, "invalid access token", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := getID(accessTokens[0])
	ssid := vars["ssid"]

	if id == 0 {
		http.Error(w, "invalid user", http.StatusUnauthorized)
		return
	}

	var hotspot Hotspot
	database.First(&hotspot, "Session = ?", ssid)
	if hotspot.Session != ssid {
		http.Error(w, "invalid hotspot ssid", http.StatusInternalServerError)
		return
	}

	success := captureHotspot(hotspot, id)
	data, err := json.Marshal(struct {
		Success bool `json:"success"`
	}{
		Success: success,
	})
	if err != nil {
		http.Error(w, "failed json parsing", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func updateHotspotHandler(w http.ResponseWriter, r *http.Request) {
	tokens, ok := r.URL.Query()["token"]
	if !ok || len(tokens) != 1 {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	token := tokens[0]

	hotspot := validateHotspot(token)
	if hotspot == nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	ssid, name, color := updateHotspot(hotspot)
	log.Println("New SSID for Hotspot: " + ssid)
	data, err := json.Marshal(struct {
		SSID    string `json:"ssid"`
		Name    string `json:"name"`
		Color   string `json:"color"`
		Capture int64  `json:"capture"`
	}{
		SSID:    ssid,
		Name:    name,
		Color:   color,
		Capture: int64(hotspot.LastCapture.Unix()),
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}
