package controllers

import (
	"net/http"
	"time"

	"lost-found-school/config"
	"lost-found-school/models"
	"lost-found-school/services"
	"lost-found-school/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminController struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func (h *AdminController) Dashboard(c *gin.Context) {
	var totalUsers, totalReports, lost, found, returned, pending, claims int64
	h.DB.Model(&models.User{}).Count(&totalUsers)
	h.DB.Model(&models.Report{}).Count(&totalReports)
	h.DB.Model(&models.Report{}).Where("type = ?", "LOST").Count(&lost)
	h.DB.Model(&models.Report{}).Where("type = ?", "FOUND").Count(&found)
	h.DB.Model(&models.Report{}).Where("status = ?", "RETURNED").Count(&returned)
	h.DB.Model(&models.Report{}).Where("status = ?", "PENDING").Count(&pending)
	h.DB.Model(&models.Claim{}).Where("status IN ?", []string{"PENDING", "REVIEWING"}).Count(&claims)

	var recent []models.Report
	h.DB.Preload("Category").Preload("User").Order("created_at DESC").Limit(5).Find(&recent)
	var pendingClaims []models.Claim
	h.DB.Preload("Report").Preload("User").Where("status IN ?", []string{"PENDING", "REVIEWING"}).Order("created_at DESC").Limit(5).Find(&pendingClaims)
	var activity []models.ActivityLog
	h.DB.Preload("User").Order("created_at DESC").Limit(8).Find(&activity)

	// per-month counts (sqlite + mysql compatible: group in Go)
	type monthCount struct {
		Month string
		Total int64
	}
	var reports []models.Report
	h.DB.Find(&reports)
	monthMap := map[string]int64{}
	for _, r := range reports {
		m := r.CreatedAt.Format("2006-01")
		monthMap[m]++
	}

	utils.Success(c, http.StatusOK, "Admin dashboard", gin.H{
		"stats": gin.H{
			"total_users": totalUsers, "total_reports": totalReports,
			"lost": lost, "found": found, "returned": returned,
			"pending": pending, "claims": claims,
		},
		"recent_reports":  recent,
		"pending_claims":  pendingClaims,
		"recent_activity": activity,
		"per_month":       monthMap,
	})
}

func (h *AdminController) Reports(c *gin.Context) {
	page, limit, offset := utils.Pagination(c, h.Cfg.DefaultLimit)
	q := h.DB.Model(&models.Report{}).Preload("Category").Preload("User")
	if t := c.Query("type"); t != "" {
		q = q.Where("reports.type = ?", t)
	}
	if st := c.Query("status"); st != "" {
		q = q.Where("reports.status = ?", st)
	}
	if s := c.Query("search"); s != "" {
		like := "%" + s + "%"
		q = q.Where("(reports.item_name LIKE ? OR reports.location LIKE ?)", like, like)
	}
	var total int64
	q.Count(&total)
	var reports []models.Report
	q.Order("reports.created_at DESC").Offset(offset).Limit(limit).Find(&reports)
	utils.Paginated(c, "Data laporan (admin)", reports, total, page, limit)
}

func (h *AdminController) UpdateStatus(c *gin.Context) {
	u := currentUser(c)
	var r models.Report
	if err := h.DB.First(&r, c.Param("id")).Error; err != nil {
		utils.Fail(c, http.StatusNotFound, "Laporan tidak ditemukan")
		return
	}
	var req struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	valid := map[string]bool{"PENDING": true, "VERIFIED": true, "PROCESSING": true, "MATCHED": true, "RETURNED": true, "REJECTED": true, "CANCELLED": true}
	if !valid[req.Status] {
		utils.Fail(c, http.StatusBadRequest, "Status tidak valid")
		return
	}
	if req.Status == "REJECTED" && req.Reason == "" {
		utils.Fail(c, http.StatusUnprocessableEntity, "Alasan penolakan wajib diisi")
		return
	}
	updates := map[string]interface{}{"status": req.Status}
	if req.Status == "REJECTED" {
		updates["reject_reason"] = req.Reason
	}
	h.DB.Model(&r).Updates(updates)
	if req.Status == "VERIFIED" {
		services.Notify(h.DB, r.UserID, "Laporan terverifikasi", "Laporan "+r.ItemName+" telah diverifikasi admin.", "REPORT_VERIFIED")
		services.LogActivity(h.DB, &u.ID, "verify_report", "Verifikasi laporan ID "+c.Param("id"), clientIP(c))
		go autoMatch(h.DB, r.ID)
	} else if req.Status == "REJECTED" {
		services.Notify(h.DB, r.UserID, "Laporan ditolak", "Alasan: "+req.Reason, "REPORT_REJECTED")
		services.LogActivity(h.DB, &u.ID, "reject_report", "Menolak laporan ID "+c.Param("id"), clientIP(c))
	} else {
		services.LogActivity(h.DB, &u.ID, "update_report_status", "Status laporan ID "+c.Param("id")+" -> "+req.Status, clientIP(c))
	}
	utils.Success(c, http.StatusOK, "Status laporan diperbarui", nil)
}

