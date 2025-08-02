package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Fingerprints struct {
		ProfilePath string `yaml:"profile_path"`
		Version     int    `yaml:"version"`
	} `yaml:"fingerprints"`
}

var config Config
var db *sql.DB

func initDatabase() {
	// Ensure profile directory exists and get the path
	profilePath := os.ExpandEnv(config.Fingerprints.ProfilePath)
	if err := os.MkdirAll(profilePath, 0755); err != nil {
		log.Fatalf("Error creating profile directory: %v", err)
	}

	// Create database path in the same directory as profiles
	dbPath := filepath.Join(profilePath, "fingerprints.db")

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Create fingerprints table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS fingerprints (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		filename TEXT NOT NULL
	);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating fingerprints table: %v", err)
	}

	// Create index on key column for better query performance
	createIndexSQL := `CREATE INDEX IF NOT EXISTS idx_fingerprints_key ON fingerprints(key);`
	_, err = db.Exec(createIndexSQL)
	if err != nil {
		log.Fatalf("Error creating index on key column: %v", err)
	}

	log.Printf("Database initialized successfully at: %s", dbPath)
}

func loadConfig() {
	// Use os.ReadFile (modern replacement for ioutil.ReadFile)
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Expand environment variables in the YAML content
	expandedData := []byte(os.ExpandEnv(string(data)))

	err = yaml.Unmarshal(expandedData, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}
}

func main() {
	// Load configuration from YAML
	loadConfig()

	// Initialize database
	initDatabase()
	defer db.Close()

	// Start the server
	r := gin.Default()
	r.POST("/v1/finger/collect/:key", HandleFingerprint)

	port := fmt.Sprintf(":%s", config.Server.Port)
	log.Printf("Server running on port %s", port)
	r.Run(port)
}
