package main

import (
	"crypto/rand"
	"encoding/base64"
)

func createRandomFileName() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(b)
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
