package https

import (
	_ "friend-help/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewHTTPServer(httpHandlers *HTTPHandlers, addr string) {
	router := gin.Default()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	apiGroup := router.Group("/api")
	{
		authGroup := apiGroup.Group("/auth")
		{
			authGroup.POST("/reg", httpHandlers.HandlerReg)       // POST /api/auth/reg
			authGroup.POST("/login", httpHandlers.HandlerLogin)   //POST /api/auth/login
			authGroup.POST("/logout", httpHandlers.HandlerLogout) //POST /api/auth/logout
		}
		user := apiGroup.Group("/user")
		user.Use(httpHandlers.AuthMiddleware())
		{
			user.GET("/profile", httpHandlers.HandlerGetProfile)
		}
	}
	router.Run(addr)
}
