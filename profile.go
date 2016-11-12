package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// The JSON profile structure
type JSONProfile struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Points  uint   `json:"points"`
	Friends []uint `json:"friends"`
	Picture string `json:"picture"`
	Color   string `json:"color"`
}

func increasePoints(id, amount uint) uint {
	var profile Profile
	count := profile.Points + amount
	database.First(&profile, "ID = ?", id)
	database.Model(&profile).Update(Profile{Points: count})
	return count
}

// Create a new profile linked to a user account.
func createAccount(device, name string) (uint, string) {
	var account Account
	var profile Profile

	database.First(&account, "Device = ?", device)
	if account.Token != "" {
		database.Model(&account).Related(&profile, "User")
		return profile.ID, account.Token
	}

	profile = Profile{
		Name:    name,
		Points:  0,
		Picture: "",
		Color:   DEFAULT_USER_COLOR,
	}
	database.Create(&profile)

	token := generateToken(device)
	account = Account{
		Device: device,
		Token:  token,
		User:   profile,
	}
	database.Create(&account)

	return account.User.ID, account.Token
}

func getFriends(id uint) []uint {
	ids := make([]uint, 0, 0)

	var friendships []Friendship
	database.Where("Source = ?", id).Find(&friendships)

	for _, element := range friendships {
		ids = append(ids, element.Target)
	}

	return ids
}

func getID(token string) uint {
	var account Account
	var profile Profile

	database.First(&account, "Token = ?", token)
	database.Model(&account).Related(&profile, "User")

	return profile.ID
}

// Retrieve profile by account token.
func getOwnProfile(token string) *JSONProfile {
	var profile Profile
	var account Account

	database.First(&account, "Token = ?", token)
	database.Model(&account).Related(&profile, "User")

	return toJSONProfile(profile)
}

// Convert a profile model to a JSON compatible structure.
func toJSONProfile(profile Profile) *JSONProfile {
	database.Model(&profile)

	profileJson := &JSONProfile{
		ID:      profile.ID,
		Name:    profile.Name,
		Points:  profile.Points,
		Friends: getFriends(profile.ID),
		Color:   profile.Color,
		Picture: "",
	}
	return profileJson
}

// Retrieve profile by ID.
func getProfile(id int) *JSONProfile {
	var profile Profile

	database.First(&profile, id)
	if profile.ID == 0 {
		return nil
	}

	return toJSONProfile(profile)
}

// Handle /profile?token=
func ownProfileHandler(w http.ResponseWriter, r *http.Request) {
	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 || validate(accessTokens[0]) == nil {
		http.Error(w, "invalid access token", http.StatusUnauthorized)
		return
	}
	token := accessTokens[0]

	profile := getOwnProfile(token)
	if profile == nil {
		http.Error(w, MISSING_PROFILE, http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(profile)
	if err != nil {
		http.Error(w, BAD_JSON, http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// Handle /profile/{id}?token=
func profileHandler(w http.ResponseWriter, r *http.Request) {
	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 || validate(accessTokens[0]) == nil {
		http.Error(w, INVALID_TOKEN, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, INVALID_USER, http.StatusNotFound)
	}
	profile := getProfile(profileID)

	if profile == nil {
		http.Error(w, MISSING_PROFILE, http.StatusNotFound)
		return
	}

	data, err := json.Marshal(*profile)
	if err != nil {
		http.Error(w, BAD_JSON, http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// Handle /register?device&name
func registerHandler(w http.ResponseWriter, r *http.Request) {
	devices, ok := r.URL.Query()["device"]
	if !ok || len(devices) != 1 {
		http.Error(w, INVALID_DEVICE, http.StatusUnauthorized)
		return
	}
	deviceID := devices[0]

	names, ok := r.URL.Query()["name"]
	if !ok || len(names) != 1 {
		http.Error(w, INVALID_NAME, http.StatusUnauthorized)
		return
	}
	name := names[0]

	id, token := createAccount(deviceID, name)
	data, err := json.Marshal(struct {
		Token string `json:"token"`
		ID    uint   `json:"id"`
	}{
		Token: token,
		ID:    id,
	})

	if err != nil {
		http.Error(w, BAD_JSON, http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func meetHandler(w http.ResponseWriter, r *http.Request) {
	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 || validate(accessTokens[0]) == nil {
		http.Error(w, INVALID_TOKEN, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	device := vars["device"]
	id := getID(accessTokens[0])

	// get device user id
	var account Account
	database.First(&account, "Device = ?", device)
	if account.Token == "" {
		http.Error(w, INVALID_DEVICE, http.StatusBadRequest)
		return
	}

	var profile Profile
	database.Model(&account).Related(&profile, "User")

	// Check if friendship exists
	var friendship Friendship
	database.First(&friendship, "Source = ? AND Target = ?", id, profile.ID)
	if friendship.Source == id {
		http.Error(w, EXISTING_FRIENDSHIP, http.StatusBadRequest)
		return
	}

	// No -> Create one
	friendship = Friendship{
		Source: id,
		Target: profile.ID,
	}
	database.Create(&friendship)

	increasePoints(friendship.Source, MEET_POINTS)

	data, err := json.Marshal(struct {
		Source uint `json:"source"`
		Target uint `json:"target"`
	}{
		Source: id,
		Target: profile.ID,
	})
	if err != nil {
		http.Error(w, BAD_JSON, http.StatusInternalServerError)
		return
	}

	w.Write(data)
}
