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
			userProtected.POST("/fcm-token", controllers.UpdateFCMToken)
		}

		protected := v1.Group("/logs")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("", controllers.GetLogs)
			protected.GET("/filter", controllers.GetFilteredLogs) // query : period=today|date|range&date=YYYY-MM-DD&start=YYYY-MM-DD&end=YYYY-MM-DD&name=

			protected.POST("", controllers.CreateLog)
			protected.DELETE("/:id", controllers.DeleteLog)
		}

		camera := v1.Group("/camera")
		{
			camera.GET("/stream", controllers.ProxyCameraStream)
			camera.GET("/snapshot", controllers.ProxyCameraSnapshot)
			camera.GET("/events", controllers.ProxyCameraEvents)
			camera.GET("/status", controllers.GetCameraStatus)
			camera.POST("/start", controllers.StartCamera)
			camera.POST("/stop", controllers.StopCamera)
			camera.GET("/config", controllers.GetCameraConfig)
			camera.POST("/config", controllers.UpdateCameraConfig)
			camera.GET("/zones", controllers.GetCameraZones)
			camera.POST("/zones", controllers.UpdateCameraZones)
			camera.POST("/test/droidcam", controllers.SetupDroidCam)
			camera.POST("/test/rtsp", controllers.SetupRTSP)
		}

		faces := v1.Group("/faces")
		{
			faces.GET("", controllers.ListFaceUsers)
			faces.POST("/enroll", controllers.EnrollFaceUser)
			faces.POST("/enroll/capture", controllers.EnrollFaceUserCapture)
			faces.DELETE("/:name", controllers.DeleteFaceUser)
			faces.POST("/:name/add-sample", controllers.AddFaceSample)
		}
	}

	v1.GET("/ws", utils.WsHandler)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
