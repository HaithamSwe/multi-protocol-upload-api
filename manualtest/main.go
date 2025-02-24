package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	fileContent := []byte("Hello, S3! This is a test file.")
	filename := "test.txt"

	uploadURL := fmt.Sprintf("http://localhost:8080/upload-to-s3?filename=%s", filename)

	// HTTP Request:
	// POST /upload-to-s3?filename=test.txt
	// Body: [Binary file data]
	resp, err := http.Post(uploadURL, "application/octet-stream", bytes.NewBuffer(fileContent))
	if err != nil {
		log.Fatalf("Error making upload request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Fatalf("Upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Expected Response:
	// {
	//    "objectKey": "generated-object-key.txt"
	// }
	var uploadResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		log.Fatalf("Error decoding upload response: %v", err)
	}
	objectKey, ok := uploadResp["objectKey"]
	if !ok {
		log.Fatalf("objectKey not found in upload response")
	}
	fmt.Printf("Upload successful. Object Key: %s\n", objectKey)

	time.Sleep(1 * time.Second)

	expires := int64(3600) // Expiration time in seconds
	presignURLAPI := fmt.Sprintf("http://localhost:8080/get-presigned-s3-url?objectKey=%s&expires=%d", objectKey, expires)

	// HTTP Request:
	// GET /get-presigned-s3-url?objectKey=generated-object-key.txt&expires=3600
	getResp, err := http.Get(presignURLAPI)
	if err != nil {
		log.Fatalf("Error making presigned URL request: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(getResp.Body)
		log.Fatalf("Presigned URL request failed with status %d: %s", getResp.StatusCode, string(bodyBytes))
	}

	// Expected Response:
	// {
	//    "presignedURL": "https://s3.amazonaws.com/bucket-name/generated-object-key.txt?AWSAccessKeyId=..."
	// }
	var presignResp map[string]string
	if err := json.NewDecoder(getResp.Body).Decode(&presignResp); err != nil {
		log.Fatalf("Error decoding presigned URL response: %v", err)
	}
	presignedURL, ok := presignResp["presignedURL"]
	if !ok {
		log.Fatalf("presignedURL not found in response")
	}
	fmt.Printf("Presigned URL: %s\n", presignedURL)

	presignedResp, err := http.Get(presignedURL)
	if err != nil {
		log.Fatalf("Error fetching file from presigned URL: %v", err)
	}
	defer presignedResp.Body.Close()

	if presignedResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(presignedResp.Body)
		log.Fatalf("Fetching file failed with status %d: %s", presignedResp.StatusCode, string(bodyBytes))
	}

	fetchedContent, err := io.ReadAll(presignedResp.Body)
	if err != nil {
		log.Fatalf("Error reading file content from presigned URL: %v", err)
	}
	fmt.Printf("Content fetched from presigned URL: %s\n", string(fetchedContent))

	if bytes.Equal(fetchedContent, fileContent) {
		fmt.Println("File content is valid and matches.")
	} else {
		fmt.Println("File content does not match. Please verify the upload process.")
	}
}
