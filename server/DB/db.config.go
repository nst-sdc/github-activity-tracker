package DB

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github-activity-tracker/models"
)

var Database *gorm.DB

func InitDB() {
	// Load .env file
	if loadErr := godotenv.Load(); loadErr != nil {
		log.Println("Warning: Could not load .env file, using default/system environment variables")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Build DSN from individual environment variables
		host := getEnv("DB_HOST", "localhost")
		user := getEnv("DB_USER", "postgres")
		password := getEnv("DB_PASSWORD", "root")
		dbname := getEnv("DB_NAME", "githubdata")
		port := getEnv("DB_PORT", "5432")
		sslmode := getEnv("DB_SSLMODE", "disable")
		timezone := getEnv("DB_TIMEZONE", "Asia/Kolkata")

		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			host, user, password, dbname, port, sslmode, timezone)
	}

	var err error
	Database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Connected to PostgreSQL!")

	Database.AutoMigrate(&models.User{}, &models.PR{}, &models.Month{}, &models.Org{}, &models.Project{})
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return Database
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
