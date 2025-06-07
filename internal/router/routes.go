package router

import (
	"fido2/internal/controller"
	"fido2/internal/usecase/impl"
	"fido2/pkg/middleware"
	"fido2/pkg/utils"
	"github.com/gin-gonic/gin"
	"os"
	"time"
)

// BuildRouter 回傳 *gin.Engine，測試用
func BuildRouter() *gin.Engine {
	logger := utils.GetLogger()

	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode
	}

	authCtl := controller.NewAuthController(impl.GetUserUseCase())

	gin.SetMode(mode)

	app := gin.New()
	app.Use(gin.Logger(), gin.Recovery())

	// Security middlewares
	app.Use(middleware.RateLimit(100, time.Minute))
	app.Use(middleware.CORS())

	// 受信任代理
	if err := app.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		logger.Fatalf("Failed to set trusted proxies: %v", err)
	}

	att := app.Group("/attestation")
	{
		att.POST("/options", authCtl.StartAttestationHandler)
		att.POST("/result", authCtl.FinishAttestationHandler)
	}

	asr := app.Group("/assertion")
	{
		asr.POST("/options", authCtl.StartAssertionHandler)
		asr.POST("/result", authCtl.FinishAssertionHandler)
	}

	wellknown := app.Group("/.well-known")
	{
		wellknown.GET("/apple-app-site-association", controller.AppleWellKnownHandler)
		wellknown.GET("/assetlinks.json", controller.AndroidWellKnownHandler)
	}

	return app
}

func InitRouter() {
	engine := BuildRouter()
	utils.GetLogger().Info("Server started on port 8080")
	if err := engine.Run(":8080"); err != nil {
		panic(err)
	}
}