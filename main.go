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
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Points  int    `json:"points"`
	Friends []int  `json:"friends"`
	Picture string `json:"picture"`
}

func validate(accessToken string) bool {
	return accessToken == "mrpot"
}

func createAccessToken(device string) string {
	return "mrpot"
}

func createProfile(name string) *Profile {
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
		ID    int    `json:"id"`
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

func main() {
	http.HandleFunc("/", versionHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/picture/get", getPictureHandler)
	http.HandleFunc("/picture/upload", uploadPictureHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
