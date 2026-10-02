package controllers

import (
	"net/http"
	"time"

	"lost-found-school/models"
	"lost-found-school/services"
	"lost-found-school/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ClaimController struct{ DB *gorm.DB }

func (h *ClaimController) Create(c *gin.Context) {
	u := currentUser(c)
	var req struct {
		ReportID    uint   `json:"report_id"`
		Description string `json:"description"`
		Evidence    string `json:"evidence"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	if req.ReportID == 0 || req.Description == "" {
		utils.Fail(c, http.StatusUnprocessableEntity, "Harap lengkapi semua field")
		return
	}
	var r models.Report
	if err := h.DB.First(&r, req.ReportID).Error; err != nil {
		utils.Fail(c, http.StatusNotFound, "Laporan tidak ditemukan")
		return
	}
	if r.Type != "FOUND" {
		utils.Fail(c, http.StatusBadRequest, "Hanya barang ditemukan yang dapat diklaim")
		return
	}
	if r.UserID == u.ID {
		utils.Fail(c, http.StatusBadRequest, "Tidak dapat mengklaim laporan sendiri")
		return
	}
	if r.Status == "RETURNED" {
		utils.Fail(c, http.StatusBadRequest, "Barang sudah dikembalikan")
		return
	}
	var cnt int64
	h.DB.Model(&models.Claim{}).Where("report_id = ? AND user_id = ? AND status IN ?", req.ReportID, u.ID, []string{"PENDING", "REVIEWING"}).Count(&cnt)
	if cnt > 0 {
		utils.Fail(c, http.StatusConflict, "Kamu sudah mengajukan klaim untuk barang ini")
		return
	}
	claim := models.Claim{ReportID: req.ReportID, UserID: u.ID, Description: req.Description, Evidence: req.Evidence, Status: "PENDING"}
	h.DB.Create(&claim)
	// notify admins
	var admins []models.User
	h.DB.Where("role = ?", "ADMIN").Find(&admins)
	for _, a := range admins {
		services.Notify(h.DB, a.ID, "Klaim baru", "Klaim untuk "+r.ItemName+" perlu direview.", "CLAIM_SUBMITTED")
	}
	services.Notify(h.DB, u.ID, "Klaim terkirim", "Klaim kamu untuk "+r.ItemName+" sedang direview admin.", "CLAIM_SUBMITTED")
	services.LogActivity(h.DB, &u.ID, "create_claim", "Mengajukan klaim untuk report ID "+itoa(req.ReportID), clientIP(c))
	utils.Success(c, http.StatusCreated, "Klaim berhasil dikirim", claim)
}

func (h *ClaimController) MyClaims(c *gin.Context) {
	u := currentUser(c)
	var claims []models.Claim
	h.DB.Preload("Report").Where("user_id = ?", u.ID).Order("created_at DESC").Find(&claims)
	utils.Success(c, http.StatusOK, "Data klaim saya", claims)
}

func (h *ClaimController) Detail(c *gin.Context) {
	u := currentUser(c)
	var claim models.Claim
	if err := h.DB.Preload("Report").Preload("User").First(&claim, c.Param("id")).Error; err != nil {
		utils.Fail(c, http.StatusNotFound, "Klaim tidak ditemukan")
		return
	}
	if claim.UserID != u.ID && u.Role != "ADMIN" {
		utils.Fail(c, http.StatusForbidden, "Tidak boleh melihat klaim orang lain")
		return
	}
	utils.Success(c, http.StatusOK, "Detail klaim", claim)
}

func (h *ClaimController) AdminList(c *gin.Context) {
	var claims []models.Claim
	q := h.DB.Preload("Report").Preload("User").Order("created_at DESC")
	if st := c.Query("status"); st != "" {
		q = q.Where("claims.status = ?", st)
	}
	q.Find(&claims)
	utils.Success(c, http.StatusOK, "Data klaim", claims)
}

func (h *ClaimController) AdminReview(c *gin.Context) {
	u := currentUser(c)
	var claim models.Claim
	if err := h.DB.Preload("Report").First(&claim, c.Param("id")).Error; err != nil {
		utils.Fail(c, http.StatusNotFound, "Klaim tidak ditemukan")
		return
	}
	var req struct {
		Status    string `json:"status"` // APPROVED, REJECTED, REVIEWING
		AdminNote string `json:"admin_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	if req.Status != "APPROVED" && req.Status != "REJECTED" && req.Status != "REVIEWING" {
		utils.Fail(c, http.StatusBadRequest, "Status tidak valid")
		return
	}
	if req.Status == "REJECTED" && req.AdminNote == "" {
		utils.Fail(c, http.StatusUnprocessableEntity, "Alasan penolakan wajib diisi")
		return
	}
	if req.Status == "APPROVED" {
		var other int64
		h.DB.Model(&models.Claim{}).Where("report_id = ? AND status = ? AND id <> ?", claim.ReportID, "APPROVED", claim.ID).Count(&other)
		if other > 0 {
			utils.Fail(c, http.StatusConflict, "Sudah ada klaim APPROVED untuk laporan ini")
			return
		}
	}
	now := time.Now()
	h.DB.Model(&claim).Updates(map[string]interface{}{
		"status": req.Status, "admin_note": req.AdminNote, "reviewed_by": u.ID, "reviewed_at": &now,
	})
	if req.Status == "APPROVED" {
		h.DB.Model(&models.Report{}).Where("id = ?", claim.ReportID).Update("status", "RETURNED")
		services.Notify(h.DB, claim.UserID, "Klaim disetujui", "Klaim kamu disetujui. Barang akan dikembalikan.", "CLAIM_APPROVED")
		var r models.Report
		h.DB.First(&r, claim.ReportID)
		services.Notify(h.DB, r.UserID, "Barang dikembalikan", "Laporan "+r.ItemName+" telah selesai (RETURNED).", "REPORT_RETURNED")
		services.LogActivity(h.DB, &u.ID, "approve_claim", "Menyetujui klaim ID "+c.Param("id"), clientIP(c))
	} else if req.Status == "REJECTED" {
		services.Notify(h.DB, claim.UserID, "Klaim ditolak", "Alasan: "+req.AdminNote, "CLAIM_REJECTED")
		services.LogActivity(h.DB, &u.ID, "reject_claim", "Menolak klaim ID "+c.Param("id"), clientIP(c))
	}
	utils.Success(c, http.StatusOK, "Klaim berhasil direview", nil)
}

func itoa(n uint) string {
	return strconv_Itoa(int(n))
}

func strconv_Itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	b := []byte{}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
