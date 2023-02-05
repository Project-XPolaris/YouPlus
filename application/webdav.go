package application

import (
	context2 "context"
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"golang.org/x/net/webdav"
	"net/http"
	"strings"
)

var webdavHandler haruka.RequestHandler = func(context *haruka.Context) {
	context.Writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	// Gets the correct user for this request.
	username, password, ok := context.Request.BasicAuth()
	if !ok {
		http.Error(context.Writer, "Not authorized", 401)
		return
	}

	if _, err := service.GetUserByUsernameWithPassword(username, password); err == nil {
		http.Error(context.Writer, "Not authorized", 401)
		return
	}
	filesystem, err := service.GetFileSystemWithUser(username)
	if err != nil {
		http.Error(context.Writer, "Not authorized", 401)
		return
	}
	//testFilesystem := afero.NewOsFs()

	davfs := &webdav.Handler{
		Prefix:     "/dav",
		FileSystem: &service.YouPlusWebdavFileSystem{Fs: filesystem},
		LockSystem: webdav.NewMemLS(),
	}
	if context.Request.Method == "GET" && strings.HasPrefix(context.Request.URL.Path, davfs.Prefix) {
		info, err := davfs.FileSystem.Stat(context2.TODO(), strings.TrimPrefix(context.Request.URL.Path, davfs.Prefix))
		if err == nil && info.IsDir() {
			context.Request.Method = "PROPFIND"
			if context.Request.Header.Get("Depth") == "" {
				context.Request.Header.Add("Depth", "1")
			}
		}
	}

	davfs.ServeHTTP(context.Writer, context.Request)
}
