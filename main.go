package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net/http"
	"strconv"
	"time"
)

type AppEndpoint struct {
	Version  string `json:"version"`
	Endpoint string `json:"endpoint"`
}

type Profile struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Points  int    `json:"points"`
	Friends []uint `json:"friends"`
	Picture string `json:"picture"`
}

type ProfileModel struct {
	gorm.Model
	Name    string
	Points  int
	Friends []ProfileModel
	Picture string
}

type AccountModel struct {
	gorm.Model
	Device string
	Token  string
}

var database *gorm.DB

func validate(accessToken string) bool {
	var account AccountModel
	database.First(&account, "token = ?", accessToken)
	if account.ID == 0 {
		return false
	}
	return true
}

func generateToken(device string) string {
	timestamp := time.Now().Format("20060102150405")
	sha := sha1.New()
	sha.Write([]byte(timestamp))
	sha.Write([]byte(device))
	return hex.EncodeToString(sha.Sum(nil))
}

func createAccessToken(device string) string {
	token := generateToken(device)
	account := &AccountModel{
		Device: device,
		Token:  token,
	}
	database.Create(account)
	return token
}

func createProfile(name string) *Profile {
	profile := ProfileModel{
		Name:    name,
		Points:  0,
		Friends: []ProfileModel{},
		Picture: "",
	}
	database.Create(&profile)
	profileJson := &Profile{
		ID:      profile.ID,
		Name:    profile.Name,
		Friends: []uint{},
	}

	var friends []ProfileModel
	database.Model(&profile).Related(&friends)

	for _, element := range friends {
		profileJson.Friends = append(profileJson.Friends, element.ID)
	}

	return profileJson
}

func getProfile(id int) *Profile {
	var profile ProfileModel

	database.Where("id = ?", id).First(&profile)
	if profile.ID == 0 {
		return nil
	}

	profileJson := &Profile{
		ID:      profile.ID,
		Name:    profile.Name,
		Points:  profile.Points,
		Friends: []uint{},
		Picture: "",
	}

	var friends []ProfileModel
	database.Model(&profile).Related(&friends)

	for _, element := range friends {
		profileJson.Friends = append(profileJson.Friends, element.ID)
	}

	return profileJson
}

func getPicture(pictureID string) []byte {
	return []byte{}
}

func addPicture(id int, data []byte) {
	return
}

var appEndpoint = AppEndpoint{
	Version:  "1.0",
	Endpoint: "honeypot4431.cloudapp.net",
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(appEndpoint)
	if err != nil {
		fmt.Fprintln(w, "backend error")
		return
	}
	w.Write(data)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var profileID int
	var accessTokens []string
	var ok bool

	if accessTokens, ok = r.URL.Query()["token"]; !ok || len(accessTokens) != 1 || !validate(accessTokens[0]) {
		http.Error(w, "invalid access token", http.StatusUnauthorized)
		return
	}

	if profileIDs, ok := r.URL.Query()["id"]; !ok || len(profileIDs) != 1 {
		http.Error(w, "invalid profile id", http.StatusInternalServerError)
		return
	} else {
		profileID, err = strconv.Atoi(profileIDs[0])
		if err != nil {
			http.Error(w, "invalid profile id", http.StatusInternalServerError)
			return
		}
	}

	profile := getProfile(profileID)
	if profile == nil {
		http.Error(w, "invalid profile id", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(*profile)
	if err != nil {
		http.Error(w, "json error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var deviceIDs, names []string
	var err error
	var ok bool

	if deviceIDs, ok = r.URL.Query()["device"]; !ok || len(deviceIDs) != 1 {
		http.Error(w, "invalid device id", http.StatusUnauthorized)
		return
	}

	deviceID := deviceIDs[0]
	if names, ok = r.URL.Query()["name"]; !ok || len(names) != 1 {
		http.Error(w, "invalid name", http.StatusUnauthorized)
		return
	}
	name := names[0]

	token := createAccessToken(deviceID)
	profile := createProfile(name)

	if profile == nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(struct {
		Token string `json:"token"`
		ID    uint   `json:"id"`
	}{
		Token: token,
		ID:    profile.ID,
	})

	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func getPictureHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "get picture handler")
}

func uploadPictureHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "upload register handler")
}

func initDatabase() {
	var err error
	database, err = gorm.Open("sqlite3", "honeypot.db")
	if err != nil {
		panic(err)
	}
	database.AutoMigrate(&ProfileModel{}, &AccountModel{})
}

func main() {
	initDatabase()

	http.HandleFunc("/", versionHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/picture/get", getPictureHandler)
	http.HandleFunc("/picture/upload", uploadPictureHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
