package controllers

import (
	"net/http"

	"lost-found-school/config"
	"lost-found-school/models"
	"lost-found-school/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PageController struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func (h *PageController) Index(c *gin.Context) {
	var recent []models.Report
	h.DB.Preload("Category").Where("status NOT IN ?", []string{"PENDING", "REJECTED"}).Order("created_at DESC").Limit(6).Find(&recent)
	var total, lost, found, returned int64
	h.DB.Model(&models.Report{}).Where("status NOT IN ?", []string{"PENDING", "REJECTED"}).Count(&total)
	h.DB.Model(&models.Report{}).Where("type = ? AND status NOT IN ?", "LOST", []string{"PENDING", "REJECTED"}).Count(&lost)
	h.DB.Model(&models.Report{}).Where("type = ? AND status NOT IN ?", "FOUND", []string{"PENDING", "REJECTED"}).Count(&found)
	h.DB.Model(&models.Report{}).Where("status = ?", "RETURNED").Count(&returned)
	var cats []models.Category
	h.DB.Order("name ASC").Find(&cats)
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Lost & Found School", "recent": recent,
		"total": total, "lost": lost, "found": found, "returned": returned,
		"categories": cats,
	})
}

func (h *PageController) Simple(page string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, page, gin.H{"title": page})
	}
}

// LoginPage: user yang sudah login langsung dilempar ke dashboard (tidak perlu login lagi).
func (h *PageController) LoginPage(c *gin.Context) {
	if dest, ok := h.dashboardIfLoggedIn(c); ok {
		c.Redirect(http.StatusFound, dest)
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{"title": "Login"})
}

// RegisterPage: sama, user yang sudah login tidak perlu daftar lagi.
func (h *PageController) RegisterPage(c *gin.Context) {
	if dest, ok := h.dashboardIfLoggedIn(c); ok {
		c.Redirect(http.StatusFound, dest)
		return
	}
	c.HTML(http.StatusOK, "register.html", gin.H{"title": "Daftar"})
}

func (h *PageController) dashboardIfLoggedIn(c *gin.Context) (string, bool) {
	if h.Cfg == nil {
		return "", false
	}
	tokenStr, err := c.Cookie("token")
	if err != nil || tokenStr == "" {
		return "", false
	}
	claims, err := utils.ParseToken(tokenStr, h.Cfg.JWTSecret)
	if err != nil {
		return "", false
	}
	var user models.User
	if err := h.DB.First(&user, claims.UserID).Error; err != nil || user.Status != "ACTIVE" {
		return "", false
	}
	if user.Role == "ADMIN" {
		return "/admin", true
	}
	return "/dashboard", true
}

func (h *PageController) ReportDetail(c *gin.Context) {
	var r models.Report
	if err := h.DB.Preload("Category").Preload("User").First(&r, c.Param("id")).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"title": "404"})
		return
	}
	r.User.Password = ""
	r.User.Email = ""
	c.HTML(http.StatusOK, "report-detail.html", gin.H{"title": r.ItemName, "report": r})
}

func (h *PageController) Dashboard(c *gin.Context) {
	u, _ := c.Get("user")
	user := u.(models.User)
	var lost, found, processing, done int64
	h.DB.Model(&models.Report{}).Where("user_id = ? AND type = ?", user.ID, "LOST").Count(&lost)
	h.DB.Model(&models.Report{}).Where("user_id = ? AND type = ?", user.ID, "FOUND").Count(&found)
	h.DB.Model(&models.Report{}).Where("user_id = ? AND status IN ?", user.ID, []string{"PENDING", "VERIFIED", "PROCESSING", "MATCHED"}).Count(&processing)
	h.DB.Model(&models.Report{}).Where("user_id = ? AND status = ?", user.ID, "RETURNED").Count(&done)
	var recent []models.Report
	h.DB.Preload("Category").Where("user_id = ?", user.ID).Order("created_at DESC").Limit(5).Find(&recent)
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Dashboard", "user": user,
		"lost": lost, "found": found, "processing": processing, "done": done,
		"recent": recent,
	})
}

func (h *PageController) AdminDashboard(c *gin.Context) {
	u, _ := c.Get("user")
	c.HTML(http.StatusOK, "admin-dashboard.html", gin.H{"title": "Admin Dashboard", "user": u})
}
