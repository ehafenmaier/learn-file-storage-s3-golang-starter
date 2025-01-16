package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"math"
	"os/exec"
	"strings"
	"time"
)

func createRandomFileName() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}

func getThumbnailFileExtension(mediaType string) string {
	switch mediaType {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	default:
		return "jpg"
	}
}

func getVideoAspectRatio(filePath string) (string, error) {
	// Build and run the ffprobe command
	var b bytes.Buffer
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	cmd.Stdout = &b
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	// Unmarshal the JSON output
	var output map[string]interface{}
	err = json.Unmarshal(b.Bytes(), &output)
	if err != nil {
		return "", err
	}

	// Calculate the aspect ratio from the width and height
	stream := output["streams"].([]interface{})[0].(map[string]interface{})
	width := stream["width"].(float64)
	height := stream["height"].(float64)
	ratio := math.Round((width/height)*100) / 100

	// Return the aspect ratio as a string
	switch ratio {
	case 1.78:
		return "16:9", nil
	case 0.56:
		return "9:16", nil
	default:
		return "other", nil
	}
}

func processVideoForFastStart(filePath string) (string, error) {
	// Build and run the ffmpeg command
	outputFilePath := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputFilePath)
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	// Return the output file path
	return outputFilePath, nil
}

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	// Create and use presign client
	presignClient := s3.NewPresignClient(s3Client)
	presignedURL, err := presignClient.PresignGetObject(context.TODO(),
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		}, s3.WithPresignExpires(expireTime))

	if err != nil {
		return "", err
	}

	// Return the presigned URL
	return presignedURL.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	// If the video URL is nil, return the video as is
	if video.VideoURL == nil {
		return video, nil
	}

	// Split the video URL into bucket and key
	urlSplit := strings.Split(*video.VideoURL, ",")
	if len(urlSplit) != 2 {
		return video, fmt.Errorf("invalid video URL format")
	}
	bucket := urlSplit[0]
	key := urlSplit[1]

	// Generate a presigned URL for the video for 10 minutes
	presignedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, 10*time.Minute)
	if err != nil {
		return video, err
	}

	// Update the video URL with the presigned URL
	video.VideoURL = &presignedURL

	return video, nil
}
