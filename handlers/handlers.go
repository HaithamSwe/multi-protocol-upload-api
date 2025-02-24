package handlers

import (
	"encoding/json"
	"github.com/haithamswe/multi-protocol-upload-api/s3"
	"io"
	"net/http"
	"strconv"
)

type Handlers interface {
	UploadToS3(w http.ResponseWriter, r *http.Request)
	GetPresignedS3Url(w http.ResponseWriter, r *http.Request)
}

type handlers struct {
	s3Client s3.S3
}

func (h handlers) UploadToS3(w http.ResponseWriter, r *http.Request) {
	fileData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "io.ReadAll returned error", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fileName := r.URL.Query().Get("filename")

	objectKey, err := h.s3Client.Upload(fileData, fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"objectKey": objectKey,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h handlers) GetPresignedS3Url(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	objectKey := r.URL.Query().Get("objectKey")
	if objectKey == "" {
		http.Error(w, "Missing objectKey parameter", http.StatusBadRequest)
		return
	}
	expiresStr := r.URL.Query().Get("expires")
	if expiresStr == "" {
		http.Error(w, "Missing expires parameter", http.StatusBadRequest)
		return
	}
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid expires parameter", http.StatusBadRequest)
		return
	}

	presignedURL := h.s3Client.PresignUrl(objectKey, expires)

	response := map[string]string{
		"presignedURL": presignedURL,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func NewHandlers(s3Client s3.S3) Handlers {
	return handlers{
		s3Client: s3Client,
	}
}
