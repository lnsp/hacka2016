package main

import "net/http"

const (
	STATUS_INVALID_COLOR = "Invalid color code"
	STATUS_INVALID_NAME  = "Invalid user name"
)

func settingsNameHandler(w http.ResponseWriter, r *http.Request) {
	token, err := validateRequest(r)
	if err != nil {
		http.Error(w, STATUS_INVALID_TOKEN, http.StatusUnauthorized)
		return
	}

	names, ok := r.URL.Query()["name"]
	if !ok || len(names) != 1 {
		http.Error(w, STATUS_INVALID_NAME, http.StatusBadRequest)
		return
	}
	name := names[0]
	profile := getOwnProfile(token)
	database.Model(&profile).Update(Profile{Name: name})

	sendJSONResponse(toJSONProfile(profile), w)
}

func settingsColorHandler(w http.ResponseWriter, r *http.Request) {
	token, err := validateRequest(r)
	if err != nil {
		http.Error(w, STATUS_INVALID_TOKEN, http.StatusUnauthorized)
		return
	}

	colors, ok := r.URL.Query()["color"]
	if !ok || len(colors) != 1 {
		http.Error(w, STATUS_INVALID_COLOR, http.StatusBadRequest)
		return
	}
	rgb := colors[0]
	profile := getOwnProfile(token)
	database.Model(profile).Update(Profile{Color: rgb})

	sendJSONResponse(toJSONProfile(profile), w)
}
