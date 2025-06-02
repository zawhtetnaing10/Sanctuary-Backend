package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func UploadFileToAWS(inputParamName string, fileType string, request *http.Request, s3Client *s3.Client, s3Bucket string, s3Region string) (string, error) {
	file, header, fileErr := request.FormFile(inputParamName)
	if fileErr != nil {
		return "", fileErr
	}
	defer file.Close()

	// Check if the uploaded file has the correct mime type
	mediaType, _, mimeErr := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if mimeErr != nil {
		return "", mimeErr
	}

	if !strings.HasPrefix(mediaType, fileType) {
		return "", errors.New("wrong file type")
	}

	// Get the original file extension
	originalExtension := filepath.Ext(header.Filename)
	if originalExtension == "" {
		return "", errors.New("the original file has no extension")
	}

	tmpFile, tmpErr := os.CreateTemp("", "sanctuary-upload-*"+originalExtension)
	if tmpErr != nil {
		return "", tmpErr
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy the contents to tmp file
	io.Copy(tmpFile, file)

	// Reset temp file's pointer to beginning
	tmpFile.Seek(0, io.SeekStart)

	// Random 32 bytes for image name
	randomBytes := make([]byte, 32)
	_, randomBytesErr := rand.Read(randomBytes)
	if randomBytesErr != nil {
		return "", randomBytesErr
	}
	random32BytesString := hex.EncodeToString(randomBytes)

	// Create file key
	fileKey := fmt.Sprintf("profiles/%v%v", random32BytesString, originalExtension)

	// Upload the file
	_, putObjErr := s3Client.PutObject(request.Context(), &s3.PutObjectInput{
		Bucket:      aws.String(s3Bucket),
		Key:         aws.String(fileKey),
		Body:        tmpFile,
		ContentType: aws.String(mediaType),
	})

	if putObjErr != nil {
		return "", putObjErr
	}

	// Get the download url
	downloadUrl := fmt.Sprintf("https://%v.s3.%v.amazonaws.com/%v", s3Bucket, s3Region, fileKey)

	return downloadUrl, nil
}
