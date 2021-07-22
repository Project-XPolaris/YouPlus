package application

import (
	"errors"
	"github.com/allentom/haruka"
	"github.com/projectxpolaris/youplus/database"
	"github.com/projectxpolaris/youplus/service"
	"net/http"
)

var (
	ShareFolderExist = errors.New("share folder exist")
)
var createShareHandler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody service.NewShareFolderOption
	err := context.ParseJson(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	count, err := database.CountShareFolderByName(requestBody.Name)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	if count != 0 {
		AbortErrorWithStatus(ShareFolderExist, context, http.StatusBadRequest)
		return
	}
	err = service.CreateNewShareFolder(&requestBody)
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	err = service.DefaultAddressConverterManager.Load()
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
	shareFolders := make([]ShareFolderTemplate, 0)
	for _, shareFolderConfig := range folderList {
		template := ShareFolderTemplate{
			Id:       shareFolderConfig.ID,
			Name:     shareFolderConfig.Name,
			Enable:   shareFolderConfig.Enable,
			Public:   shareFolderConfig.Public,
			Readonly: shareFolderConfig.Readonly,
		}
		if shareFolderConfig.Public {
			template.Guest = service.UserShareFolder{
				Name:   shareFolderConfig.Name,
				Access: true,
				Read:   true,
				Write:  !shareFolderConfig.Readonly,
			}
		} else {
			template.Guest = service.UserShareFolder{
				Name:   shareFolderConfig.Name,
				Access: false,
			}
		}
		template.Other = service.UserShareFolder{
			Name:   shareFolderConfig.Name,
			Access: true,
			Read:   true,
			Write:  !shareFolderConfig.Readonly,
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
		validUsers := make([]ShareFolderUsers, 0)
		for _, user := range shareFolderConfig.ValidUsers {
			systemUser := service.DefaultUserManager.GetUserByName(user.Username)
			if systemUser == nil {
				continue
			}
			validUsers = append(validUsers, ShareFolderUsers{
				Uid:  systemUser.Uid,
				Name: systemUser.Username,
			})
		}
		template.ValidUsers = validUsers
		invalidUsers := make([]ShareFolderUsers, 0)
		for _, user := range shareFolderConfig.InvalidUsers {
			systemUser := service.DefaultUserManager.GetUserByName(user.Username)
			if systemUser == nil {
				continue
			}
			invalidUsers = append(invalidUsers, ShareFolderUsers{
				Uid:  systemUser.Uid,
				Name: systemUser.Username,
			})
		}
		template.InvalidUsers = invalidUsers
		readUsers := make([]ShareFolderUsers, 0)
		for _, user := range shareFolderConfig.ReadUsers {
			systemUser := service.DefaultUserManager.GetUserByName(user.Username)
			if systemUser == nil {
				continue
			}
			readUsers = append(readUsers, ShareFolderUsers{
				Uid:  systemUser.Uid,
				Name: systemUser.Username,
			})
		}
		template.ReadUsers = readUsers
		writeUsers := make([]ShareFolderUsers, 0)
		for _, user := range shareFolderConfig.WriteUsers {
			systemUser := service.DefaultUserManager.GetUserByName(user.Username)
			if systemUser == nil {
				continue
			}
			writeUsers = append(writeUsers, ShareFolderUsers{
				Uid:  systemUser.Uid,
				Name: systemUser.Username,
			})
		}
		template.WriteUsers = writeUsers
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
	err = service.DefaultAddressConverterManager.Load()
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
	err = service.DefaultAddressConverterManager.Load()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}
