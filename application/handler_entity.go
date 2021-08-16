package application

import (
	"errors"
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"net/http"
)

var getEntryByName haruka.RequestHandler = func(context *haruka.Context) {
	name := context.GetQueryString("name")
	entry := service.DefaultRegisterManager.GetOnlineEntryByName(name)
	if entry == nil {
		AbortErrorWithStatus(errors.New("entry not found"), context, http.StatusNotFound)
		return
	}
	template := EntryTemplate{}
	template.Assign(entry)
	context.JSON(template)
}

var getEntityList haruka.RequestHandler = func(context *haruka.Context) {
	data := SerializerEntityList(service.DefaultRegisterManager.Entries)
	context.JSON(haruka.JSON{
		"entities": data,
	})
}
