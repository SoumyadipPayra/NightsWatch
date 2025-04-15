package conn

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	//generally it is secret stored int the cloud : for simplicity using local
	host     = "localhost"
	port     = "5432"
	user     = "admin"
	password = "admin123"
	dbName   = "nightswatch"
)

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
	config := Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
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
