package controllers

import (
	"net/http"

	"lost-found-school/models"
	"lost-found-school/services"
	"lost-found-school/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MatchController struct{ DB *gorm.DB }

func (h *MatchController) List(c *gin.Context) {
	var matches []models.Match
	q := h.DB.Preload("LostReport").Preload("FoundReport").Order("score DESC")
	if st := c.Query("status"); st != "" {
		q = q.Where("status = ?", st)
	}
	q.Find(&matches)
	utils.Success(c, http.StatusOK, "Data matching", matches)
}

func (h *MatchController) Detail(c *gin.Context) {
	var m models.Match
	if err := h.DB.Preload("LostReport").Preload("FoundReport").First(&m, c.Param("id")).Error; err != nil {
		utils.Fail(c, http.StatusNotFound, "Data tidak ditemukan")
		return
	}
	utils.Success(c, http.StatusOK, "Detail matching", gin.H{
		"match": m, "label": services.Label(m.Score),
		"disclaimer": "Hasil pencocokan merupakan rekomendasi sistem dan tidak menjamin kepemilikan barang.",
	})
}

// POST /api/matches/run — hitung ulang semua pasangan LOST x FOUND yang verified
func (h *MatchController) Run(c *gin.Context) {
	u := currentUser(c)
	created := runMatching(h.DB)
	services.LogActivity(h.DB, &u.ID, "run_matching", "Menjalankan matching, hasil baru terculty", clientIP(c))
	utils.Success(c, http.StatusOK, "Matching selesai", gin.H{"new_matches": created})
}

func (h *MatchController) Update(c *gin.Context) {
	u := currentUser(c)
	var m models.Match
	if err := h.DB.First(&m, c.Param("id")).Error; err != nil {
		utils.Fail(c, http.StatusNotFound, "Data tidak ditemukan")
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	if req.Status != "REVIEWING" && req.Status != "CONFIRMED" && req.Status != "REJECTED" && req.Status != "PENDING" {
		utils.Fail(c, http.StatusBadRequest, "Status tidak valid")
		return
	}
	h.DB.Model(&m).Update("status", req.Status)
	services.LogActivity(h.DB, &u.ID, "update_match", "Mengubah status match ID "+c.Param("id")+" ke "+req.Status, clientIP(c))
	utils.Success(c, http.StatusOK, "Status matching diperbarui", nil)
}

// MyMatches: untuk user — cocokkan laporan miliknya dengan skor >= 40
func (h *MatchController) MyMatches(c *gin.Context) {
	u := currentUser(c)
	var myReports []models.Report
	h.DB.Where("user_id = ? AND type = ?", u.ID, "LOST").Find(&myReports)
	myIDs := []uint{}
	for _, r := range myReports {
		myIDs = append(myIDs, r.ID)
	}
	var matches []models.Match
	if len(myIDs) > 0 {
		h.DB.Preload("LostReport").Preload("FoundReport").
			Where("lost_report_id IN ? AND score >= 40", myIDs).
			Order("score DESC").Limit(10).Find(&matches)
	}
	utils.Success(c, http.StatusOK, "Kemungkinan barang cocok", matches)
}

func runMatching(db *gorm.DB) int {
	var losts []models.Report
	var founds []models.Report
	db.Where("type = ? AND status IN ?", "LOST", []string{"VERIFIED", "PROCESSING", "MATCHED"}).Find(&losts)
	db.Where("type = ? AND status IN ?", "FOUND", []string{"VERIFIED", "PROCESSING", "MATCHED"}).Find(&founds)
	created := 0
	for _, l := range losts {
		for _, f := range founds {
			score := services.ScoreReports(l, f)
			if score < 40 {
				continue
			}
			var cnt int64
			db.Model(&models.Match{}).Where("lost_report_id = ? AND found_report_id = ?", l.ID, f.ID).Count(&cnt)
			if cnt > 0 {
				db.Model(&models.Match{}).Where("lost_report_id = ? AND found_report_id = ?", l.ID, f.ID).Update("score", score)
				continue
			}
			db.Create(&models.Match{LostReportID: l.ID, FoundReportID: f.ID, Score: score, Status: "PENDING"})
			created++
			services.Notify(db, l.UserID, "Kemungkinan cocok ditemukan", "Laporan "+l.ItemName+" kemungkinan cocok dengan "+f.ItemName+" ("+itoa2(int(score))+"%).", "POSSIBLE_MATCH")
		}
	}
	return created
}

func autoMatch(db *gorm.DB, reportID uint) {
	var r models.Report
	if err := db.First(&r, reportID).Error; err != nil {
		return
	}
	var others []models.Report
	if r.Type == "LOST" {
		db.Where("type = ? AND status IN ?", "FOUND", []string{"VERIFIED", "PROCESSING", "MATCHED"}).Find(&others)
		for _, o := range others {
			score := services.ScoreReports(r, o)
			if score >= 40 {
				db.Create(&models.Match{LostReportID: r.ID, FoundReportID: o.ID, Score: score, Status: "PENDING"})
			}
		}
	} else {
		db.Where("type = ? AND status IN ?", "LOST", []string{"VERIFIED", "PROCESSING", "MATCHED"}).Find(&others)
		for _, o := range others {
			score := services.ScoreReports(o, r)
			if score >= 40 {
				db.Create(&models.Match{LostReportID: o.ID, FoundReportID: r.ID, Score: score, Status: "PENDING"})
				services.Notify(db, o.UserID, "Kemungkinan cocok ditemukan", "Laporan "+o.ItemName+" kemungkinan cocok dengan "+r.ItemName+".", "POSSIBLE_MATCH")
			}
		}
	}
}

func itoa2(n int) string { return strconv_Itoa(n) }
