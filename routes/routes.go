	package routes

import (
	"net/http"

	"lost-found-school/config"
	"lost-found-school/controllers"
	"lost-found-school/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	auth := &controllers.AuthController{DB: db, Cfg: cfg}
	report := &controllers.ReportController{DB: db, Cfg: cfg}
	category := &controllers.CategoryController{DB: db}
	claim := &controllers.ClaimController{DB: db}
	match := &controllers.MatchController{DB: db}
	notif := &controllers.NotificationController{DB: db}
	admin := &controllers.AdminController{DB: db, Cfg: cfg}
	page := &controllers.PageController{DB: db, Cfg: cfg}

	// ---- PUBLIC PAGES ----
	r.GET("/", page.Index)
	r.GET("/login", page.LoginPage)
	r.GET("/register", page.RegisterPage)
	r.GET("/reports", page.Simple("reports.html"))
	r.GET("/reports/lost", page.Simple("reports.html"))
	r.GET("/reports/found", page.Simple("reports.html"))
	r.GET("/reports/create", middleware.AuthRequired(db, cfg), page.Simple("create-report.html"))
	r.GET("/reports/:id", page.ReportDetail)
	r.GET("/about", page.Simple("about.html"))
	r.GET("/help", page.Simple("about.html"))

	// ---- PROTECTED PAGES ----
	authPages := r.Group("/")
	authPages.Use(middleware.AuthRequired(db, cfg))
	{
		authPages.GET("/dashboard", page.Dashboard)
		authPages.GET("/my-reports", page.Simple("my-reports.html"))
		authPages.GET("/claims", page.Simple("claims.html"))
		authPages.GET("/notifications", page.Simple("notifications.html"))
		authPages.GET("/profile", page.Simple("profile.html"))
	}

	// ---- ADMIN PAGES ----
	adminPages := r.Group("/admin")
	adminPages.Use(middleware.AuthRequired(db, cfg), middleware.AdminRequired())
	{
		adminPages.GET("", page.AdminDashboard)
		adminPages.GET("/", page.AdminDashboard)
		adminPages.GET("/reports", page.Simple("admin-reports.html"))
		adminPages.GET("/claims", page.Simple("admin-claims.html"))
		adminPages.GET("/users", page.Simple("admin-users.html"))
		adminPages.GET("/categories", page.Simple("admin-categories.html"))
		adminPages.GET("/matches", page.Simple("admin-matches.html"))
		adminPages.GET("/activity-logs", page.Simple("admin-activity-logs.html"))
		adminPages.GET("/settings", page.Simple("admin-settings.html"))
		adminPages.GET("/notifications", page.Simple("notifications.html"))
	}

	// ---- API AUTH ----
	api := r.Group("/api")
	{
		api.POST("/register", auth.Register)
		api.POST("/login", auth.Login)
		api.POST("/logout", middleware.AuthRequired(db, cfg), auth.Logout)
		api.GET("/profile", middleware.AuthRequired(db, cfg), auth.Profile)
		api.PUT("/profile", middleware.AuthRequired(db, cfg), auth.UpdateProfile)
		api.PUT("/profile/password", middleware.AuthRequired(db, cfg), auth.ChangePassword)

		// reports (public list + detail)
		api.GET("/reports", report.List)
		api.GET("/reports/lost", func(c *gin.Context) {
			c.Request.URL.RawQuery += "&type=LOST"
			if c.Query("type") == "" {
				q := c.Request.URL.Query()
				q.Set("type", "LOST")
				c.Request.URL.RawQuery = q.Encode()
			}
			report.List(c)
		})
		api.GET("/reports/found", func(c *gin.Context) {
			q := c.Request.URL.Query()
			q.Set("type", "FOUND")
			c.Request.URL.RawQuery = q.Encode()
			report.List(c)
		})
		api.GET("/reports/:id", report.Detail)
		api.GET("/my-reports", middleware.AuthRequired(db, cfg), report.MyReports)
		api.POST("/reports", middleware.AuthRequired(db, cfg), report.Create)
		api.PUT("/reports/:id", middleware.AuthRequired(db, cfg), report.Update)
		api.DELETE("/reports/:id", middleware.AuthRequired(db, cfg), report.Delete)

		// categories
		api.GET("/categories", category.List)
		api.POST("/categories", middleware.AuthRequired(db, cfg), middleware.AdminRequired(), category.Create)
		api.PUT("/categories/:id", middleware.AuthRequired(db, cfg), middleware.AdminRequired(), category.Update)
		api.DELETE("/categories/:id", middleware.AuthRequired(db, cfg), middleware.AdminRequired(), category.Delete)

		// claims
		api.POST("/claims", middleware.AuthRequired(db, cfg), claim.Create)
		api.GET("/my-claims", middleware.AuthRequired(db, cfg), claim.MyClaims)
		api.GET("/claims/:id", middleware.AuthRequired(db, cfg), claim.Detail)
		api.GET("/admin/claims", middleware.AuthRequired(db, cfg), middleware.AdminRequired(), claim.AdminList)
		api.PUT("/admin/claims/:id", middleware.AuthRequired(db, cfg), middleware.AdminRequired(), claim.AdminReview)

		// matches
		api.GET("/matches", middleware.AuthRequired(db, cfg), match.List)
		api.GET("/matches/:id", middleware.AuthRequired(db, cfg), match.Detail)
		api.POST("/matches/run", middleware.AuthRequired(db, cfg), middleware.AdminRequired(), match.Run)
		api.PUT("/matches/:id", middleware.AuthRequired(db, cfg), middleware.AdminRequired(), match.Update)
		api.GET("/my-matches", middleware.AuthRequired(db, cfg), match.MyMatches)

		// notifications
		api.GET("/notifications", middleware.AuthRequired(db, cfg), notif.List)
		api.GET("/notifications/unread-count", middleware.AuthRequired(db, cfg), notif.UnreadCount)
		api.PUT("/notifications/:id/read", middleware.AuthRequired(db, cfg), notif.MarkRead)
		api.PUT("/notifications/read-all", middleware.AuthRequired(db, cfg), notif.ReadAll)

		// admin
		adm := api.Group("/admin")
		adm.Use(middleware.AuthRequired(db, cfg), middleware.AdminRequired())
		{
			adm.GET("/dashboard", admin.Dashboard)
			adm.GET("/reports", admin.Reports)
			adm.PUT("/reports/:id/status", admin.UpdateStatus)
			adm.DELETE("/reports/:id", admin.DeleteReport)
			adm.GET("/users", admin.Users)
			adm.PUT("/users/:id", admin.UpdateUser)
			adm.DELETE("/users/:id", admin.DeleteUser)
			adm.GET("/activity-logs", admin.ActivityLogs)
			adm.GET("/settings", admin.GetSettings)
			adm.PUT("/settings", admin.UpdateSettings)
		}

		api.GET("/stats", admin.PublicStats)
	}

	// ---- ERROR PAGES ----
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"title": "404"})
	})
}
