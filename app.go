package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func HandleFingerprint(c *gin.Context) {
	// Read and parse the JSON payload
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Generate filename components
	now := time.Now()
	uuidStr := uuid.New().String()
	datetimeStr := now.Format("2006-01-02 15:04:05")
	hashInput := uuidStr + datetimeStr

	// Create MD5 hash
	hasher := md5.New()
	hasher.Write([]byte(hashInput))
	hashValue := hex.EncodeToString(hasher.Sum(nil))

	// Format date as YYYYMMDD
	dateStr := now.Format("20060102")

	// Create filename
	filename := fmt.Sprintf("%s_VERSION_%d_%s.enc", hashValue, config.Fingerprints.Version, dateStr)

	// Ensure profile directory exists
	profilePath := os.ExpandEnv(config.Fingerprints.ProfilePath)
	if err := os.MkdirAll(profilePath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile directory"})
		return
	}

	// Full file path
	fullPath := filepath.Join(profilePath, filename)

	// Write the JSON payload to file
	if err := os.WriteFile(fullPath, body, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save profile file"})
		return
	}

	// Return success with filename
	c.JSON(http.StatusCreated, gin.H{"filename": filename})
}
