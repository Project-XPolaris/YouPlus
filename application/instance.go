package application

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/middleware"
	"github.com/rs/cors"
	"youplus/config"
)

func RunApplication() {
	e := haruka.NewEngine()
	e.UseMiddleware(middleware.NewLoggerMiddleware())
	e.Router.GET("/apps", appListHandler)
	e.Router.POST("/app/run", startAppHandler)
	e.Router.POST("/app/stop", appStopHandler)
	e.UseCors(cors.AllowAll())
	e.RunAndListen(config.Config.Addr)
}
