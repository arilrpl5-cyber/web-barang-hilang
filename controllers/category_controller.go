package controllers

import (
	"net/http"

	"lost-found-school/models"
	"lost-found-school/services"
	"lost-found-school/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryController struct{ DB *gorm.DB }

func (h *CategoryController) List(c *gin.Context) {
	var cats []models.Category
	h.DB.Order("name ASC").Find(&cats)
	utils.Success(c, http.StatusOK, "Data kategori", cats)
}

func (h *CategoryController) Create(c *gin.Context) {
	u := currentUser(c)
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	if req.Name == "" {
		utils.Fail(c, http.StatusUnprocessableEntity, "Nama kategori wajib diisi")
		return
	}
	cat := models.Category{Name: req.Name, Description: req.Description}
	if err := h.DB.Create(&cat).Error; err != nil {
		utils.Fail(c, http.StatusConflict, "Kategori sudah ada")
		return
	}
	services.LogActivity(h.DB, &u.ID, "create_category", "Membuat kategori: "+req.Name, clientIP(c))
	utils.Success(c, http.StatusCreated, "Kategori berhasil dibuat", cat)
}

func (h *CategoryController) Update(c *gin.Context) {
	u := currentUser(c)
	var cat models.Category
	if err := h.DB.First(&cat, c.Param("id")).Error; err != nil {
		utils.Fail(c, http.StatusNotFound, "Kategori tidak ditemukan")
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	h.DB.Model(&cat).Updates(map[string]interface{}{"name": req.Name, "description": req.Description})
	services.LogActivity(h.DB, &u.ID, "update_category", "Memperbarui kategori ID "+c.Param("id"), clientIP(c))
	utils.Success(c, http.StatusOK, "Kategori berhasil diperbarui", cat)
}

func (h *CategoryController) Delete(c *gin.Context) {
	u := currentUser(c)
	var cat models.Category
	if err := h.DB.First(&cat, c.Param("id")).Error; err != nil {
		utils.Fail(c, http.StatusNotFound, "Kategori tidak ditemukan")
		return
	}
	var cnt int64
	h.DB.Model(&models.Report{}).Where("category_id = ? AND status NOT IN ?", cat.ID, []string{"RETURNED", "REJECTED", "CANCELLED"}).Count(&cnt)
	if cnt > 0 {
		utils.Fail(c, http.StatusConflict, "Kategori masih digunakan laporan aktif")
		return
	}
	h.DB.Delete(&cat)
	services.LogActivity(h.DB, &u.ID, "delete_category", "Menghapus kategori: "+cat.Name, clientIP(c))
	utils.Success(c, http.StatusOK, "Kategori berhasil dihapus", nil)
}
