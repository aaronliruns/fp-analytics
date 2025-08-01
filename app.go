package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type TouchSupport struct {
	MaxTouchPoints int  `json:"maxTouchPoints"`
	TouchEvent     bool `json:"touchEvent"`
	TouchStart     bool `json:"touchStart"`
}

type VideoCard struct {
	Vendor   string `json:"vendor"`
	Renderer string `json:"renderer"`
}

type Components struct {
	ScreenResolution struct {
		Value    []int `json:"value"`
		Duration int   `json:"duration"`
	} `json:"screenResolution"`
	HardwareConcurrency struct {
		Value    int `json:"value"`
		Duration int `json:"duration"`
	} `json:"hardwareConcurrency"`
	Platform struct {
		Value    string `json:"value"`
		Duration int    `json:"duration"`
	} `json:"platform"`
	TouchSupport struct {
		Value    TouchSupport `json:"value"`
		Duration int          `json:"duration"`
	} `json:"touchSupport"`
	VideoCard struct {
		Value    VideoCard `json:"value"`
		Duration int       `json:"duration"`
	} `json:"videoCard"`
	WebGlBasics struct {
		Value struct {
			VendorUnmasked   string `json:"vendorUnmasked"`
			RendererUnmasked string `json:"rendererUnmasked"`
		} `json:"value"`
		Duration int `json:"duration"`
	} `json:"webGlBasics"`
	Architecture struct {
		Value    int `json:"value"`
		Duration int `json:"duration"`
	} `json:"architecture"`
}

type FingerprintResponse struct {
	RowNumber           int          `json:"row_number"`
	UserAgent           string       `json:"user_agent"`
	ScreenResolution    []int        `json:"screen_resolution"`
	HardwareConcurrency int          `json:"hardware_concurrency"`
	Platform            string       `json:"platform"`
	TouchSupport        TouchSupport `json:"touch_support"`
	VideoCard           VideoCard    `json:"video_card"`
	Architecture        int          `json:"architecture"`
	DPR                 float64      `json:"dpr"`
}

type Fingerprint struct {
	VisitorID  string `json:"visitor_id"`
	UserAgent  string `json:"user_agent"`
	Components string `json:"components"`
	DPR        string `json:"dpr"` // New field for device pixel ratio
}

var Db *sql.DB

func InitDatabase(database *sql.DB) {
	Db = database
	createTable()
}

func createTable() {
	createTableSQL := `CREATE TABLE IF NOT EXISTS fingerprints (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          visitor_id TEXT UNIQUE,
          user_agent TEXT,
          components TEXT,
          dpr REAL
      );
      
      -- Create indexes for better performance
      CREATE INDEX IF NOT EXISTS idx_fingerprints_id ON fingerprints (id);
      CREATE INDEX IF NOT EXISTS idx_fingerprints_visitor_id ON fingerprints (visitor_id);`
	_, err := Db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func HandleFingerprint(c *gin.Context) {
	var fp Fingerprint
	if err := c.ShouldBindJSON(&fp); err != nil {
		var fieldErrors []string
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			fieldErrors = append(fieldErrors, fmt.Sprintf("Invalid type for field %s", err.(*json.UnmarshalTypeError).Field))
		} else if errs, ok := err.(validator.ValidationErrors); ok {
			for _, e := range errs {
				fieldErrors = append(fieldErrors, fmt.Sprintf("Invalid value for field %s", e.Field()))
			}
		}

		if len(fieldErrors) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": fieldErrors})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		}
		return
	}

	err := SaveFingerprint(fp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save fingerprint"})
		return
	}

	c.Status(http.StatusCreated)
}

func SaveFingerprint(fp Fingerprint) error {
	// Convert DPR string to float
	dpr, err := strconv.ParseFloat(fp.DPR, 64)
	if err != nil {
		return err
	}

	insertSQL := `INSERT INTO fingerprints (visitor_id, user_agent, components, dpr) 
                  VALUES (?, ?, ?, ?)
                  ON CONFLICT(visitor_id) DO UPDATE SET
                  user_agent = excluded.user_agent,
                  components = excluded.components,
                  dpr = excluded.dpr;`
	_, err = Db.Exec(insertSQL, fp.VisitorID, fp.UserAgent, fp.Components, dpr)
	return err
}

func HandleFingerprintCount(c *gin.Context) {
	var count int
	err := Db.QueryRow("SELECT COUNT(id) FROM fingerprints").Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get fingerprint count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func HandleFingerprintRow(c *gin.Context) {
	// Get row number from query parameter
	rowNum := c.Query("row")
	if rowNum == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Row number is required"})
		return
	}

	// Convert row number to integer
	rowNumber, err := strconv.Atoi(rowNum)
	if err != nil || rowNumber < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid row number"})
		return
	}

	// Query the specific row using direct id access instead of OFFSET
	var userAgent, componentsStr string
	var dpr float64
	err = Db.QueryRow("SELECT user_agent, components, dpr FROM fingerprints WHERE id = ?", rowNumber).Scan(&userAgent, &componentsStr, &dpr)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Row not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get fingerprint row"})
		return
	}

	// Parse components JSON string
	var components Components
	err = json.Unmarshal([]byte(componentsStr), &components)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse components data"})
		return
	}

	// Create response object
	response := FingerprintResponse{
		RowNumber:           rowNumber,
		UserAgent:           userAgent,
		ScreenResolution:    components.ScreenResolution.Value,
		HardwareConcurrency: components.HardwareConcurrency.Value,
		Platform:            components.Platform.Value,
		TouchSupport:        components.TouchSupport.Value,
		Architecture:        components.Architecture.Value,
		DPR:                 dpr,
	}

	// Handle video card mapping based on webGlBasics availability
	if components.WebGlBasics.Value.VendorUnmasked != "" {
		// When webGlBasics is available, use vendorUnmasked and rendererUnmasked
		response.VideoCard = VideoCard{
			Vendor:   components.WebGlBasics.Value.VendorUnmasked,
			Renderer: components.WebGlBasics.Value.RendererUnmasked,
		}
	} else {
		// When webGlBasics is not available, use the original videoCard values
		response.VideoCard = components.VideoCard.Value
	}

	c.JSON(http.StatusOK, response)
}
