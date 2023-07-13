package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	*gorm.DB
}

func Connect(url string) (*DB, error) {
	config := &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		})}

	db, err := gorm.Open(postgres.Open(url), config)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) Close() error {
	sql, err := db.DB.DB()
	if err != nil {
		return err
	}

	return sql.Close()
}
