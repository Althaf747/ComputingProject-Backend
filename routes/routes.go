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
		v1.POST("/users/register", controllers.Register)
		v1.POST("/users/login", controllers.Login)
		v1.POST("/users/logout", controllers.Logout)

		protected := v1.Group("/logs")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("", controllers.GetLogs)
			protected.GET("/filter", controllers.GetFilteredLogs) // query : period=today|date|range&date=YYYY-MM-DD&start=YYYY-MM-DD&end=YYYY-MM-DD&name=

			protected.POST("", controllers.CreateLog)
			protected.DELETE("/:id", controllers.DeleteLog)
		}

	}

	r.GET("api/camera", controllers.ProxyWebcam)
	r.GET("/ws", utils.WsHandler)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
