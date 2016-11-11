package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type AppEndpoint struct {
	Version  string `json:"version"`
	Endpoint string `json:"endpoint"`
}

type Profile struct {
	ID       int    `json:"id"`
	DeviceID string `json:"device"`
	Name     string `json:"name"`
	Points   int    `json:"points"`
	Friends  []int  `json:"friends"`
	Picture  string `json:"picture"`
}

func validate(accessToken string) bool {
	return true
}

func createProfile(device string, name string) *Profile {
	return &Profile{
		ID:      1,
		Name:    "Mr. Pot",
		Points:  38,
		Friends: []int{},
		Picture: "potpicture",
	}
}

func getProfile(id int) *Profile {
	return &Profile{
		ID:      1,
		Name:    "Mr. Pot",
		Points:  38,
		Friends: []int{},
		Picture: "potpicture",
	}
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

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var profileID int
	var accessToken []string
	var ok bool

	if accessToken, ok = r.URL.Query()["token"]; !ok || len(accessToken) != 1 || !validate(accessToken[0]) {
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

func profileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "profile handler")
}

func getPictureHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "get picture handler")
}

func uploadPictureHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "upload register handler")
}

func main() {
	http.HandleFunc("/", versionHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/picture/get", getPictureHandler)
	http.HandleFunc("/picture/upload", uploadPictureHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
