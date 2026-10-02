package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"lost-found-school/config"
	"lost-found-school/models"
	"lost-found-school/services"
	"lost-found-school/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReportController struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func (h *ReportController) List(c *gin.Context) {
	page, limit, offset := utils.Pagination(c, h.Cfg.DefaultLimit)
	q := h.DB.Model(&models.Report{}).Preload("Category").Preload("User")

	if t := strings.ToUpper(c.Query("type")); t == "LOST" || t == "FOUND" {
		q = q.Where("reports.type = ?", t)
	}
	if cid := c.Query("category_id"); cid != "" {
		q = q.Where("reports.category_id = ?", cid)
	}
	if st := strings.ToUpper(c.Query("status")); st != "" {
		q = q.Where("reports.status = ?", st)
	}
	if loc := c.Query("location"); loc != "" {
		q = q.Where("reports.location LIKE ?", "%"+loc+"%")
	}
	if s := c.Query("search"); s != "" {
		like := "%" + s + "%"
		q = q.Where("(reports.item_name LIKE ? OR reports.description LIKE ? OR reports.brand LIKE ? OR reports.color LIKE ? OR reports.characteristics LIKE ?)", like, like, like, like, like)
	}
	// Public only sees VERIFIED+ unless owner/admin — keep simple: hide PENDING/REJECTED from guests
	_, hasToken := c.Cookie("token")
	if hasToken != nil && c.GetHeader("Authorization") == "" {
		// guest
		if c.Query("status") == "" && c.Query("mine") == "" {
			q = q.Where("reports.status NOT IN ?", []string{"PENDING", "REJECTED"})
		}
	}

	var total int64
	q.Count(&total)

	sort := c.DefaultQuery("sort", "newest")
	switch sort {
	case "oldest":
		q = q.Order("reports.created_at ASC")
	case "az":
		q = q.Order("reports.item_name ASC")
	case "za":
		q = q.Order("reports.item_name DESC")
	default:
		q = q.Order("reports.created_at DESC")
	}

	var reports []models.Report
	q.Offset(offset).Limit(limit).Find(&reports)
	// hide sensitive user fields
	for i := range reports {
		reports[i].User.Password = ""
		reports[i].User.NisNisn = ""
	}
	utils.Paginated(c, "Data laporan", reports, total, page, limit)
}

func (h *ReportController) Detail(c *gin.Context) {
	id := c.Param("id")
	var r models.Report
	if err := h.DB.Preload("Category").Preload("User").First(&r, id).Error; err != nil {
		utils.Fail(c, http.StatusNotFound, "Laporan tidak ditemukan")
		return
	}
	r.User.Password = ""
	utils.Success(c, http.StatusOK, "Detail laporan", r)
}

func (h *ReportController) Create(c *gin.Context) {
	u := currentUser(c)
	reportType := strings.ToUpper(c.PostForm("type"))
	if reportType != "LOST" && reportType != "FOUND" {
		// try JSON
		var j struct {
			Type string `json:"type"`
		}
		_ = c.ShouldBindJSON(&j)
		if j.Type != "" {
			reportType = strings.ToUpper(j.Type)
		}
	}
	if reportType != "LOST" && reportType != "FOUND" {
		utils.Fail(c, http.StatusBadRequest, "Jenis laporan harus LOST atau FOUND")
		return
	}
	// support both multipart and JSON
	itemName := c.PostForm("item_name")
	categoryIDStr := c.PostForm("category_id")
	description := c.PostForm("description")
	brand := c.PostForm("brand")
	color := c.PostForm("color")
	characteristics := c.PostForm("characteristics")
	location := c.PostForm("location")
	eventDate := c.PostForm("event_date")
	eventTime := c.PostForm("event_time")
	storage := c.PostForm("storage_location")
	note := c.PostForm("additional_note")
	if itemName == "" {
		var j struct {
			ItemName        string `json:"item_name"`
			CategoryID      uint   `json:"category_id"`
			Description     string `json:"description"`
			Brand           string `json:"brand"`
			Color           string `json:"color"`
			Characteristics string `json:"characteristics"`
			Location        string `json:"location"`
			EventDate       string `json:"event_date"`
			EventTime       string `json:"event_time"`
			StorageLocation string `json:"storage_location"`
			AdditionalNote  string `json:"additional_note"`
		}
		if err := c.ShouldBindJSON(&j); err == nil && j.ItemName != "" {
			itemName = j.ItemName
			description = j.Description
			brand = j.Brand
			color = j.Color
			characteristics = j.Characteristics
			location = j.Location
			eventDate = j.EventDate
			eventTime = j.EventTime
			storage = j.StorageLocation
			note = j.AdditionalNote
			categoryIDStr = strconv.Itoa(int(j.CategoryID))
		}
	}
	if strings.TrimSpace(itemName) == "" {
		utils.Fail(c, http.StatusUnprocessableEntity, "Nama barang wajib diisi")
		return
	}
	var categoryID uint
	if categoryIDStr != "" {
		if n, err := strconv.Atoi(categoryIDStr); err == nil {
			categoryID = uint(n)
		}
	}
	photo := ""
	if file, err := c.FormFile("photo"); err == nil && file != nil {
		p, err := utils.SaveUpload(file, "uploads", h.Cfg.MaxUploadMB)
		if err != nil {
			utils.Fail(c, http.StatusBadRequest, err.Error())
			return
		}
		photo = p
	}
	r := models.Report{
		UserID: u.ID, CategoryID: categoryID, Type: reportType,
		ItemName: strings.TrimSpace(itemName), Description: description,
		Photo: photo, Brand: brand, Color: color, Characteristics: characteristics,
		Location: location, EventDate: eventDate, EventTime: eventTime,
		StorageLocation: storage, AdditionalNote: note, Status: "PENDING",
	}
	if err := h.DB.Create(&r).Error; err != nil {
		utils.Fail(c, http.StatusInternalServerError, "Gagal menyimpan laporan")
		return
	}
	services.Notify(h.DB, u.ID, "Laporan diterima", "Laporan "+r.ItemName+" sedang diperiksa admin.", "REPORT_VERIFIED")
	services.LogActivity(h.DB, &u.ID, "create_report", "Membuat laporan "+reportType+": "+r.ItemName, clientIP(c))
	// auto matching trigger (async best-effort)
	go autoMatch(h.DB, r.ID)
	utils.Success(c, http.StatusCreated, "Laporan berhasil dibuat", r)
}

