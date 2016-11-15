package main

import (
	"encoding/json"
	"errors"
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
	MAX_UNIQ_SSID_LEN     = 24
	TIMESTAMP_FORMAT      = 20060102150405
	MEET_POINTS           = 1
	CONQUER_POINTS        = 5
	CONQUER_POINTS_SCALAR = 100

	DEFAULT_SSID_PREFIX   = "honeypot"
	DEFAULT_USER_COLOR    = "FF4081"
	DEFAULT_CONQUEROR     = ""
	DEFAULT_HOTSPOT_COLOR = "00FF00"
	DEFAULT_PICTURE_PATH  = ""
	DEFAULT_POINTS        = 0

	STATUS_BAD_JSON = "Failed JSON parsing"
	STATUS_BAD_COPY = "Failed byte copying"

	MAX_DISTANCE = 0.1
	ULTIMATE_KEY = "lebonbon"
	CAPTURE_TIME = 180
)

var invalidTokenError error = errors.New("Invalid access token")

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

	router.HandleFunc("/settings/name", settingsNameHandler).Methods("GET")
	router.HandleFunc("/settings/picture/{id}", getPictureHandler).Methods("GET")
	router.HandleFunc("/settings/picture", uploadPictureHandler).Methods("POST")
	router.HandleFunc("/settings/color", settingsColorHandler).Methods("GET")

	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(":8080", handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)))
}

func sendJSONResponse(element interface{}, w http.ResponseWriter) {
	data, err := json.Marshal(element)
	if err != nil {
		http.Error(w, STATUS_BAD_JSON, http.StatusInternalServerError)
		return
	}

	w.Write(data)
}
