package main

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleFingerprint(t *testing.T) {
	// Mock the database
	Db = setupMockDatabase()
	defer Db.Close()

	// Initialize the database and create table
	InitDatabase(Db)

	// Set up Gin router
	router := gin.Default()
	router.POST("/fingerprint", HandleFingerprint)

	// Test case 1: Valid payload
	t.Run("Valid payload", func(t *testing.T) {
		// Define a valid JSON payload
		payload := `{
			"visitor_id": "test_visitor_123",
			"user_agent": "Mozilla/5.0",
			"components": "{\"key\":\"value\"}"
		}`

		// Simulate HTTP POST request
		req, _ := http.NewRequest("POST", "/fingerprint", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		// Assert the response
		assert.Equal(t, http.StatusCreated, resp.Code)
	})

	// Test case 2: Invalid JSON payload
	t.Run("Invalid payload", func(t *testing.T) {
		// Define an invalid JSON payload
		payload := `{
			"visitor_id": "test_invalid",
			"user_agent": "Mozilla/5.0"
			"components": "{\"key\":\"value\""
		}`

		// Simulate HTTP POST request
		req, _ := http.NewRequest("POST", "/fingerprint", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		// Assert the response
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "Invalid request payload")
	})

	// Test case 3: Database error
	t.Run("Database error", func(t *testing.T) {
		// Tear down mock database to simulate failure
		Db.Close()

		// Define a valid JSON payload
		payload := `{
			"visitor_id": "test_error",
			"user_agent": "Mozilla/5.0",
			"components": "{\"key\":\"value\"}"
		}`

		// Simulate HTTP POST request
		req, _ := http.NewRequest("POST", "/fingerprint", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		// Assert the response
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Contains(t, resp.Body.String(), "Failed to save fingerprint")
	})
}

// setupMockDatabase creates a mock in-memory SQLite database for testing
func setupMockDatabase() *sql.DB {
	database, _ := sql.Open("sqlite3", ":memory:")
	return database
}
