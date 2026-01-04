package MangaController

import "net/http"

func New() {
	http.HandleFunc("/convert", handleConvert)
	http.HandleFunc("/download", handleDownload)
}

func handleConvert(w http.ResponseWriter, r *http.Request) {

}

func handleDownload(w http.ResponseWriter, r *http.Request) {

}
