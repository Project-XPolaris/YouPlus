package application

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/middleware"
	"github.com/projectxpolaris/youplus/config"
	"github.com/rs/cors"
	"strings"
)

func RunApplication() {
	e := haruka.NewEngine()
	e.UseMiddleware(middleware.NewLoggerMiddleware())
	e.Router.GET("/apps", appListHandler)
	e.Router.POST("/apps", addAppHandler)
	e.Router.DELETE("/apps", removeAppHandler)
	e.Router.GET("/app/icon", appIconHandler)
	e.Router.POST("/app/run", startAppHandler)
	e.Router.POST("/app/stop", appStopHandler)
	e.Router.POST("/autoStartApps", appSetAutoStart)
	e.Router.DELETE("/autoStartApps", appRemoveAutoStart)
	e.Router.GET("/disks", getDiskListHandler)
	e.Router.POST("/share", createShareHandler)
	e.Router.GET("/share", getShareFolderList)
	e.Router.POST("/users", createUserHandler)
	e.Router.GET("/users", getUserList)
	e.Router.POST("/storage", newStorage)
	e.Router.GET("/storage", getStorageListHandler)
	e.Router.DELETE("/storage", removeStorage)
	e.Router.POST("/zpool", createZFSPoolHandler)
	e.Router.GET("/zpool", getZFSPoolListHandler)
	e.Router.DELETE("/zpool", removePoolHandler)
	e.Router.POST("/user/auth", userLoginHandler)
	e.Router.GET("/user/auth", checkTokenHandler)
	e.UseCors(cors.AllowAll())
	e.UseMiddleware(middleware.NewJWTMiddleware(&middleware.NewJWTMiddlewareOption{
		ReadTokenString: func(ctx *haruka.Context) string {
			rawString := ctx.Request.Header.Get("Authorization")
			rawString = strings.Replace(rawString, "Bearer ", "", 1)
			return rawString
		},
		JWTKey: []byte(config.Config.ApiKey),
	}))
	e.UseMiddleware(&AuthMiddleware{})
	e.RunAndListen(config.Config.Addr)
}
