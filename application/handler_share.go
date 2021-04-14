package application

import (
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/service"
	"github.com/projectxpolaris/youplus/yousmb"
	"net/http"
	"strings"
)

var createShareHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody service.NewShareFolderOption
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.CreateNewShareFolder(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var getShareFolderList haruka.RequestHandler = func(context *haruka.Context) {
	folderList, err := service.GetShareFolders()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	smbConfig, err := yousmb.DefaultClient.GetConfig()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	shareFolders := make([]ShareFolderTemplate, 0)
	for _, shareFolderConfig := range folderList {
		template := ShareFolderTemplate{
			Id:   shareFolderConfig.ID,
			Name: shareFolderConfig.Name,
		}
		sid := ""
		if len(shareFolderConfig.PartStorageId) != 0 {
			sid = shareFolderConfig.PartStorageId
		}
		if len(shareFolderConfig.ZFSStorageId) != 0 {
			sid = shareFolderConfig.ZFSStorageId
		}
		storage := service.DefaultStoragePool.GetStorageById(sid)
		if storage == nil {
			continue
		}
		storageTemplate := StorageTemplate{}
		storageTemplate.Assign(storage)
		template.Storage = storageTemplate
		// get config
		var targetSection *yousmb.SMBSection
		for _, section := range smbConfig.Sections {
			if section.Name == shareFolderConfig.Name {
				targetSection = &section
				break
			}
		}
		if targetSection == nil {
			continue
		}
		// validate user
		if rawUser, exist := targetSection.Fields["valid users"]; exist {
			userNames := strings.Split(rawUser, ",")
			validaUsers := make([]ShareFolderUsers, 0)
			for _, userName := range userNames {
				user := service.DefaultUserManager.GetUserByName(userName)
				if user == nil {
					continue
				}
				userTemplate := ShareFolderUsers{
					Uid:  user.Uid,
					Name: user.Username,
				}
				validaUsers = append(validaUsers, userTemplate)
			}
			template.ValidateUsers = validaUsers
		}
		if rawUser, exist := targetSection.Fields["write list"]; exist {
			userNames := strings.Split(rawUser, ",")
			writeUsers := make([]ShareFolderUsers, 0)
			for _, userName := range userNames {
				user := service.DefaultUserManager.GetUserByName(userName)
				if user == nil {
					continue
				}
				userTemplate := ShareFolderUsers{
					Uid:  user.Uid,
					Name: user.Username,
				}
				writeUsers = append(writeUsers, userTemplate)
			}
			template.WriteableUsers = writeUsers
		}
		if public, exist := targetSection.Fields["public"]; exist {
			template.Public = public
		} else {
			template.Public = "Not set"
		}
		if readonly, exist := targetSection.Fields["read only"]; exist {
			template.Readonly = readonly
		} else {
			template.Readonly = "Not set"
		}
		if writable, exist := targetSection.Fields["writable"]; exist {
			template.Writable = writable
		} else {
			template.Writable = "Not set"
		}
		shareFolders = append(shareFolders, template)
	}
	context.JSON(haruka.JSON{
		"folders": shareFolders,
	})
}

var updateShareFolder haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody service.UpdateShareFolderOption
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.UpdateSMBConfig(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var removeShareHandler haruka.RequestHandler = func(context *haruka.Context) {
	id, err := context.GetQueryInt("id")
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	err = service.RemoveShare(uint(id))
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
