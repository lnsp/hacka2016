package main

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	STATUS_INVALID_USER        = "Invalid user ID"
	STATUS_INVALID_SSID        = "Invalid session ID"
	STATUS_INVALID_DEVICE      = "Invalid device hash"
	STATUS_MISSING_IMAGE       = "Image not found"
	STATUS_MISSING_PROFILE     = "Profile not found"
	STATUS_EXISTING_FRIENDSHIP = "Friendship already exists"
)

// The JSON profile structure
type userProfile struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Points  uint   `json:"points"`
	Friends []uint `json:"friends"`
	Picture string `json:"picture"`
	Color   string `json:"color"`
}

// Increase user points
func increasePoints(id, amount uint) uint {
	// Find profile
	var profile Profile
	database.First(&profile, SQL_FIND_PROFILE_BY_ID, id)

	// Update profile points
	count := profile.Points + amount
	database.Model(&profile).Update(Profile{Points: count})

	// Return updated value
	return count
}

// Create a new profile linked to a user account.
func createAccount(device, name string) (uint, string) {
	var account Account
	var profile Profile

	database.First(&account, SQL_FIND_ACCOUNT_BY_DEVICE, device)

	// Account with device ID already exists
	if account.Token != "" {
		database.Model(&account).Related(&profile, ACCOUNT_PROFILE_RELATION)
		return profile.ID, account.Token
	}

	// Store new profile
	profile = Profile{
		Name:    name,
		Points:  DEFAULT_POINTS,
		Picture: DEFAULT_PICTURE_PATH,
		Color:   DEFAULT_USER_COLOR,
	}
	database.Create(&profile)

	// Store associated account
	token := generateToken(device)
	account = Account{
		Device: device,
		Token:  token,
		User:   profile,
	}
	database.Create(&account)

	// Return new credentials
	return account.User.ID, account.Token
}

// Get friend ids of specific user
func getFriends(id uint) []uint {
	var ids []uint
	var friendships []Friendship

	database.Where(SQL_FIND_FRIENDSHIPS_BY_SOURCE, id).Find(&friendships)
	for _, element := range friendships {
		ids = append(ids, element.Target)
	}

	return ids
}

// Get profile ID by account token
func getID(token string) uint {
	var account Account
	var profile Profile

	database.First(&account, SQL_FIND_ACCOUNT_BY_TOKEN, token)
	database.Model(&account).Related(&profile, ACCOUNT_PROFILE_RELATION)

	return profile.ID
}

// Retrieve profile by account token.
func getOwnProfile(token string) *Profile {
	var profile Profile
	var account Account

	database.First(&account, SQL_FIND_ACCOUNT_BY_TOKEN, token)
	database.Model(&account).Related(&profile, ACCOUNT_PROFILE_RELATION)

	return &profile
}

// Retrieve profile by ID.
func getProfile(id int) *Profile {
	var profile Profile

	database.First(&profile, id)

	return &profile
}

// Convert a profile model to a JSON compatible structure.
func toJSONProfile(profile *Profile) *userProfile {
	database.Model(&profile)

	profileJson := &userProfile{
		ID:      profile.ID,
		Name:    profile.Name,
		Points:  profile.Points,
		Friends: getFriends(profile.ID),
		Color:   profile.Color,
		Picture: profile.Picture,
	}

	return profileJson
}

// Handle /profile?token=
func ownProfileHandler(w http.ResponseWriter, r *http.Request) {
	token, err := validateRequest(r)
	if err != nil {
		http.Error(w, STATUS_INVALID_TOKEN, http.StatusUnauthorized)
		return
	}

	sendJSONResponse(toJSONProfile(getOwnProfile(token)), w)
}

// Handle /profile/{id}?token=
func profileHandler(w http.ResponseWriter, r *http.Request) {
	_, err := validateRequest(r)
	if err != nil {
		http.Error(w, STATUS_INVALID_TOKEN, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, STATUS_INVALID_USER, http.StatusNotFound)
	}
	sendJSONResponse(toJSONProfile(getProfile(profileID)), w)
}

// Handle /register?device&name
func registerHandler(w http.ResponseWriter, r *http.Request) {
	devices, ok := r.URL.Query()["device"]
	if !ok || len(devices) != 1 {
		http.Error(w, STATUS_INVALID_DEVICE, http.StatusUnauthorized)
		return
	}
	deviceID := devices[0]

	names, ok := r.URL.Query()["name"]
	if !ok || len(names) != 1 {
		http.Error(w, STATUS_INVALID_NAME, http.StatusUnauthorized)
		return
	}
	name := names[0]

	id, token := createAccount(deviceID, name)
	sendJSONResponse(struct {
		Token string `json:"token"`
		ID    uint   `json:"id"`
	}{
		Token: token,
		ID:    id,
	}, w)
}

func meetHandler(w http.ResponseWriter, r *http.Request) {
	token, err := validateRequest(r)
	if err != nil {
		http.Error(w, STATUS_INVALID_TOKEN, http.StatusUnauthorized)
		return
	}

	id := getID(token)

	// get nearby device id
	vars := mux.Vars(r)
	device := vars["device"]

	// get device user id
	var account Account
	database.First(&account, SQL_FIND_ACCOUNT_BY_DEVICE, device)
	if account.Token == "" {
		http.Error(w, STATUS_INVALID_DEVICE, http.StatusBadRequest)
		return
	}

	var profile Profile
	database.Model(&account).Related(&profile, ACCOUNT_PROFILE_RELATION)

	// Check if friendship exists
	var friendship Friendship
	database.First(&friendship, SQL_FIND_FRIENDSHIP, id, profile.ID)

	if friendship.Source == id {
		http.Error(w, STATUS_EXISTING_FRIENDSHIP, http.StatusBadRequest)
		return
	}

	// No -> Create one
	friendship = Friendship{
		Source: id,
		Target: profile.ID,
	}
	database.Create(&friendship)

	increasePoints(friendship.Source, MEET_POINTS)

	sendJSONResponse(struct {
		Source uint `json:"source"`
		Target uint `json:"target"`
	}{
		Source: id,
		Target: profile.ID,
	}, w)
}
