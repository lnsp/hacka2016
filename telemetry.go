package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/kellydunn/golang-geo"
)

const (
	STATUS_INVALID_GPS = "Invalid GPS location"
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
	database.Where(SQL_FIND_LATEST, lastHour).Order(SQL_DATE_DESC_ORDER).Find(&positions)

	sourcePoint := geo.NewPoint(latitude, longitude)
	entries := []nearbyEntry{}

	dates := make(map[uint]time.Time)

	for _, element := range positions {
		if element.Source == id {
			continue
		}

		existing, ok := dates[element.Source]
		if existing.After(element.Date) && ok {
			continue
		}
		dates[element.Source] = element.Date

		nearbyPoint := geo.NewPoint(element.Latitude, element.Longitude)
		distance := sourcePoint.GreatCircleDistance(nearbyPoint)
		if distance < MAX_DISTANCE {
			var account Account
			var profile Profile

			database.Where(SQL_FIND_PROFILE_BY_ID, element.Source).First(&profile)
			database.Where(SQL_FIND_PROFILE_BY_ID, profile.AccountID).First(&account)

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
	token, err := validateRequest(r)
	if err != nil {
		http.Error(w, STATUS_INVALID_TOKEN, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	latitude, err := strconv.ParseFloat(vars["latitude"], 64)
	if err != nil {
		http.Error(w, STATUS_INVALID_GPS, http.StatusBadRequest)
		return
	}

	longitude, err := strconv.ParseFloat(vars["longitude"], 64)
	if err != nil {
		http.Error(w, STATUS_INVALID_GPS, http.StatusBadRequest)
		return
	}

	id := getID(token)
	updatePosition(id, latitude, longitude)
	nearby := getNearby(id, latitude, longitude)

	sendJSONResponse(struct {
		Nearby []nearbyEntry `json:"nearby"`
	}{
		Nearby: nearby,
	}, w)
}
