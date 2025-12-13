package routes

import (
	"comproBackend/controllers"
	"comproBackend/middleware"
	"comproBackend/utils"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/api")
	{
		user := v1.Group("/users")
		user.POST("/register", controllers.Register)
		user.POST("/login", controllers.Login)
		user.POST("/logout", controllers.Logout)
		user.POST("/reset_request", controllers.ResetPassword)

		userProtected := v1.Group("/users")
		userProtected.Use(middleware.AuthMiddleware())
		{
			userProtected.GET("/pending", controllers.GetPendingAndResetUsers)
			userProtected.POST("/approve", controllers.Approval)
		}

		protected := v1.Group("/logs")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("", controllers.GetLogs)
			protected.GET("/filter", controllers.GetFilteredLogs) // query : period=today|date|range&date=YYYY-MM-DD&start=YYYY-MM-DD&end=YYYY-MM-DD&name=

			protected.POST("", controllers.CreateLog)
			protected.DELETE("/:id", controllers.DeleteLog)
		}

		v1.GET("/camera", controllers.ProxyWebcam)

	}

	r.GET("/ws", utils.WsHandler)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
