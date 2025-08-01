package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
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

	// Start the server
	r := gin.Default()
	r.POST("/v1/finger/collect", HandleFingerprint)

	port := fmt.Sprintf(":%s", config.Server.Port)
	log.Printf("Server running on port %s", port)
	r.Run(port)
}
