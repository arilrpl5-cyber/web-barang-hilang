package seed

import (
	"log"
	"time"

	"lost-found-school/models"
	"lost-found-school/utils"

	"gorm.io/gorm"
)

func Run(db *gorm.DB) {
	// Categories
	cats := []models.Category{
		{Name: "Elektronik", Description: "HP, laptop, charger, earphone"},
		{Name: "Dompet", Description: "Dompet dan uang"},
		{Name: "Kartu", Description: "Kartu pelajar, KTP, ATM"},
		{Name: "Buku", Description: "Buku pelajaran, novel, modul"},
		{Name: "Pakaian", Description: "Seragam, jaket, topi"},
		{Name: "Tas", Description: "Tas sekolah, ransel"},
		{Name: "Alat Tulis", Description: "Pulpen, pensil, penggaris"},
		{Name: "Aksesori", Description: "Jam, kacamata, gelang"},
		{Name: "Lainnya", Description: "Barang lain"},
	}
	for _, c := range cats {
		var cnt int64
		db.Model(&models.Category{}).Where("name = ?", c.Name).Count(&cnt)
		if cnt == 0 {
			db.Create(&c)
		}
	}

	// Admin
	var cnt int64
	db.Model(&models.User{}).Where("email = ?", "admin@school.local").Count(&cnt)
	if cnt == 0 {
		hash, _ := utils.HashPassword("Admin123!")
		db.Create(&models.User{Name: "Administrator", NisNisn: "ADM001", Email: "admin@school.local", Password: hash, Role: "ADMIN", Status: "ACTIVE"})
		log.Println("Seed admin: admin@school.local / Admin123!")
	}
	db.Model(&models.User{}).Where("email = ?", "user@school.local").Count(&cnt)
	if cnt == 0 {
		hash, _ := utils.HashPassword("User123!")
		db.Create(&models.User{Name: "Siswa Demo", NisNisn: "12345678", Email: "user@school.local", Password: hash, Role: "USER", Status: "ACTIVE"})
		log.Println("Seed user: user@school.local / User123!")
	}

	// Dummy reports
	var reportCnt int64
	db.Model(&models.Report{}).Count(&reportCnt)
	if reportCnt == 0 {
		var user models.User
		db.Where("email = ?", "user@school.local").First(&user)
		var catElektronik, catDompet, catKartu models.Category
		db.Where("name = ?", "Elektronik").First(&catElektronik)
		db.Where("name = ?", "Dompet").First(&catDompet)
		db.Where("name = ?", "Kartu").First(&catKartu)
		dummies := []models.Report{
			{UserID: user.ID, CategoryID: catDompet.ID, Type: "LOST", ItemName: "Dompet Hitam Kulit", Description: "Dompet hitam berisi kartu pelajar", Brand: "-", Color: "Hitam", Characteristics: "Ada gantungan kunci bintang", Location: "Kantin", EventDate: "2026-09-10", EventTime: "12:30", Status: "VERIFIED"},
			{UserID: user.ID, CategoryID: catDompet.ID, Type: "FOUND", ItemName: "Dompet Hitam", Description: "Ditemukan di kantin dekat kasir", Brand: "-", Color: "Hitam", Characteristics: "Gantungan bintang", Location: "Kantin", EventDate: "2026-09-10", EventTime: "13:00", StorageLocation: "Ruang TU", Status: "VERIFIED"},
			{UserID: user.ID, CategoryID: catKartu.ID, Type: "LOST", ItemName: "Kartu Pelajar", Description: "Kartu pelajar atas nama Siswa Demo", Brand: "-", Color: "Putih", Characteristics: "Foto berseragam", Location: "Perpustakaan", EventDate: "2026-09-08", Status: "PROCESSING"},
			{UserID: user.ID, CategoryID: catElektronik.ID, Type: "FOUND", ItemName: "Earphone Putih", Description: "Earphone kabel putih", Brand: "Generic", Color: "Putih", Characteristics: "Case retak sedikit", Location: "Lab Komputer", EventDate: "2026-09-09", StorageLocation: "Ruang Lab", Status: "VERIFIED"},
		}
		for _, d := range dummies {
			db.Create(&d)
		}
		db.Create(&models.Notification{UserID: user.ID, Title: "Selamat datang!", Message: "Ini notifikasi contoh. Laporkan barangmu sekarang.", Type: "REPORT_VERIFIED"})
		db.Create(&models.ActivityLog{Action: "seed", Description: "Seed data awal", CreatedAt: time.Now()})
		log.Println("Seed dummy reports done")
	}

	// Default settings
	defaults := map[string]string{
		"school_name": "Sekolah", "app_name": "Lost & Found School",
		"admin_email": "admin@school.local", "max_upload_mb": "5", "default_limit": "10",
	}
	for k, v := range defaults {
		var s models.Setting
		if err := db.Where("`key` = ?", k).First(&s).Error; err != nil {
			db.Create(&models.Setting{Key: k, Value: v})
		}
	}
}
