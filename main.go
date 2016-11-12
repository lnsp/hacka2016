package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const (
	MEET_POINTS           = 25
	CONQUER_POINTS        = 10
	CONQUER_POINTS_SCALAR = 100

	DEFAULT_USER_COLOR    = "FF4081"
	DEFAULT_CONQUEROR     = ""
	DEFAULT_HOTSPOT_COLOR = "00FF00"

	BAD_JSON            = "Failed JSON parsing"
	BAD_COPY            = "Failed byte copying"
	INVALID_TOKEN       = "Invalid authentication token"
	INVALID_USER        = "Invalid user ID"
	INVALID_SSID        = "Invalid session ID"
	INVALID_NAME        = "Invalid user name"
	INVALID_DEVICE      = "Invalid device hash"
	MISSING_IMAGE       = "Image not found"
	MISSING_PROFILE     = "Profile not found"
	EXISTING_FRIENDSHIP = "Fuck off"

	MAX_DISTANCE = 0.1
	ULTIMATE_KEY = "lebonbon"
	CAPTURE_TIME = 180
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

// Handle generic endpoint requests.
func versionHandler(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(appEndpoint)
	if err != nil {
		fmt.Fprintln(w, "backend error")
		return
	}
	w.Write(data)
}

// Init database handle
func initDatabase() {
	var err error
	database, err = gorm.Open("sqlite3", "honeypot.db")
	if err != nil {
		panic(err)
	}
	database.AutoMigrate(&Hotspot{}, &Position{}, &Friendship{}, &Profile{}, &Account{})
	database.LogMode(true)
}

func main() {
	initDatabase()

	router := mux.NewRouter()

	router.HandleFunc("/", versionHandler)
	router.HandleFunc("/register", registerHandler).Methods("GET")
	router.HandleFunc("/profile/{id}", profileHandler).Methods("GET")
	router.HandleFunc("/profile", ownProfileHandler).Methods("GET")
	router.HandleFunc("/meet/{device}", meetHandler).Methods("GET")
	router.HandleFunc("/nearby/{latitude}/{longitude}", nearbyHandler).Methods("GET")
	router.HandleFunc("/capture/{ssid}", captureHotspotHandler).Methods("GET")
	router.HandleFunc("/hotspot/setup", setupHotspotHandler).Methods("GET")
	router.HandleFunc("/hotspot/fetch", fetchHotspotHandler).Methods("GET")
	router.HandleFunc("/hotspot/update", updateHotspotHandler).Methods("GET")

	router.HandleFunc("/settings/picture/{id}", getPictureHandler).Methods("GET")
	router.HandleFunc("/settings/picture", uploadPictureHandler).Methods("POST")
	router.HandleFunc("/settings/color", settingsColorHandler).Methods("GET")

	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(":8080", handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)))
}
