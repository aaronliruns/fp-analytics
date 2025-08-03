package main

import (
	"crypto/md5"
	"database/sql"
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
	// Get key from URL parameter
	keyStr := c.Param("key")
	if keyStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing key parameter in URL"})
		return
	}

	// Read and parse the JSON payload
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		// Check if error is due to request body being too large
		if err.Error() == "http: request body too large" {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request body too large. Maximum allowed size is 50MB",
				"code":  "REQUEST_TOO_LARGE",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		}
		return
	}

	// Check if key already exists in database
	var existingID int
	err = db.QueryRow("SELECT id FROM fingerprints WHERE key = ?", keyStr).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		// Key already exists, return duplicate response
		c.JSON(http.StatusConflict, gin.H{"message": "Fingerprint with this key already exists", "duplicate": true})
		return
	}

	// Generate filename components (line 40 equivalent)
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

	// Insert into database
	_, err = db.Exec("INSERT INTO fingerprints (key, filename) VALUES (?, ?)", keyStr, filename)
	if err != nil {
		// If database insert fails, clean up the file
		os.Remove(fullPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save fingerprint to database"})
		return
	}

	// Return success with filename
	c.JSON(http.StatusCreated, gin.H{"filename": filename, "key": keyStr, "duplicate": false})
}
