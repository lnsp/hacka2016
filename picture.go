package main

import (
	"fmt"
	"net/http"
)

// Retrieve picture by hash.
func getPicture(pictureID string) []byte {
	return []byte{}
}

// Upload a new picture.
func addPicture(id int, data []byte) string {
	return ""
}

// Handle /picture GET
func getPictureHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "get picture handler")
}

// Handle /picture POST
func uploadPictureHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "upload register handler")
}
