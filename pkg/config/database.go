package config

import (
	"fmt"
	"hifzhun-api/pkg/entities"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	slowThreshold := 500 * time.Millisecond
	if envThreshold := os.Getenv("SLOW_QUERY_THRESHOLD_MS"); envThreshold != "" {
		if ms, parseErr := time.ParseDuration(envThreshold + "ms"); parseErr == nil && ms > 0 {
			slowThreshold = ms
		}
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: NewGormSlowQueryLogger(slowThreshold),
	})
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
		&entities.BookItemOverride{},
		&entities.ClassBook{},
		&entities.IntervalReviewLog{},
		&entities.BookUpdateRequest{},
		&entities.ImportedBook{},
		&entities.QuranPageProgress{},
		&entities.QuranJuzCatalog{},
		&entities.QuranPageCatalog{},
	)
	if err != nil {
		log.Fatal("❌ Failed to migrate:", err)
	}

	log.Println("✅ Database connected and migrated successfully!")
	if err := SeedQuranCatalog(db); err != nil {
		log.Println("⚠️ Quran Catalog seeding error:", err)
	}
	BackfillImportedBooks(db)
}

func BackfillImportedBooks(db *gorm.DB) {
	log.Println("🔄 Backfilling legacy imported books...")
	var items []entities.Item
	err := db.Where("source_type = ? AND content_ref LIKE ?", "book", "book:%").Find(&items).Error
	if err != nil {
		log.Println("⚠️ Failed to fetch legacy items for backfill:", err)
		return
	}

	if len(items) == 0 {
		log.Println("✅ Backfill completed: no legacy items found")
		return
	}

	bookIDSet := make(map[string]bool)
	for _, item := range items {
		parts := strings.Split(item.ContentRef, ":")
		if len(parts) == 4 && parts[0] == "book" && parts[2] == "item" {
			bookIDSet[parts[1]] = true
		}
	}

	if len(bookIDSet) == 0 {
		log.Println("✅ Backfill completed: no valid book IDs found")
		return
	}

	var bookIDs []string
	for id := range bookIDSet {
		bookIDs = append(bookIDs, id)
	}

	var books []entities.Book
	err = db.Where("id IN ? AND status = ?", bookIDs, "published").Find(&books).Error
	if err != nil {
		log.Println("⚠️ Failed to fetch books for backfill:", err)
		return
	}

	bookMap := make(map[uuid.UUID]entities.Book)
	for _, book := range books {
		bookMap[book.ID] = book
	}

	count := 0
	for _, item := range items {
		parts := strings.Split(item.ContentRef, ":")
		if len(parts) != 4 || parts[0] != "book" || parts[2] != "item" {
			continue
		}
		bookUUID, err := uuid.Parse(parts[1])
		if err != nil {
			continue
		}

		book, exists := bookMap[bookUUID]
		if !exists {
			continue
		}

		if book.OwnerID != item.OwnerID {
			imported := entities.ImportedBook{
				ID:        uuid.New(),
				UserID:    item.OwnerID,
				BookID:    book.ID,
				CreatedAt: time.Now(),
			}
			// Use OnConflict DoNothing to ensure idempotency and avoid duplicate key errors
			err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&imported).Error
			if err == nil {
				// Note: GORM's Create returns no error even on conflict if DoNothing is used.
				// However, GORM sets affected rows. If we want exact count, we can check or just log general completion.
				count++
			}
		}
	}
	log.Printf("✅ Backfill completed: processed %d potential imported book records", count)
}

// DBPoolStats represents the runtime connection pool metrics from database/sql
type DBPoolStats struct {
	MaxOpenConnections int   `json:"max_open_connections"`
	OpenConnections    int   `json:"open_connections"`
	InUse              int   `json:"in_use"`
	Idle               int   `json:"idle"`
	WaitCount          int64 `json:"wait_count"`
	WaitDurationMs     int64 `json:"wait_duration_ms"`
	MaxIdleClosed      int64 `json:"max_idle_closed"`
	MaxIdleTimeClosed  int64 `json:"max_idle_time_closed"`
	MaxLifetimeClosed  int64 `json:"max_lifetime_closed"`
}

// GetDBPoolStats safely extracts in-memory sql.DB runtime stats without querying PostgreSQL
func GetDBPoolStats() (*DBPoolStats, bool) {
	if DB == nil {
		return nil, false
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return nil, false
	}
	stats := sqlDB.Stats()
	return &DBPoolStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDurationMs:     stats.WaitDuration.Milliseconds(),
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxIdleTimeClosed:  stats.MaxIdleTimeClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}, true
}

