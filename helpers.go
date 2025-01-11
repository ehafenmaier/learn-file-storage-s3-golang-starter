package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"math"
	"os/exec"
)

func createRandomFileName() string {
	b := make([]byte, 32)
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
