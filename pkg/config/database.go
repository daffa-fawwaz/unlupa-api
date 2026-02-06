package config

import (
	"fmt"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/seeders"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {

	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Warning: .env file not found, using system environment variables")
	}

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")
	timezone := os.Getenv("DB_TIMEZONE")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host, user, password, dbname, port, sslmode, timezone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect database:", err)
	}

	DB = db

	log.Println("🚀 Running AutoMigrate...")
	err = db.AutoMigrate(
		&entities.User{},
		&entities.Kitab{},
		&entities.Class{},
		&entities.ClassMember{},
		&entities.Card{},
		&entities.CardState{},
		&entities.TeacherRequest{},
		&entities.ItemState{},
		&entities.Item{},
		&entities.ItemGraduation{},
		&entities.FSRSState{},
		&entities.DailyTask{},
		&entities.ReviewState{},
		&entities.ReviewLog{},
		&entities.EngineControl{},
		&entities.FSRSWeights{},
		&entities.Card{},
		&entities.Juz{},
		&entities.JuzItem{},
		&entities.Book{},
		&entities.BookModule{},
		&entities.BookItem{},
		&entities.ClassBook{},

	)
	if err != nil {
		log.Fatal("❌ Failed to migrate:", err)
	}

	log.Println("✅ Database connected and migrated successfully!")

	// Run seeders
	var seedUser entities.User
	if err := db.First(&seedUser).Error; err == nil {
		if err := seeders.SeedItems(db, seedUser.ID); err != nil {
			log.Println("⚠️  Failed to seed items:", err)
		} else {
			log.Println("✅ Items seeded successfully!")
		}
	} else {
		log.Println("⚠️  No user found, skipping item seeder")
	}

}

