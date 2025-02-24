package s3_test

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/haithamswe/multi-protocol-upload-api/mocks"
	"github.com/haithamswe/multi-protocol-upload-api/s3"
	"github.com/stretchr/testify/assert"
)

func TestPresignUrl(t *testing.T) {
	fixedTime := time.Date(2025, 2, 24, 15, 4, 5, 0, time.UTC)
	mockTimeUtil := mocks.NewTimeUtil(t)
	mockTimeUtil.On("Now").Return(fixedTime)

	mockUUIDUtil := mocks.NewUUIDUtil(t)

	bucket := "testbucket"
	region := "us-test-1"
	accessKey := "TESTACCESSKEY"
	secretKey := "TESTSECRETKEY"
	s3Instance := s3.NewS3(bucket, region, accessKey, secretKey, mockTimeUtil, mockUUIDUtil)

	objectKey := "test.txt"
	expires := int64(3600)
	presignedURL := s3Instance.PresignUrl(objectKey, expires)

	parsedURL, err := url.Parse(presignedURL)
	assert.NoError(t, err)
	assert.Equal(t, "https", parsedURL.Scheme)

	expectedHost := bucket + ".s3." + region + ".amazonaws.com"
	assert.Equal(t, expectedHost, parsedURL.Host)

	q := parsedURL.Query()
	assert.Equal(t, "AWS4-HMAC-SHA256", q.Get("X-Amz-Algorithm"))
	assert.NotEmpty(t, q.Get("X-Amz-Credential"))
	assert.NotEmpty(t, q.Get("X-Amz-Date"))
	assert.Equal(t, "UNSIGNED-PAYLOAD", q.Get("X-Amz-Content-Sha256"))
	assert.NotEmpty(t, q.Get("X-Amz-Signature"))

	mockTimeUtil.AssertExpectations(t)
}

func TestUpload(t *testing.T) {
	fixedTime := time.Date(2025, 2, 24, 15, 4, 5, 0, time.UTC)
	fixedUUID := "fixed-uuid"

	mockTimeUtil := mocks.NewTimeUtil(t)
	mockTimeUtil.On("Now").Return(fixedTime)

	mockUUIDUtil := mocks.NewUUIDUtil(t)
	mockUUIDUtil.On("Generate").Return(fixedUUID)

	bucket := "testbucket"
	region := "us-test-1"
	accessKey := "TESTACCESSKEY"
	secretKey := "TESTSECRETKEY"
	s3Instance := s3.NewS3(bucket, region, accessKey, secretKey, mockTimeUtil, mockUUIDUtil)

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
	}))
	defer ts.Close()

	origTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = origTransport }()

	http.DefaultTransport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial(network, ts.Listener.Addr().String())
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	fileContent := []byte("file content")
	fileName := "filename.txt"
	objectKey, err := s3Instance.Upload(fileContent, fileName)
	assert.NoError(t, err)

	expectedObjectKey := fixedUUID + "_" + fileName
	assert.Equal(t, expectedObjectKey, objectKey)

	mockTimeUtil.AssertExpectations(t)
	mockUUIDUtil.AssertExpectations(t)
}
