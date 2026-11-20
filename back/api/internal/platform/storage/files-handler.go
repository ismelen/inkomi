package storage

import (
	"errors"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
)

// FilesHandler serves the signed urls of LocalStorage: PUT uploads a file the way a presigned
// R2 url would and GET downloads it as an attachment. The signature is the only authorization.
type FilesHandler struct {
	store  *LocalStorage
	signer *URLSigner
}

func NewFilesHandler(store *LocalStorage, signer *URLSigner) *FilesHandler {
	return &FilesHandler{store: store, signer: signer}
}

// Setup registers /files/{id}/{filename}. It must stay out of request timeouts, since an upload
// from a phone can take minutes.
func (h *FilesHandler) Setup(r chi.Router) {
	r.Put("/files/{id}/{filename}", h.handlePut)
	r.Get("/files/{id}/{filename}", h.handleGet)
}

func (h *FilesHandler) handlePut(w http.ResponseWriter, r *http.Request) {
	id, filename := chi.URLParam(r, "id"), chi.URLParam(r, "filename")
	max, err := h.signer.Verify(http.MethodPut, id, filename, r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if max >= 0 && r.ContentLength > max {
		http.Error(w, ErrTooLarge.Error(), http.StatusRequestEntityTooLarge)
		return
	}

	n, err := h.store.Write(id, filename, r.Body, max)
	switch {
	case errors.Is(err, ErrTooLarge):
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	case err != nil:
		log.Printf("files: put %s/%s: %v", id, filename, err)
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}
	log.Printf("files: stored %s/%s (%d bytes)", id, filename, n)
	w.WriteHeader(http.StatusOK)
}

func (h *FilesHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, filename := chi.URLParam(r, "id"), chi.URLParam(r, "filename")
	if _, err := h.signer.Verify(http.MethodGet, id, filename, r.URL.Query()); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	path, err := h.store.Path(id, filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "download failed", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		http.Error(w, "download failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", AttachmentDisposition(filename))
	http.ServeContent(w, r, filename, info.ModTime(), file)
}

// AttachmentDisposition encodes filename as RFC 5987 so non-ASCII titles survive.
func AttachmentDisposition(filename string) string {
	encoded := strings.ReplaceAll(url.QueryEscape(filename), "+", "%20")
	return "attachment; filename*=UTF-8''" + encoded
}
