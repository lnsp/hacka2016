package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
)

const PICTURE_FOLDER = ""

// Upload a new picture.
func addPicture(id uint, stream io.Reader) (string, error) {
	path := PICTURE_FOLDER + strconv.FormatUint(uint64(id), 10)
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		return "", err
	}

	return path, nil
}

// Handle /picture POST
func uploadPictureHandler(w http.ResponseWriter, r *http.Request) {
	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 || validate(accessTokens[0]) == nil {
		http.Error(w, "invalid access token", http.StatusUnauthorized)
		return
	}
	token := accessTokens[0]
	id := getID(token)
	path, err := addPicture(id, r.Body)
	if err != nil {
		http.Error(w, "failed to upload picture", http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(struct {
		Path string `json:"path"`
	}{
		Path: path,
	})
	if err != nil {
		http.Error(w, "json error", http.StatusInternalServerError)
	}

	w.Write(data)
}