func (h *ReportController) Update(c *gin.Context) {
	u := currentUser(c)
	id := c.Param("id")
	var r models.Report
	if err := h.DB.First(&r, id).Error; err != nil {
		utils.Fail(c, http.StatusNotFound, "Laporan tidak ditemukan")
		return
	}
	if r.UserID != u.ID && u.Role != "ADMIN" {
		utils.Fail(c, http.StatusForbidden, "Tidak boleh mengubah laporan orang lain")
		return
	}
	if r.Status == "RETURNED" || r.Status == "REJECTED" {
		if u.Role != "ADMIN" {
			utils.Fail(c, http.StatusForbidden, "Laporan "+r.Status+" tidak dapat diubah")
			return
		}
	}
	var req map[string]interface{}
	ct := c.ContentType()
	if strings.Contains(ct, "multipart") || c.PostForm("item_name") != "" {
		req = map[string]interface{}{}
		for _, k := range []string{"item_name", "description", "brand", "color", "characteristics", "location", "event_date", "event_time", "storage_location", "additional_note"} {
			if v := c.PostForm(k); v != "" {
				dbk := k
				switch k {
				case "event_date":
					dbk = "event_date"
				}
				req[dbk] = v
			}
		}
		if v := c.PostForm("category_id"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				req["category_id"] = n
			}
		}
		if file, err := c.FormFile("photo"); err == nil && file != nil {
			p, err := utils.SaveUpload(file, "uploads", h.Cfg.MaxUploadMB)
			if err != nil {
				utils.Fail(c, http.StatusBadRequest, err.Error())
				return
			}
			req["photo"] = p
		}
	} else {
		var j map[string]interface{}
		if err := c.ShouldBindJSON(&j); err != nil {
			utils.Fail(c, http.StatusBadRequest, "Data tidak valid")
			return
		}
		allowed := map[string]bool{"item_name": true, "category_id": true, "description": true, "brand": true, "color": true, "characteristics": true, "location": true, "event_date": true, "event_time": true, "storage_location": true, "additional_note": true, "photo": true}
		req = map[string]interface{}{}
		for k, v := range j {
			if allowed[k] {
				req[k] = v
			}
		}
	}
	if len(req) == 0 {
		utils.Fail(c, http.StatusBadRequest, "Tidak ada perubahan")
		return
	}
	h.DB.Model(&r).Updates(req)
	services.LogActivity(h.DB, &u.ID, "update_report", "Memperbarui laporan ID "+id, clientIP(c))
	utils.Success(c, http.StatusOK, "Laporan berhasil diperbarui", nil)
}

func (h *ReportController) Delete(c *gin.Context) {
	u := currentUser(c)
	id := c.Param("id")
	var r models.Report
	if err := h.DB.First(&r, id).Error; err != nil {
		utils.Fail(c, http.StatusNotFound, "Laporan tidak ditemukan")
		return
	}
	if r.UserID != u.ID && u.Role != "ADMIN" {
		utils.Fail(c, http.StatusForbidden, "Tidak boleh menghapus laporan orang lain")
		return
	}
	h.DB.Delete(&r)
	services.LogActivity(h.DB, &u.ID, "delete_report", "Menghapus laporan ID "+id, clientIP(c))
	utils.Success(c, http.StatusOK, "Laporan berhasil dihapus", nil)
}

func (h *ReportController) MyReports(c *gin.Context) {
	u := currentUser(c)
	page, limit, offset := utils.Pagination(c, h.Cfg.DefaultLimit)
	var total int64
	h.DB.Model(&models.Report{}).Where("user_id = ?", u.ID).Count(&total)
	var reports []models.Report
	h.DB.Preload("Category").Where("user_id = ?", u.ID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&reports)
	utils.Paginated(c, "Laporan saya", reports, total, page, limit)
}
