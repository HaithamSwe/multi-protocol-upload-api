package s3

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/haithamswe/multi-protocol-upload-api/utils/hashutil"
	"github.com/haithamswe/multi-protocol-upload-api/utils/timeutil"
	"github.com/haithamswe/multi-protocol-upload-api/utils/uuidutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type s3 struct {
	bucket    string
	region    string
	accessKey string
	secretKey string
	timeUtil  timeutil.TimeUtil
	uuidUtil  uuidutil.UUIDUtil
}

type S3 interface {
	PresignUrl(objectKey string, expires int64) string
	Upload(fileData []byte, fileName string) (string, error)
}

func (s s3) PresignUrl(objectKey string, expires int64) string {
	host := fmt.Sprintf("%s.s3.%s.amazonaws.com", s.bucket, s.region)
	canonicalURI := "/" + objectKey

	t := s.timeUtil.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	queryParams := map[string]string{
		"X-Amz-Algorithm":      "AWS4-HMAC-SHA256",
		"X-Amz-Credential":     fmt.Sprintf("%s/%s/%s/%s/aws4_request", s.accessKey, dateStamp, s.region, "s3"),
		"X-Amz-Date":           amzDate,
		"X-Amz-Expires":        fmt.Sprintf("%d", expires),
		"X-Amz-SignedHeaders":  "host",
		"X-Amz-Content-Sha256": "UNSIGNED-PAYLOAD",
	}

	var keys []string
	for k := range queryParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var canonicalQueryParts []string
	for _, k := range keys {
		canonicalQueryParts = append(canonicalQueryParts, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(queryParams[k])))
	}
	canonicalQueryString := strings.Join(canonicalQueryParts, "&")

	canonicalHeaders := fmt.Sprintf("host:%s\n", host)
	signedHeaders := "host"
	payloadHash := "UNSIGNED-PAYLOAD"
	canonicalRequest := fmt.Sprintf("GET\n%s\n%s\n%s\n%s\n%s", canonicalURI, canonicalQueryString, canonicalHeaders, signedHeaders, payloadHash)
	hashedCanonicalRequest := hashutil.HashSHA256([]byte(canonicalRequest))

	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, s.region, "s3")
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s", amzDate, credentialScope, hashedCanonicalRequest)

	signingKey := s.getSignatureKey(dateStamp)
	signature := hex.EncodeToString(hashutil.HmacSHA256(signingKey, []byte(stringToSign)))

	finalQueryString := canonicalQueryString + "&" + "X-Amz-Signature=" + signature

	presignedURL := fmt.Sprintf("https://%s%s?%s", host, canonicalURI, finalQueryString)

	return presignedURL
}

func (s *s3) Upload(fileData []byte, fileName string) (string, error) {
	if fileName == "" {
		fileName = "default_filename"
	}
	objectKey := fmt.Sprintf("%s_%s", s.uuidUtil.Generate(), fileName)

	req, err := s.signRequest(objectKey, fileData)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error from S3, status code: %d", resp.StatusCode)
	}

	return objectKey, nil
}

func (s *s3) signRequest(objectKey string, payload []byte) (*http.Request, error) {
	host := fmt.Sprintf("%s.s3.%s.amazonaws.com", s.bucket, s.region)
	endpoint := fmt.Sprintf("https://%s/%s", host, objectKey)

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	t := s.timeUtil.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	req.Header.Set("Host", host)
	req.Header.Set("x-amz-date", amzDate)
	hashedPayload := hashutil.HashSHA256(payload)
	req.Header.Set("x-amz-content-sha256", hashedPayload)

	canonicalURI := "/" + objectKey
	canonicalQueryString := ""

	headersForSigning := map[string]string{
		"host":                 host,
		"x-amz-content-sha256": hashedPayload,
		"x-amz-date":           amzDate,
	}
	var headerKeys []string
	for k := range headersForSigning {
		headerKeys = append(headerKeys, strings.ToLower(k))
	}
	sort.Strings(headerKeys)
	signedHeaders := strings.Join(headerKeys, ";")

	canonicalRequest := buildCanonicalRequest(canonicalURI, canonicalQueryString, headersForSigning, signedHeaders, hashedPayload)
	hashedCanonicalRequest := hashutil.HashSHA256([]byte(canonicalRequest))

	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, s.region, "s3")
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s", amzDate, credentialScope, hashedCanonicalRequest)

	signingKey := s.getSignatureKey(dateStamp)
	signatureHMAC := hashutil.HmacSHA256(signingKey, []byte(stringToSign))
	signature := hex.EncodeToString(signatureHMAC)

	authorizationHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.accessKey, credentialScope, signedHeaders, signature)
	req.Header.Set("Authorization", authorizationHeader)

	return req, nil
}

func (s s3) getSignatureKey(dateStamp string) []byte {
	kDate := hashutil.HmacSHA256([]byte("AWS4"+s.secretKey), []byte(dateStamp))
	kRegion := hashutil.HmacSHA256(kDate, []byte(s.region))
	kService := hashutil.HmacSHA256(kRegion, []byte("s3"))
	kSigning := hashutil.HmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}

func buildCanonicalRequest(canonicalURI, canonicalQueryString string, headers map[string]string, signedHeaders, hashedPayload string) string {
	lowerHeaders := make(map[string]string)
	var headerKeys []string

	for k, v := range headers {
		lowerKey := strings.ToLower(k)
		lowerHeaders[lowerKey] = strings.TrimSpace(v)
		headerKeys = append(headerKeys, lowerKey)
	}

	sort.Strings(headerKeys)

	canonicalHeaders := ""
	for _, k := range headerKeys {
		canonicalHeaders += fmt.Sprintf("%s:%s\n", k, lowerHeaders[k])
	}

	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		http.MethodPut,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		hashedPayload,
	)
}

func NewS3(bucket, region, accessKey, secretKey string, timeUtil timeutil.TimeUtil, uuidUtil uuidutil.UUIDUtil) S3 {
	return &s3{
		bucket:    bucket,
		region:    region,
		accessKey: accessKey,
		secretKey: secretKey,
		timeUtil:  timeUtil,
		uuidUtil:  uuidUtil,
	}
}
