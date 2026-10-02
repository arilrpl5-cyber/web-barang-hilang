package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv       string
	AppPort      string
	DBDialect    string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPass       string
	DBName       string
	JWTSecret    string
	MaxUploadMB  int64
	DefaultLimit int
	SchoolName   string
	AppName      string
}

func Load() *Config {
	maxMB, _ := strconv.ParseInt(getenv("MAX_UPLOAD_MB", "5"), 10, 64)
	defLimit, _ := strconv.Atoi(getenv("DEFAULT_LIMIT", "10"))
	return &Config{
		AppEnv:       getenv("APP_ENV", "development"),
		AppPort:      getenv("APP_PORT", "8080"),
		DBDialect:    getenv("DB_DIALECT", "auto"),
		DBHost:       getenv("DB_HOST", "localhost"),
		DBPort:       getenv("DB_PORT", "3306"),
		DBUser:       getenv("DB_USER", "root"),
		DBPass:       getenv("DB_PASSWORD", ""),
		DBName:       getenv("DB_NAME", "lost_found_school"),
		JWTSecret:    getenv("JWT_SECRET", "change-this-secret"),
		MaxUploadMB:  maxMB,
		DefaultLimit: defLimit,
		SchoolName:   getenv("SCHOOL_NAME", "Sekolah"),
		AppName:      getenv("APP_NAME", "Lost & Found School"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
