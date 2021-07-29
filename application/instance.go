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
	e.Router.POST("/apps/upload", uploadAppHandler)
	e.Router.POST("/apps/install", installAppHandler)
	e.Router.POST("/apps/uninstall", uninstallAppHandler)
	e.Router.GET("/app/icon", appIconHandler)
	e.Router.POST("/app/run", startAppHandler)
	e.Router.POST("/app/stop", appStopHandler)
	e.Router.POST("/autoStartApps", appSetAutoStart)
	e.Router.DELETE("/autoStartApps", appRemoveAutoStart)
	e.Router.GET("/disks", getDiskListHandler)
	e.Router.GET("/disk/smart", diskSmartHandler)
	e.Router.POST("/share", createShareHandler)
	e.Router.GET("/share", getShareFolderList)
	e.Router.DELETE("/share", removeShareHandler)
	e.Router.POST("/share/update", updateShareFolder)
	e.Router.POST("/users", createUserHandler)
	e.Router.GET("/users", getUserList)
	e.Router.DELETE("/users", removeUserHandler)
	e.Router.POST("/storage", newStorage)
	e.Router.GET("/storage", getStorageListHandler)
	e.Router.DELETE("/storage", removeStorage)
	e.Router.POST("/zpool", createZFSPoolHandler)
	e.Router.POST("/zpool/conf", createZFSPoolWithNodeHandler)
	e.Router.GET("/zpool", getZFSPoolListHandler)
	e.Router.DELETE("/zpool", removePoolHandler)
	e.Router.POST("/user/auth", generateAuthHandler)
	e.Router.POST("/admin/auth", userLoginHandler)
	e.Router.GET("/user/auth", checkTokenHandler)
	e.Router.GET("/user/share", getUserShareFolderListHandler)
	e.Router.GET("/groups", userGroupListHandler)
	e.Router.POST("/groups", addUserGroup)
	e.Router.DELETE("/groups", removeUserGroup)
	e.Router.GET("/group/{name}", userGroupHandler)
	e.Router.DELETE("/group/{name}/users", removeUserFromUserGroupHandler)
	e.Router.POST("/group/{name}/users", addUserToUserGroupHandler)
	e.Router.POST("/account/password", changeAccountPasswordHandler)
	e.Router.GET("/system/info", getSystemInfoHandler)
	e.Router.GET("/system/monitor", getSystemMonitor)
	e.Router.GET("/tasks", tasksListHandler)
	e.Router.GET("/path/readdir", ReadDirHandler)
	e.Router.GET("/path/realpath", GetRealPathHandler)
	e.Router.GET("/info", serviceInfoHandler)
	e.Router.POST("/os/shutdown", shutdownHandler)
	e.Router.POST("/os/reboot", rebootHandler)
	e.Router.GET("/device/info", deviceInfoHandler)
	e.Router.GET("/network", networkStatusHandler)
	e.Router.AddHandler("/notification", notificationSocketHandler)
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
