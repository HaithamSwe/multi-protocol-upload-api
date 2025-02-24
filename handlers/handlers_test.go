package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/haithamswe/multi-protocol-upload-api/handlers"
	"github.com/haithamswe/multi-protocol-upload-api/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUploadToS3(t *testing.T) {
	mockS3 := mocks.NewS3(t)
	mockS3.On("Upload", mock.AnythingOfType("[]uint8"), "test.txt").
		Return("uploaded-test.txt", nil)

	h := handlers.NewHandlers(mockS3)

	req := httptest.NewRequest(http.MethodPost, "/upload?filename=test.txt", bytes.NewBufferString("file content"))
	rec := httptest.NewRecorder()

	h.UploadToS3(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "uploaded-test.txt", response["objectKey"])

	mockS3.AssertExpectations(t)
}

func TestGetPresignedS3Url(t *testing.T) {
	mockS3 := mocks.NewS3(t)
	mockS3.On("PresignUrl", "test.txt", int64(3600)).
		Return("http://example.com/test.txt?expires=3600")

	h := handlers.NewHandlers(mockS3)

	// --- Valid Request ---
	req := httptest.NewRequest(http.MethodGet, "/presign?objectKey=test.txt&expires=3600", nil)
	rec := httptest.NewRecorder()

	h.GetPresignedS3Url(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com/test.txt?expires=3600", response["presignedURL"])

	// --- Edge Case: Missing expires parameter ---
	reqMissing := httptest.NewRequest(http.MethodGet, "/presign?objectKey=test.txt", nil)
	recMissing := httptest.NewRecorder()

	h.GetPresignedS3Url(recMissing, reqMissing)
	assert.Equal(t, http.StatusBadRequest, recMissing.Code)

	// --- Edge Case: Invalid expires parameter ---
	reqInvalid := httptest.NewRequest(http.MethodGet, "/presign?objectKey=test.txt&expires=notanumber", nil)
	recInvalid := httptest.NewRecorder()

	h.GetPresignedS3Url(recInvalid, reqInvalid)
	assert.Equal(t, http.StatusBadRequest, recInvalid.Code)

	mockS3.AssertExpectations(t)
}
