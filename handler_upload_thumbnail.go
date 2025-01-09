package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	// Get the JWT from the request
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	// Validate the JWT and get the user ID
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// Set maximum memory to 10MB
	const maxMemory = 10 << 20

	// Parse the multipart form
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse form", err)
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get file", err)
		return
	}
	defer file.Close()

	// Get file media type from the header
	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get media type", err)
		return
	}

	// Check if the media type is an image
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "File must be an image", nil)
		return
	}

	// Get the video metadata from the database
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get video from database", err)
		return
	}

	// Check if the user is the owner of the video
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "User is not the owner of the video", nil)
		return
	}

	// Create a random file name
	randomFileName := createRandomFileName()

	// Create thumbnail file name and path
	thumbExt := getThumbnailFileExtension(mediaType)
	thumbFileName := fmt.Sprintf("%s.%s", randomFileName, thumbExt)
	thumbFilePath := filepath.Join(cfg.assetsRoot, thumbFileName)

	// Save the thumbnail to the file system
	serverFile, err := os.Create(thumbFilePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create file on the server", err)
		return
	}
	defer serverFile.Close()

	_, err = io.Copy(serverFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save file to the server", err)
		return
	}

	// Update the video in the database
	thumbUrl := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, thumbFileName)
	video.ThumbnailURL = &thumbUrl
	video.UpdatedAt = time.Now()
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video in database", err)
		return
	}

	// Respond with the updated video
	respondWithJSON(w, http.StatusOK, video)
}
