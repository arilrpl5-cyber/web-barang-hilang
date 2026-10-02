package controllers

import (
	"net/http"
	"strings"

	"lost-found-school/config"
	"lost-found-school/models"
	"lost-found-school/services"
	"lost-found-school/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func currentUser(c *gin.Context) models.User {
	u, _ := c.Get("user")
	return u.(models.User)
}

func clientIP(c *gin.Context) string {
	return c.ClientIP()
}

// POST /api/register
func (h *AuthController) Register(c *gin.Context) {
	var req struct {
		Name            string `json:"name"`
		NisNisn         string `json:"nis_nisn"`
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// try form
		req.Name = c.PostForm("name")
		req.NisNisn = c.PostForm("nis_nisn")
		req.Email = c.PostForm("email")
		req.Password = c.PostForm("password")
		req.ConfirmPassword = c.PostForm("confirm_password")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.NisNisn = strings.TrimSpace(req.NisNisn)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Name == "" || req.NisNisn == "" || req.Email == "" || req.Password == "" {
		utils.Fail(c, http.StatusUnprocessableEntity, "Harap lengkapi semua field")
		return
	}
	if !strings.Contains(req.Email, "@") {
		utils.Fail(c, http.StatusUnprocessableEntity, "Email tidak valid")
		return
	}
	if len(req.Password) < 8 {
		utils.Fail(c, http.StatusUnprocessableEntity, "Password minimal 8 karakter")
		return
	}
	if req.Password != req.ConfirmPassword {
		utils.Fail(c, http.StatusUnprocessableEntity, "Konfirmasi password tidak sama")
		return
	}
	var count int64
	h.DB.Model(&models.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		utils.Fail(c, http.StatusConflict, "Email sudah terdaftar")
		return
	}
	h.DB.Model(&models.User{}).Where("nis_nisn = ?", req.NisNisn).Count(&count)
	if count > 0 {
		utils.Fail(c, http.StatusConflict, "NIS/NISN sudah terdaftar")
		return
	}
	hash, _ := utils.HashPassword(req.Password)
	user := models.User{Name: req.Name, NisNisn: req.NisNisn, Email: req.Email, Password: hash, Role: "USER", Status: "ACTIVE"}
	if err := h.DB.Create(&user).Error; err != nil {
		utils.Fail(c, http.StatusInternalServerError, "Gagal mendaftar")
		return
	}
	services.Notify(h.DB, user.ID, "Selamat datang!", "Akun kamu berhasil dibuat. Silakan laporkan atau cari barang.", "REPORT_VERIFIED")
	services.LogActivity(h.DB, &user.ID, "register", "User mendaftar: "+user.Email, clientIP(c))
	utils.Success(c, http.StatusCreated, "Pendaftaran berhasil, silakan login", gin.H{"id": user.ID})
}

// POST /api/login
func (h *AuthController) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Email = c.PostForm("email")
		req.Password = c.PostForm("password")
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		utils.Fail(c, http.StatusBadRequest, "Email dan password wajib diisi")
		return
	}
	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		utils.Fail(c, http.StatusUnauthorized, "Email atau password salah")
		return
	}
	if user.Status != "ACTIVE" {
		utils.Fail(c, http.StatusForbidden, "Akun tidak aktif")
		return
	}
	if !utils.CheckPassword(user.Password, req.Password) {
		utils.Fail(c, http.StatusUnauthorized, "Email atau password salah")
		return
	}
	token, _ := utils.GenerateToken(user.ID, user.Role, h.Cfg.JWTSecret)
	c.SetCookie("token", token, 86400, "/", "", false, true)
	services.LogActivity(h.DB, &user.ID, "login", "User login: "+user.Email, clientIP(c))
	utils.Success(c, http.StatusOK, "Login berhasil", gin.H{
		"token": token,
		"user":  gin.H{"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role},
	})
}

// POST /api/logout
func (h *AuthController) Logout(c *gin.Context) {
	u := currentUser(c)
	services.LogActivity(h.DB, &u.ID, "logout", "User logout: "+u.Email, clientIP(c))
	c.SetCookie("token", "", -1, "/", "", false, true)
	utils.Success(c, http.StatusOK, "Logout berhasil", nil)
}

// GET /api/profile
func (h *AuthController) Profile(c *gin.Context) {
	u := currentUser(c)
	utils.Success(c, http.StatusOK, "Profil ditemukan", gin.H{
		"id": u.ID, "name": u.Name, "nis_nisn": u.NisNisn, "email": u.Email,
		"role": u.Role, "status": u.Status, "profile_photo": u.ProfilePhoto,
		"created_at": u.CreatedAt,
	})
}

// PUT /api/profile
func (h *AuthController) UpdateProfile(c *gin.Context) {
	u := currentUser(c)
	var req struct {
		Name    string `json:"name"`
		NisNisn string `json:"nis_nisn"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	updates := map[string]interface{}{}
	if strings.TrimSpace(req.Name) != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.NisNisn) != "" && req.NisNisn != u.NisNisn {
		var cnt int64
		h.DB.Model(&models.User{}).Where("nis_nisn = ? AND id <> ?", req.NisNisn, u.ID).Count(&cnt)
		if cnt > 0 {
			utils.Fail(c, http.StatusConflict, "NIS/NISN sudah digunakan")
			return
		}
		updates["nis_nisn"] = strings.TrimSpace(req.NisNisn)
	}
	if len(updates) == 0 {
		utils.Fail(c, http.StatusBadRequest, "Tidak ada perubahan")
		return
	}
	h.DB.Model(&models.User{}).Where("id = ?", u.ID).Updates(updates)
	services.LogActivity(h.DB, &u.ID, "update_profile", "User memperbarui profil", clientIP(c))
	utils.Success(c, http.StatusOK, "Profil berhasil diperbarui", nil)
}

// PUT /api/profile/password
func (h *AuthController) ChangePassword(c *gin.Context) {
	u := currentUser(c)
	var req struct {
		OldPassword     string `json:"old_password"`
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	var user models.User
	h.DB.First(&user, u.ID)
	if !utils.CheckPassword(user.Password, req.OldPassword) {
		utils.Fail(c, http.StatusBadRequest, "Password lama salah")
		return
	}
	if len(req.NewPassword) < 8 {
		utils.Fail(c, http.StatusUnprocessableEntity, "Password baru minimal 8 karakter")
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		utils.Fail(c, http.StatusUnprocessableEntity, "Konfirmasi password tidak sama")
		return
	}
	hash, _ := utils.HashPassword(req.NewPassword)
	h.DB.Model(&user).Update("password", hash)
	services.LogActivity(h.DB, &u.ID, "change_password", "User mengganti password", clientIP(c))
	utils.Success(c, http.StatusOK, "Password berhasil diubah", nil)
}
