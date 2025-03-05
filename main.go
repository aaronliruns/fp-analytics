package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Name      string `yaml:"name"`
		TableName string `yaml:"table_name"`
	} `yaml:"database"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
}

var config Config

func loadConfig() {
	// Use os.ReadFile (modern replacement for ioutil.ReadFile)
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}
}

func main() {
	// Load configuration from YAML
	loadConfig()

	// Initialize database connection
	var err error
	Db, err = sql.Open("sqlite3", config.Database.Name)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer Db.Close()

	// Initialize database with dynamic table name
	InitDatabase(Db)

	// Start the server
	r := gin.Default()
	r.POST("/fingerprint", HandleFingerprint)
	r.GET("/fingerprints/count", HandleFingerprintCount)
	r.GET("/fingerprints/row", HandleFingerprintRow)

	port := fmt.Sprintf(":%s", config.Server.Port)
	log.Printf("Server running on port %s", port)
	r.Run(port)
}
