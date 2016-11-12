package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/kellydunn/golang-geo"
)

func updatePosition(id uint, latitude, longitude float64) {
	position := &Position{
		Source:    id,
		Latitude:  latitude,
		Longitude: longitude,
		Date:      time.Now(),
	}
	database.Create(position)
}

type nearbyEntry struct {
	ID       uint      `json:"id"`
	Distance float64   `json:"distance"`
	Date     time.Time `json:"date"`
	Device   string    `json:"device"`
}

// Retrieve all nearby entities.
func getNearby(id uint, latitude, longitude float64) []nearbyEntry {
	lastHour := time.Now().Add(-time.Minute)

	positions := []Position{}
	database.Where("date > ?", lastHour).Order("date desc").Find(&positions)

	sourcePoint := geo.NewPoint(latitude, longitude)
	entries := []nearbyEntry{}

	dates := make(map[uint]time.Time)

	for _, element := range positions {
		if element.Source == id {
			continue
		}

		if existing, ok := dates[element.Source]; ok {
			if existing.After(element.Date) {
				continue
			} else {
				dates[element.Source] = element.Date
			}
		} else {
			dates[element.Source] = element.Date
		}

		nearbyPoint := geo.NewPoint(element.Latitude, element.Longitude)
		distance := sourcePoint.GreatCircleDistance(nearbyPoint)
		if distance < MAX_DISTANCE {
			var account Account
			var profile Profile

			database.Where("ID = ?", element.Source).First(&profile)
			database.Where("ID = ?", profile.AccountID).First(&account)

			entries = append(entries, nearbyEntry{
				ID:       element.Source,
				Date:     element.Date,
				Distance: distance,
				Device:   account.Device,
			})
		}
	}

	return entries
}

// Handle nearby
func nearbyHandler(w http.ResponseWriter, r *http.Request) {
	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 || validate(accessTokens[0]) == nil {
		http.Error(w, "invalid access token", http.StatusUnauthorized)
		return
	}
	token := accessTokens[0]

	vars := mux.Vars(r)
	latitude, err := strconv.ParseFloat(vars["latitude"], 64)
	if err != nil {
		http.Error(w, "invalid latitude", http.StatusInternalServerError)
		return
	}
	longitude, err := strconv.ParseFloat(vars["longitude"], 64)
	if err != nil {
		http.Error(w, "invalid longitude", http.StatusInternalServerError)
		return
	}

	id := getID(token)
	updatePosition(id, latitude, longitude)
	nearby := getNearby(id, latitude, longitude)

	data, err := json.Marshal(struct {
		Nearby []nearbyEntry `json:"nearby"`
	}{
		Nearby: nearby,
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
