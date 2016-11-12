package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/kellydunn/golang-geo"
)

const (
	MAX_DISTANCE = 0.1
)

// The database endpoint
var database *gorm.DB

// The API endpoint
var appEndpoint = JSONEndpoint{
	Version:  "1.0",
	Endpoint: "honeypot4431.cloudapp.net",
}

// The JSON endpoint structure
type JSONEndpoint struct {
	Version  string `json:"version"`
	Endpoint string `json:"endpoint"`
}

// The JSON profile structure
type JSONProfile struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Points  int    `json:"points"`
	Friends []uint `json:"friends"`
	Picture string `json:"picture"`
}

type Position struct {
	Date      time.Time
	Latitude  float64
	Longitude float64
	Source    uint
}

type Friendship struct {
	gorm.Model
	Source uint
	Target uint
}

// The internal profile model
type Profile struct {
	gorm.Model
	Name      string
	Points    int
	Picture   string
	AccountID uint `gorm:"index"`
}

// The internal account model
type Account struct {
	gorm.Model
	Device string
	Token  string
	User   Profile
}

// Validate an access token.
func validate(accessToken string) bool {
	var account Account
	database.First(&account, "Token = ?", accessToken)
	if account.ID == 0 {
		return false
	}
	return true
}

// Generate a new access token.
func generateToken(device string) string {
	timestamp := time.Now().Format("20060102150405")
	sha := sha1.New()
	sha.Write([]byte(timestamp))
	sha.Write([]byte(device))
	return hex.EncodeToString(sha.Sum(nil))
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

// Create a new profile linked to a user account.
func createAccount(device, name string) (uint, string) {
	var account Account
	var profile Profile

	database.First(&account, "Device = ?", device)
	if account.Token != "" {
		return account.User.ID, account.Token
	}

	profile = Profile{
		Name:    name,
		Points:  0,
		Picture: "",
	}
	database.Create(&profile)

	token := generateToken(device)
	account = Account{
		Device: device,
		Token:  token,
		User:   profile,
	}
	database.Create(&account)

	// Create demo relationships
	otherProfiles := make([]Profile, 0)
	database.Find(&otherProfiles)

	for _, element := range otherProfiles {
		if element.ID == profile.ID {
			continue
		}

		database.Create(&Friendship{
			Source: profile.ID,
			Target: element.ID,
		})
	}

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
}

func getNearby(id uint, latitude, longitude float64) []nearbyEntry {
	lastHour := time.Now().Add(-time.Hour)

	positions := []Position{}
	database.Where("date > ?", lastHour).Find(&positions)

	sourcePoint := geo.NewPoint(latitude, longitude)
	entries := []nearbyEntry{}

	for _, element := range positions {
		if element.Source == id {
			continue
		}

		nearbyPoint := geo.NewPoint(element.Latitude, element.Longitude)
		distance := sourcePoint.GreatCircleDistance(nearbyPoint)
		if distance < MAX_DISTANCE {
			entries = append(entries, nearbyEntry{
				ID:       element.Source,
				Date:     element.Date,
				Distance: distance,
			})
		}
	}

	return entries
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

// Retrieve picture by hash.
func getPicture(pictureID string) []byte {
	return []byte{}
}

// Upload a new picture.
func addPicture(id int, data []byte) string {
	return ""
}

// Handle nearby
func nearbyHandler(w http.ResponseWriter, r *http.Request) {
	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 || !validate(accessTokens[0]) {
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

	data, err := json.Marshal(nearby)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// Handle generic endpoint requests.
func versionHandler(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(appEndpoint)
	if err != nil {
		fmt.Fprintln(w, "backend error")
		return
	}
	w.Write(data)
}

// Handle /profile?token=
func ownProfileHandler(w http.ResponseWriter, r *http.Request) {
	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 || !validate(accessTokens[0]) {
		http.Error(w, "invalid access token", http.StatusUnauthorized)
		return
	}
	token := accessTokens[0]

	profile := getOwnProfile(token)
	if profile == nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(profile)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// Handle /profile/{id}?token=
func profileHandler(w http.ResponseWriter, r *http.Request) {
	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 || !validate(accessTokens[0]) {
		http.Error(w, "invalid access token", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	profileID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "malformed profile id", http.StatusNotFound)
	}
	profile := getProfile(profileID)

	if profile == nil {
		http.Error(w, "profile not found", http.StatusNotFound)
		return
	}

	data, err := json.Marshal(*profile)
	if err != nil {
		http.Error(w, "json error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// Handle /register?device&name
func registerHandler(w http.ResponseWriter, r *http.Request) {
	devices, ok := r.URL.Query()["device"]
	if !ok || len(devices) != 1 {
		http.Error(w, "invalid device id", http.StatusUnauthorized)
		return
	}
	deviceID := devices[0]

	names, ok := r.URL.Query()["name"]
	if !ok || len(names) != 1 {
		http.Error(w, "invalid name", http.StatusUnauthorized)
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
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// Handle /picture GET
func getPictureHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "get picture handler")
}

// Handle /picture POST
func uploadPictureHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "upload register handler")
}

// Init database handle
func initDatabase() {
	var err error
	database, err = gorm.Open("sqlite3", "honeypot.db")
	if err != nil {
		panic(err)
	}
	database.AutoMigrate(&Position{}, &Friendship{}, &Profile{}, &Account{})
	database.LogMode(true)
}

func main() {
	initDatabase()

	router := mux.NewRouter()

	router.HandleFunc("/", versionHandler)
	router.HandleFunc("/register", registerHandler).Methods("GET")
	router.HandleFunc("/profile/{id}", profileHandler).Methods("GET")
	router.HandleFunc("/profile", ownProfileHandler).Methods("GET")
	router.HandleFunc("/picture", getPictureHandler).Methods("GET")
	router.HandleFunc("/picture", uploadPictureHandler).Methods("POST")
	router.HandleFunc("/nearby/{latitude}/{longitude}", nearbyHandler).Methods("GET")

	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(":8080", handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)))
}
