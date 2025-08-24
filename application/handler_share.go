package application

import (
	"errors"
	"net/http"
	"strings"

	"github.com/allentom/haruka"
	librpc "github.com/project-xpolaris/youplustoolkit/yousmb/rpc"
	"github.com/projectxpolaris/youplus/database"
	"github.com/projectxpolaris/youplus/service"
	"github.com/projectxpolaris/youplus/yousmb"
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
	err = service.InitFileSystem()
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
		if len(shareFolderConfig.PathStorageId) != 0 {
			sid = shareFolderConfig.PathStorageId
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
		template.ValidGroups = SerializeGroups(shareFolderConfig.ValidGroups)
		template.InvalidGroups = SerializeGroups(shareFolderConfig.InvalidGroups)
		template.ReadGroups = SerializeGroups(shareFolderConfig.ReadGroups)
		template.WriteGroups = SerializeGroups(shareFolderConfig.WriteGroups)
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
	err = service.InitFileSystem()
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
	err = service.InitFileSystem()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
	})
}

var getSMBStatusHandler haruka.RequestHandler = func(context *haruka.Context) {
	status, err := service.GetSMBStatus()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	processList := make([]SMBProcessStatusTemplate, 0)
	if status.Process != nil {
		for _, process := range status.Process {
			template := SMBProcessStatusTemplate{
				PID:      *process.PID,
				Username: *process.Username,
				Group:    *process.Group,
				Machine:  *process.Machine,
			}
			processList = append(processList, template)
		}
	}
	sharesList := make([]SMBSharesStatusTemplate, 0)
	if status.Shares != nil {
		for _, share := range status.Shares {
			template := SMBSharesStatusTemplate{
				Service:   *share.Service,
				PID:       *share.PID,
				Machine:   *share.Machine,
				ConnectAt: *share.ConnectAt,
			}
			sharesList = append(sharesList, template)
		}
	}
	context.JSON(haruka.JSON{
		"success": true,
		"process": processList,
		"shares":  sharesList,
	})
}

var syncStorageAndShareHandler haruka.RequestHandler = func(context *haruka.Context) {
	// 1) sync ZFS mounts to storage
	createdStorages, updatedStorages, err := service.SyncZFSMountsToStorage()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	// reload storage pool in-memory
	service.DefaultStoragePool = service.StoragePool{Storages: []service.Storage{}}
	_ = service.DefaultStoragePool.LoadStorage()

	// 2) sync SMB shares to DB share folders
	smbRes, err := service.SyncSmbSharesToDB()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	// 3) rebuild filesystem view
	_ = service.InitFileSystem()
	context.JSON(haruka.JSON{
		"success":         true,
		"createdStorages": createdStorages,
		"updatedStorages": updatedStorages,
		"createdShares":   smbRes.CreatedShares,
		"updatedShares":   smbRes.UpdatedShares,
		"errors":          smbRes.Errors,
	})
}

// list SMB sections with share-folder flag
var listSMBSectionsHandler haruka.RequestHandler = func(context *haruka.Context) {
	// load all share folder names for quick lookup
	var shareFolders []database.ShareFolder
	err := database.Instance.Find(&shareFolders).Error
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	nameToShare := map[string]database.ShareFolder{}
	for _, f := range shareFolders {
		nameToShare[f.Name] = f
	}

	var cfg *librpc.ConfigReply
	err = yousmb.ExecWithRPCClient(func(client librpc.YouSMBServiceClient) error {
		var e error
		cfg, e = client.GetConfig(yousmb.GetRPCTimeoutContext(), &librpc.Empty{})
		return e
	})
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	type SectionTemplate struct {
		Name          string            `json:"name"`
		Fields        map[string]string `json:"fields"`
		IsShareFolder bool              `json:"isShareFolder"`
		ShareFolderId uint              `json:"shareFolderId,omitempty"`
	}
	list := make([]SectionTemplate, 0)
	if cfg != nil && cfg.Sections != nil {
		for _, s := range cfg.Sections {
			if s == nil || s.Name == nil || s.Fields == nil {
				continue
			}
			name := strings.TrimSpace(*s.Name)
			tmpl := SectionTemplate{
				Name:          name,
				Fields:        s.Fields,
				IsShareFolder: false,
			}
			if sf, ok := nameToShare[name]; ok {
				tmpl.IsShareFolder = true
				tmpl.ShareFolderId = sf.ID
			}
			list = append(list, tmpl)
		}
	}
	context.JSON(haruka.JSON{
		"sections": list,
	})
}

