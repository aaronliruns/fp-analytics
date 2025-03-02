package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
          visitor_id TEXT PRIMARY KEY,
          user_agent TEXT,
          components TEXT,
          dpr REAL
      );`
	_, err := Db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func HandleFingerprint(c *gin.Context) {
	var fp Fingerprint
	if err := c.ShouldBindJSON(&fp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
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
