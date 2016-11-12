package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

const PICTURE_FOLDER = "picture/"

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

// Handle /picture GET
func getPictureHandler(w http.ResponseWriter, r *http.Request) {
	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 || validate(accessTokens[0]) == nil {
		http.Error(w, INVALID_TOKEN, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)

	path := PICTURE_FOLDER + strconv.FormatUint(id, 10)
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, MISSING_IMAGE, http.StatusNotFound)
		return
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, BAD_COPY, http.StatusInternalServerError)
		return
	}
}

// Handle /picture POST
func uploadPictureHandler(w http.ResponseWriter, r *http.Request) {
	accessTokens, ok := r.URL.Query()["token"]
	if !ok || len(accessTokens) != 1 || validate(accessTokens[0]) == nil {
		http.Error(w, INVALID_TOKEN, http.StatusUnauthorized)
		return
	}
	token := accessTokens[0]
	id := getID(token)
	path, err := addPicture(id, r.Body)
	if err != nil {
		http.Error(w, BAD_COPY, http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(struct {
		Path string `json:"path"`
	}{
		Path: path,
	})
	if err != nil {
		http.Error(w, BAD_JSON, http.StatusInternalServerError)
	}

	w.Write(data)
}
