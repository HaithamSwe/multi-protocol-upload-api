package s3

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"
)

func TestGetSignatureKey(t *testing.T) {
	s := s3{
		secretKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
		region:    "us-east-1",
	}
	dateStamp := "20130524"
	expectedHex := "f117494eff5d09da21cbf7f0339559ea04fc9582d31299cb992be70a6b27c97a"
	key := s.getSignatureKey(dateStamp)
	keyHex := hex.EncodeToString(key)
	if keyHex != expectedHex {
		t.Errorf("Expected signature key %s, got %s", expectedHex, keyHex)
	}
}

func TestGetSignatureKey_DifferentDates(t *testing.T) {
	s := s3{
		secretKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
		region:    "us-east-1",
	}
	dateStamp1 := "20130524"
	dateStamp2 := "20130525"
	key1 := s.getSignatureKey(dateStamp1)
	key2 := s.getSignatureKey(dateStamp2)
	if hex.EncodeToString(key1) == hex.EncodeToString(key2) {
		t.Error("Expected different signature keys for different dateStamps")
	}
}

func TestBuildCanonicalRequest(t *testing.T) {
	tests := []struct {
		name                 string
		canonicalURI         string
		canonicalQueryString string
		headers              map[string]string
		signedHeaders        string
		hashedPayload        string
		expected             string
	}{
		{
			name:                 "Empty headers",
			canonicalURI:         "/object",
			canonicalQueryString: "a=b",
			headers:              map[string]string{},
			signedHeaders:        "",
			hashedPayload:        "hash123",
			expected:             "PUT\n/object\na=b\n\n\nhash123",
		},
		{
			name:                 "Multiple headers with sorting and trimming",
			canonicalURI:         "/object",
			canonicalQueryString: "",
			headers: map[string]string{
				"Host":         "example.amazonaws.com",
				"X-Amz-Date":   " 20230101T123456Z ",
				"Content-Type": " application/json ",
			},
			signedHeaders: "content-type;host;x-amz-date",
			hashedPayload: "hash456",
			expected: fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
				http.MethodPut,
				"/object",
				"",
				"content-type:application/json\nhost:example.amazonaws.com\nx-amz-date:20230101T123456Z\n",
				"content-type;host;x-amz-date",
				"hash456",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCanonicalRequest(tt.canonicalURI, tt.canonicalQueryString, tt.headers, tt.signedHeaders, tt.hashedPayload)
			if result != tt.expected {
				t.Errorf("Test %s failed:\nExpected:\n%q\nGot:\n%q", tt.name, tt.expected, result)
			}
		})
	}
}