// get raw smb config text (reconstructed)
var getSMBRawConfigHandler haruka.RequestHandler = func(context *haruka.Context) {
	var cfg *librpc.ConfigReply
	err := yousmb.ExecWithRPCClient(func(client librpc.YouSMBServiceClient) error {
		var e error
		cfg, e = client.GetConfig(yousmb.GetRPCTimeoutContext(), &librpc.Empty{})
		return e
	})
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	var sb strings.Builder
	if cfg != nil && cfg.Sections != nil {
		for _, s := range cfg.Sections {
			if s == nil || s.Name == nil || s.Fields == nil {
				continue
			}
			sb.WriteString("[")
			sb.WriteString(*s.Name)
			sb.WriteString("]\n")
			for k, v := range s.Fields {
				sb.WriteString("    ")
				sb.WriteString(k)
				sb.WriteString(" = ")
				sb.WriteString(v)
				sb.WriteString("\n")
			}
			sb.WriteString("\n\n")
		}
	}
	context.JSON(haruka.JSON{
		"raw": sb.String(),
	})
}

// import SMB sections into DB with strict path matching
var importShareFromSMBHandler haruka.RequestHandler = func(context *haruka.Context) {
	res, err := service.ImportSmbSharesToDBStrict()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success":       true,
		"createdShares": res.CreatedShares,
		"updatedShares": res.UpdatedShares,
		"errors":        res.Errors,
	})
}

// restart SMB service by issuing a no-op update to trigger SaveFileAndRestart
var restartSMBHandler haruka.RequestHandler = func(context *haruka.Context) {
	var cfg *librpc.ConfigReply
	err := yousmb.ExecWithRPCClient(func(client librpc.YouSMBServiceClient) error {
		var e error
		cfg, e = client.GetConfig(yousmb.GetRPCTimeoutContext(), &librpc.Empty{})
		return e
	})
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	if cfg == nil || cfg.Sections == nil || len(cfg.Sections) == 0 {
		AbortErrorWithStatus(errors.New("no smb sections to touch"), context, http.StatusBadRequest)
		return
	}
	// pick first section and send an update with same fields
	sec := cfg.Sections[0]
	if sec == nil || sec.Name == nil {
		AbortErrorWithStatus(errors.New("invalid smb section"), context, http.StatusInternalServerError)
		return
	}
	err = yousmb.ExecWithRPCClient(func(client librpc.YouSMBServiceClient) error {
		_, e := client.UpdateFolderConfig(yousmb.GetRPCTimeoutContext(), &librpc.AddConfigMessage{Name: sec.Name, Properties: sec.Fields})
		return e
	})
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{"success": true})
}

var getSMBInfoHandler haruka.RequestHandler = func(context *haruka.Context) {
	reply, err := service.GetSMBInfo()
	if err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{
		"success": true,
		"name":    reply.GetName(),
		"status":  reply.GetStatus(),
	})
}

type UpdateSMBRawBody struct {
	Raw string `json:"raw"`
}

var updateSMBRawHandler haruka.RequestHandler = func(context *haruka.Context) {
	var body UpdateSMBRawBody
	if err := context.ParseJson(&body); err != nil {
		AbortErrorWithStatus(err, context, http.StatusBadRequest)
		return
	}
	if err := service.ApplySMBRawConfig(body.Raw); err != nil {
		AbortErrorWithStatus(err, context, http.StatusInternalServerError)
		return
	}
	context.JSON(haruka.JSON{"success": true})
}
