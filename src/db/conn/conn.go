package conn

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config holds the database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

var DB *gorm.DB //singleton

// Initialize creates a connection to the database and
// stores the reference to it in the DB variable
func Initialize(ctx context.Context, models ...interface{}) error {
	// Get configuration from environment variables with defaults
	config := Config{
		Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:     getEnvOrDefault("POSTGRES_PORT", "5432"),
		User:     getEnvOrDefault("POSTGRES_USER", "admin"),
		Password: getEnvOrDefault("POSTGRES_PASSWORD", "admin123"),
		DBName:   getEnvOrDefault("POSTGRES_DB", "nightswatch"),
	}
	logger := ctx.Value("logger").(*zap.Logger)

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect to database",
			zap.Error(err),
			zap.String("host", config.Host),
			zap.String("port", config.Port),
			zap.String("dbname", config.DBName),
		)
		return err
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("failed to get underlying *sql.DB",
			zap.Error(err),
		)
		return err
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		logger.Error("failed to ping database",
			zap.Error(err),
		)
		return err
	}

	DB = db
	for _, model := range models {
		err = db.AutoMigrate(model)
		if err != nil {
			logger.Error("failed to migrate model", zap.Error(err))
			return err
		}
	}
	logger.Info("Database migrated successfully")
	return nil
}

// GetDB returns a handle to the DB object
func GetDB(ctx context.Context) *gorm.DB {
	return DB.WithContext(ctx)
}

// getEnvOrDefault returns the environment variable value or the default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