func (h *AdminController) DeleteReport(c *gin.Context) {
	u := currentUser(c)
	h.DB.Delete(&models.Report{}, c.Param("id"))
	services.LogActivity(h.DB, &u.ID, "delete_report", "Admin menghapus laporan ID "+c.Param("id"), clientIP(c))
	utils.Success(c, http.StatusOK, "Laporan dihapus", nil)
}

func (h *AdminController) Users(c *gin.Context) {
	page, limit, offset := utils.Pagination(c, h.Cfg.DefaultLimit)
	q := h.DB.Model(&models.User{})
	if r := c.Query("role"); r != "" {
		q = q.Where("role = ?", r)
	}
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	if s := c.Query("search"); s != "" {
		like := "%" + s + "%"
		q = q.Where("(name LIKE ? OR email LIKE ? OR nis_nisn LIKE ?)", like, like, like)
	}
	var total int64
	q.Count(&total)
	var users []models.User
	q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users)
	for i := range users {
		users[i].Password = ""
	}
	utils.Paginated(c, "Data pengguna", users, total, page, limit)
}

func (h *AdminController) UpdateUser(c *gin.Context) {
	u := currentUser(c)
	var target models.User
	if err := h.DB.First(&target, c.Param("id")).Error; err != nil {
		utils.Fail(c, http.StatusNotFound, "User tidak ditemukan")
		return
	}
	var req struct {
		Role   string `json:"role"`
		Status string `json:"status"`
		Name   string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	updates := map[string]interface{}{}
	if req.Role == "USER" || req.Role == "ADMIN" {
		updates["role"] = req.Role
	}
	if req.Status == "ACTIVE" || req.Status == "INACTIVE" || req.Status == "BANNED" {
		updates["status"] = req.Status
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if len(updates) == 0 {
		utils.Fail(c, http.StatusBadRequest, "Tidak ada perubahan")
		return
	}
	h.DB.Model(&target).Updates(updates)
	services.LogActivity(h.DB, &u.ID, "update_user", "Memperbarui user ID "+c.Param("id"), clientIP(c))
	utils.Success(c, http.StatusOK, "User berhasil diperbarui", nil)
}

func (h *AdminController) DeleteUser(c *gin.Context) {
	u := currentUser(c)
	if itoa(uint(0)) == "" {
		_ = time.Now()
	}
	id := c.Param("id")
	if itoa(u.ID) == id {
		utils.Fail(c, http.StatusBadRequest, "Tidak dapat menghapus akun sendiri")
		return
	}
	h.DB.Delete(&models.User{}, id)
	services.LogActivity(h.DB, &u.ID, "delete_user", "Menghapus user ID "+id, clientIP(c))
	utils.Success(c, http.StatusOK, "User dihapus", nil)
}

func (h *AdminController) ActivityLogs(c *gin.Context) {
	page, limit, offset := utils.Pagination(c, h.Cfg.DefaultLimit)
	var total int64
	h.DB.Model(&models.ActivityLog{}).Count(&total)
	var logs []models.ActivityLog
	h.DB.Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs)
	utils.Paginated(c, "Activity logs", logs, total, page, limit)
}

func (h *AdminController) GetSettings(c *gin.Context) {
	var settings []models.Setting
	h.DB.Find(&settings)
	m := gin.H{}
	for _, s := range settings {
		m[s.Key] = s.Value
	}
	utils.Success(c, http.StatusOK, "Pengaturan", m)
}

func (h *AdminController) UpdateSettings(c *gin.Context) {
	u := currentUser(c)
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	for k, v := range req {
		var s models.Setting
		if err := h.DB.Where("`key` = ?", k).First(&s).Error; err != nil {
			h.DB.Create(&models.Setting{Key: k, Value: v})
		} else {
			h.DB.Model(&s).Update("value", v)
		}
	}
	services.LogActivity(h.DB, &u.ID, "update_settings", "Memperbarui pengaturan", clientIP(c))
	utils.Success(c, http.StatusOK, "Pengaturan disimpan", nil)
}

func (h *AdminController) PublicStats(c *gin.Context) {
	var total, lost, found, returned int64
	h.DB.Model(&models.Report{}).Where("status NOT IN ?", []string{"PENDING", "REJECTED"}).Count(&total)
	h.DB.Model(&models.Report{}).Where("type = ? AND status NOT IN ?", "LOST", []string{"PENDING", "REJECTED"}).Count(&lost)
	h.DB.Model(&models.Report{}).Where("type = ? AND status NOT IN ?", "FOUND", []string{"PENDING", "REJECTED"}).Count(&found)
	h.DB.Model(&models.Report{}).Where("status = ?", "RETURNED").Count(&returned)
	utils.Success(c, http.StatusOK, "Statistik publik", gin.H{
		"total": total, "lost": lost, "found": found, "returned": returned,
	})
}
