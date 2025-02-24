package main

import (
	"fmt"
	"github.com/haithamswe/multi-protocol-upload-api/handlers"
	"github.com/haithamswe/multi-protocol-upload-api/s3"
	"github.com/haithamswe/multi-protocol-upload-api/utils/timeutil"
	"github.com/haithamswe/multi-protocol-upload-api/utils/uuidutil"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Error: Unable to get caller info")
	}

	projectRoot := filepath.Dir(filepath.Dir(filename))
	envPath := filepath.Join(projectRoot, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Fatal("Error loading .env file from root:", err)
	}

	bucket := os.Getenv("S3_BUCKET")
	region := os.Getenv("S3_REGION")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	port := os.Getenv("SERVER_PORT")

	if bucket == "" || region == "" || accessKey == "" || secretKey == "" || port == "" {
		log.Fatal("Missing required environment variables")
	}

	timeUtil := timeutil.NewTimeUtil()
	uuidUtil := uuidutil.NewUUIDUtil()
	s3Client := s3.NewS3(bucket, region, accessKey, secretKey, timeUtil, uuidUtil)
	handlers := handlers.NewHandlers(s3Client)

	http.HandleFunc("/upload-to-s3", handlers.UploadToS3)
	http.HandleFunc("/get-presigned-s3-url", handlers.GetPresignedS3Url)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
