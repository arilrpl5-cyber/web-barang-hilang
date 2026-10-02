package main

import (
	"log"
	"os"
	"path/filepath"

	"lost-found-school/config"
	"lost-found-school/routes"
	"lost-found-school/seed"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	db := config.Connect(cfg)
	seed.Run(db)

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)
	// templates: dukung templates/*.html dan templates/admin/*.html (nama template = nama file)
	files, _ := filepath.Glob("templates/*.html")
	adminFiles, _ := filepath.Glob("templates/admin/*.html")
	files = append(files, adminFiles...)
	if len(files) == 0 {
		log.Fatal("tidak ada template ditemukan di templates/")
	}
	r.LoadHTMLFiles(files...)
	r.Static("/static", "./static")
	r.Static("/uploads", "./uploads")

	routes.Setup(r, db, cfg)

	port := cfg.AppPort
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	log.Println("Server running at http://localhost:" + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
