package config

import (
	"log"
	"os"
	"time"
)

var AppLocation *time.Location

func InitAppLocation() {
	tz := os.Getenv("APP_TIMEZONE")
	if tz == "" {
		tz = os.Getenv("DB_TIMEZONE")
	}
	if tz == "" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Printf("⚠️  Failed to load timezone '%s', defaulting to UTC: %v", tz, err)
		loc = time.UTC
	}
	AppLocation = loc
	log.Printf("🕒 App timezone set to %s", tz)
}
