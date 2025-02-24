# Multi-Protocol Upload API (S3 Only)

This project provides a simple API for **uploading files** to **Amazon S3** and generating **pre-signed URLs** for accessing stored objects.

## Features 🚀
- Upload files to an S3 bucket.
- Generate pre-signed URLs for secure access to uploaded files.
- Simple, lightweight API with minimal dependencies.

---

## Installation & Setup 🛠

### Prerequisites
- **Go** (latest stable version recommended)
- **AWS Account & S3 Bucket**
- **task to run the Taskfile (optional, for easier development workflow)**

### 1️⃣ Clone the Repository
```sh
git clone https://github.com/haithamswe/multi-protocol-upload-api.git
cd multi-protocol-upload-api
```

### 2️⃣ Set Up Environment Variables
Create a `.env` file in the project root (or rename `.env.example` to `.env`) and update it with your S3 credentials:

```sh
S3_BUCKET=your-bucket-name
S3_REGION=your-region
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key
SERVER_PORT=8080
```

### 3️⃣ Install Dependencies
```sh
go mod tidy
```

### 4️⃣ Run the Application
```sh
go run cmd/main.go
```

Alternatively, if you have **Taskfile installed**, you can use:
```sh
task run
```

---

## API Endpoints 📡

### **1️⃣ Upload File to S3**
#### Endpoint:
```
POST /upload-to-s3?filename=<file_name>
```
#### Request:
- **Body:** Raw file data (binary)
- **Query Parameters:**
  - `filename` (string, required) - Name of the file being uploaded

#### Response:
```json
{
  "objectKey": "generated-object-key"
}
```
#### Example Usage (cURL):
```sh
curl -X POST "http://localhost:8080/upload-to-s3?filename=test.txt" \
     --data-binary @test.txt
```

---

### **2️⃣ Generate Pre-Signed S3 URL**
#### Endpoint:
```
GET /get-presigned-s3-url?objectKey=<file_key>&expires=<seconds>
```
#### Query Parameters:
- `objectKey` (string, required) - The key of the file in S3
- `expires` (integer, required) - Expiry time in seconds for the signed URL

#### Response:
```json
{
  "presignedURL": "https://s3.amazonaws.com/your-bucket/file?..."
}
```
#### Example Usage (cURL):
```sh
curl -X GET "http://localhost:8080/get-presigned-s3-url?objectKey=test.txt&expires=3600"
```

---

## Running Tests 🧪

### 1️⃣ Unit Tests:
```sh
go test ./... -v
```
or using Taskfile:
```sh
task test
```

### 2️⃣ Manual Testing:
Run:
```sh
go run manualtest/main.go
```

---

## Contributing 🤝

Feel free to fork the repo and submit a pull request! 🚀

---

## License 📜
This project is licensed under the **MIT License**.

---

## Author ✨
Developed by **[Haitham Alsaeed](https://github.com/haithamswe)**.

