package ereader

import (
	"net/http"
	"os"
)

func NewController() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/session/", handleSession)
	http.HandleFunc("/poll", handlePoll)

	os.MkdirAll("downloads", os.ModeAppend)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {

}

func handleSession(w http.ResponseWriter, r *http.Request) {

}

func handlePoll(w http.ResponseWriter, r *http.Request) {

}
