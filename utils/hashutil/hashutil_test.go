package hashutil_test

import (
	"encoding/hex"
	"github.com/haithamswe/multi-protocol-upload-api/utils/hashutil"
	"testing"
)

func TestHashSHA256(t *testing.T) {
	input := []byte("")
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	result := hashutil.HashSHA256(input)
	if result != expected {
		t.Errorf("Expected SHA256(\"\") = %s, got %s", expected, result)
	}

	input = []byte("hello")
	expected = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	result = hashutil.HashSHA256(input)
	if result != expected {
		t.Errorf("Expected SHA256(\"hello\") = %s, got %s", expected, result)
	}
}

func TestHmacSHA256(t *testing.T) {
	key := []byte("key")
	data := []byte("The quick brown fox jumps over the lazy dog")
	expectedHex := "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"
	result := hashutil.HmacSHA256(key, data)
	resultHex := hex.EncodeToString(result)
	if resultHex != expectedHex {
		t.Errorf("Expected HMAC-SHA256(\"key\", \"The quick brown fox jumps over the lazy dog\") = %s, got %s", expectedHex, resultHex)
	}
}
