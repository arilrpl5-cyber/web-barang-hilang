package config

import (
	"fmt"
	"log"

	"lost-found-school/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *Config) *gorm.DB {
	// Try MySQL first if dialect is mysql or auto
	if cfg.DBDialect == "mysql" || cfg.DBDialect == "auto" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err == nil {
			log.Println("Connected to MySQL:", cfg.DBName)
			autoMigrate(db)
			return db
		}
		log.Println("MySQL connection failed:", err)
		if cfg.DBDialect == "mysql" {
			log.Fatal("DB_DIALECT=mysql but connection failed. Create DB first: CREATE DATABASE " + cfg.DBName + ";")
		}
		log.Println("Falling back to SQLite (data.db). Set DB_DIALECT=mysql + DB_* to use MySQL.")
	}

	db, err := gorm.Open(sqlite.Open("data.db?_pragma=busy_timeout%285000%29&_pragma=journal_mode%28WAL%29"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("failed to connect sqlite:", err)
	}
	log.Println("Connected to SQLite: data.db")
	autoMigrate(db)
	return db
}

func autoMigrate(db *gorm.DB) {
	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Report{},
		&models.Match{},
		&models.Claim{},
		&models.Notification{},
		&models.ActivityLog{},
		&models.Setting{},
	); err != nil {
		log.Fatal("automigrate failed:", err)
	}
}
