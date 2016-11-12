package main

import (
	"encoding/json"
	"net/http"
)

func settingsNameHandler(w http.ResponseWriter, r *http.Request) {
	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 || validate(accessTokens[0]) == nil {
		http.Error(w, "invalid access token", http.StatusUnauthorized)
		return
	}

	names, ok := r.URL.Query()["name"]
	if !ok || len(names) != 1 {
		http.Error(w, "invalid color", http.StatusBadRequest)
		return
	}
	name := names[0]
	id := getID(accessTokens[0])

	var profile Profile
	database.First(&profile, "ID = ?", id)
	database.Model(&profile).Update(Profile{Name: name})

	data, err := json.Marshal(toJSONProfile(profile))
	if err != nil {
		http.Error(w, "json error", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func settingsColorHandler(w http.ResponseWriter, r *http.Request) {
	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 || validate(accessTokens[0]) == nil {
		http.Error(w, "invalid access token", http.StatusUnauthorized)
		return
	}

	colors, ok := r.URL.Query()["color"]
	if !ok || len(colors) != 1 {
		http.Error(w, "invalid color", http.StatusBadRequest)
		return
	}
	rgb := colors[0]
	id := getID(accessTokens[0])

	var profile Profile
	database.First(&profile, "ID = ?", id)
	database.Model(&profile).Update(Profile{Color: rgb})

	data, err := json.Marshal(toJSONProfile(profile))
	if err != nil {
		http.Error(w, "json error", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
