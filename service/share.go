package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/project-xpolaris/youplustoolkit/yousmb/rpc"
	"github.com/projectxpolaris/youplus/database"
	"github.com/projectxpolaris/youplus/utils"
	"github.com/projectxpolaris/youplus/yousmb"
	"gorm.io/gorm"
)

var (
	PartNotFoundError = errors.New("target part name not found")
	PartNotMountError = errors.New("target part not mounted")
)

type NewShareFolderOption struct {
	StorageId string `json:"storageId,omitempty"`
	Name      string `json:"name,omitempty"`
	Public    bool   `json:"public,omitempty"`
}

func CreateNewShareFolder(option *NewShareFolderOption) error {
	// get storage
	storage := DefaultStoragePool.GetStorageById(option.StorageId)
	if storage == nil {
		return StorageNotFoundError
	}
	// create share directory
	shareFolderPath := filepath.Join(storage.GetRootPath(), option.Name)
	if !strings.HasPrefix(shareFolderPath, "/") {
		shareFolderPath = "/" + shareFolderPath
	}
	err := os.MkdirAll(shareFolderPath, os.ModePerm)
	if err != nil {
		return err
	}
	// create share
	shareFolder := database.ShareFolder{
		Name:   option.Name,
		Public: option.Public,
		Path:   shareFolderPath,
		Enable: false,
	}
	switch storage.(type) {
	case *ZFSPoolStorage:
		shareFolder.ZFSStorageId = storage.GetId()
	case *DiskPartStorage:
		shareFolder.PartStorageId = storage.GetId()
	case *PathStorage:
		shareFolder.PathStorageId = storage.GetId()
	}
	err = SyncShareFolderOptionToSMB(&shareFolder)
	if err != nil {
		return err
	}
	err = database.Instance.Save(&shareFolder).Error
	if err != nil {
		return err
	}
	return nil
}
func getSMBUserAndUserGroupList(users []*database.User, groups []*database.UserGroup) []string {
	list := make([]string, 0)
	if users == nil && groups == nil {
		return list
	}
	if users != nil {
		for _, user := range users {
			list = append(list, user.Username)
		}
	}

	if groups != nil {
		for _, group := range groups {
			systemGroup := DefaultUserManager.GetGroupById(group.Gid)
			if systemGroup == nil {
				continue
			}
			list = append(list, fmt.Sprintf("@%s", systemGroup.Name))
		}
	}

	return list
}
func SyncShareFolderOptionToSMB(folder *database.ShareFolder) error {
	properties := map[string]string{
		"path":           folder.Path,
		"create mask":    "0775",
		"directory mask": "0775",
		"read only":      utils.GetSmbBoolText(folder.Readonly),
		"available":      utils.GetSmbBoolText(folder.Enable),
		"browseable":     utils.GetSmbBoolText(folder.Enable),
		"public":         utils.GetSmbBoolText(folder.Public),
		"force user":     "root",
		"force group":    "root",
	}
	validUsers := getSMBUserAndUserGroupList(folder.ValidUsers, folder.ValidGroups)
	if len(validUsers) > 0 {
		properties["valid users"] = strings.Join(validUsers, ",")
	}
	invalidUsers := getSMBUserAndUserGroupList(folder.InvalidUsers, folder.InvalidGroups)
	if len(invalidUsers) > 0 {
		properties["invalid users"] = strings.Join(invalidUsers, ",")
	}
	readUsers := getSMBUserAndUserGroupList(folder.ReadUsers, folder.ReadGroups)
	if len(readUsers) > 0 {
		properties["read list"] = strings.Join(readUsers, ",")
	}
	writeUsers := getSMBUserAndUserGroupList(folder.WriteUsers, folder.WriteGroups)
	if len(writeUsers) > 0 {
		properties["write list"] = strings.Join(writeUsers, ",")
	}
	if folder.Public {
		properties["public"] = "yes"
	}
	err := yousmb.ExecWithRPCClient(func(client rpc.YouSMBServiceClient) error {
		response, err := client.GetConfig(yousmb.GetRPCTimeoutContext(), &rpc.Empty{})
		if err != nil {
			return err
		}
		if response.Sections == nil {
			return errors.New("cannot get smb config")
		}
		for _, section := range response.Sections {
			// update
			if *section.Name == folder.Name {
				_, err = client.UpdateFolderConfig(yousmb.GetRPCTimeoutContext(), &rpc.AddConfigMessage{Name: &folder.Name, Properties: properties})
				return err
			}
		}
		//create new
		createReply, err := client.AddFolderConfig(yousmb.GetRPCTimeoutContext(), &rpc.AddConfigMessage{Name: &folder.Name, Properties: properties})
		if !createReply.GetSuccess() {
			return errors.New(createReply.GetReason())
		}
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
func GetShareFolders() ([]*database.ShareFolder, error) {
	var folders []*database.ShareFolder
	err := database.Instance.
		Preload("ValidUsers").
		Preload("InvalidUsers").
		Preload("ReadUsers").
		Preload("WriteUsers").
		Preload("ValidGroups").
		Preload("InvalidGroups").
		Preload("ReadGroups").
		Preload("WriteGroups").
		Find(&folders).Error
	if err != nil {
		return nil, err
	}
	return folders, nil
}
func GetShareFolderCount() (int64, error) {
	var count int64
	err := database.Instance.Model(&database.ShareFolder{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func putFolderUserList(folder *database.ShareFolder, usernameList []string, rel string) error {
	var users []*database.User
	err := database.Instance.Where("username in ?", usernameList).Find(&users).Error
	if err != nil {
		return err
	}
	err = database.Instance.Model(folder).Association(rel).Clear()
	if err != nil {
		return err
	}
	err = database.Instance.Model(folder).Association(rel).Append(users)
	if err != nil {
		return err
	}
	return nil
}
func putFolderGroupList(folder *database.ShareFolder, groupNameList []string, rel string) error {
	var groups []*database.UserGroup
	groupIds := make([]string, 0)
	for _, name := range groupNameList {
		group := DefaultUserManager.GetUserGroupByName(name)
		if group == nil {
			continue
		}
		groupIds = append(groupIds, group.Gid)
	}
	err := database.Instance.Where("gid in ?", groupIds).Find(&groups).Error
	if err != nil {
		return err
	}
	err = database.Instance.Model(folder).Association(rel).Clear()
	if err != nil {
		return err
	}
	err = database.Instance.Model(folder).Association(rel).Append(groups)
	if err != nil {
		return err
	}
	return nil
}

type UpdateShareFolderOption struct {
	Name          string   `json:"name"`
	ReadUsers     []string `json:"readUsers"`
	WriteUsers    []string `json:"writeUsers"`
	ValidUsers    []string `json:"validUsers"`
	InvalidUsers  []string `json:"invalidUsers"`
	ReadGroups    []string `json:"readGroups"`
	WriteGroups   []string `json:"writeGroups"`
	ValidGroups   []string `json:"validGroups"`
	InvalidGroups []string `json:"invalidGroups"`
	Public        *bool    `json:"public"`
	Readonly      *bool    `json:"readonly"`
	Enable        *bool    `json:"enable"`
	StorageId     *string  `json:"storageId"`
	NewName       *string  `json:"newName"`
}

func UpdateSMBConfig(option *UpdateShareFolderOption) error {
	var folder database.ShareFolder
	err := database.Instance.Where("name = ?", option.Name).
		Preload("ValidUsers").
		Preload("InvalidUsers").
		Preload("ReadUsers").
		Preload("WriteUsers").
		Preload("ValidGroups").
		Preload("InvalidGroups").
		Preload("ReadGroups").
		Preload("WriteGroups").
		First(&folder).
		Error
	if err != nil {
		return err
	}
	// handle storage change
	if option.StorageId != nil && len(*option.StorageId) > 0 {
		storage := DefaultStoragePool.GetStorageById(*option.StorageId)
		if storage == nil {
			return StorageNotFoundError
		}
		newPath := filepath.Join(storage.GetRootPath(), folder.Name)
		if !strings.HasPrefix(newPath, "/") {
			newPath = "/" + newPath
		}
		if mkErr := os.MkdirAll(newPath, os.ModePerm); mkErr != nil {
			return mkErr
		}
		folder.Path = newPath
		folder.ZFSStorageId = ""
		folder.PartStorageId = ""
		folder.PathStorageId = ""
		switch storage.(type) {
		case *ZFSPoolStorage:
			folder.ZFSStorageId = storage.GetId()
		case *DiskPartStorage:
			folder.PartStorageId = storage.GetId()
		case *PathStorage:
			folder.PathStorageId = storage.GetId()
		}
	}

	// handle rename
	if option.NewName != nil && len(*option.NewName) > 0 && *option.NewName != folder.Name {
		oldName := folder.Name
		oldPath := folder.Path
		// check source exists
		if _, statErr := os.Stat(oldPath); statErr != nil {
			return statErr
		}
		baseDir := filepath.Dir(oldPath)
		destPath := filepath.Join(baseDir, *option.NewName)
		if !strings.HasPrefix(destPath, "/") {
			destPath = "/" + destPath
		}
		if _, statErr := os.Stat(destPath); statErr == nil {
			return errors.New("target directory already exists")
		}
		if rnErr := os.Rename(oldPath, destPath); rnErr != nil {
			return rnErr
		}
		folder.Name = *option.NewName
		folder.Path = destPath
		// sync new section first, then remove old
		if err := SyncShareFolderOptionToSMB(&folder); err != nil {
			return err
		}
		rmErr := yousmb.ExecWithRPCClient(func(client rpc.YouSMBServiceClient) error {
			reply, e := client.RemoveFolderConfig(yousmb.GetRPCTimeoutContext(), &rpc.RemoveConfigMessage{Name: &oldName})
			if e != nil {
				return e
			}
			if !reply.GetSuccess() {
				return errors.New(reply.GetReason())
			}
			return nil
		})
		if rmErr != nil {
			return rmErr
		}
	}

	// ACLs and flags
	if option.ValidUsers != nil {
		err = putFolderUserList(&folder, option.ValidUsers, "ValidUsers")
		if err != nil {
			return err
		}
	}
	if option.InvalidUsers != nil {
		err = putFolderUserList(&folder, option.InvalidUsers, "InvalidUsers")
		if err != nil {
			return err
		}
	}
	if option.ReadUsers != nil {
		err = putFolderUserList(&folder, option.ReadUsers, "ReadUsers")
		if err != nil {
			return err
		}
	}
	if option.WriteUsers != nil {
		err = putFolderUserList(&folder, option.WriteUsers, "WriteUsers")
		if err != nil {
			return err
		}
	}
	if option.ValidGroups != nil {
		err = putFolderGroupList(&folder, option.ValidGroups, "ValidGroups")
		if err != nil {
			return err
		}
	}
	if option.InvalidGroups != nil {
		err = putFolderGroupList(&folder, option.InvalidGroups, "InvalidGroups")
		if err != nil {
			return err
		}
	}
	if option.ReadGroups != nil {
		err = putFolderGroupList(&folder, option.ReadGroups, "ReadGroups")
		if err != nil {
			return err
		}
	}
	if option.WriteGroups != nil {
		err = putFolderGroupList(&folder, option.WriteGroups, "WriteGroups")
		if err != nil {
			return err
		}
	}
	if option.Public != nil {
		folder.Public = *option.Public
	}
	if option.Enable != nil {
		folder.Enable = *option.Enable
	}
	if option.Readonly != nil {
		folder.Readonly = *option.Readonly
	}

	// for non-rename updates, ensure SMB section updated
	if err := SyncShareFolderOptionToSMB(&folder); err != nil {
		return err
	}

	err = database.Instance.Save(&folder).Error
	if err != nil {
		return err
	}

	return err
}

func RemoveShare(id uint) error {
	shareFolder := database.ShareFolder{
		Model: gorm.Model{ID: id},
	}
	err := database.Instance.Find(&shareFolder).Error
	if err != nil {
		return err
	}
	err = yousmb.ExecWithRPCClient(func(client rpc.YouSMBServiceClient) error {
		reply, err := client.RemoveFolderConfig(yousmb.GetRPCTimeoutContext(), &rpc.RemoveConfigMessage{Name: &shareFolder.Name})
		if err != nil {
			return err
		}
		if !reply.GetSuccess() {
			return errors.New(reply.GetReason())
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = database.Instance.Model(&database.ShareFolder{}).
		Unscoped().
		Delete(&database.ShareFolder{Model: gorm.Model{ID: shareFolder.ID}}).
		Error
	if err != nil {
		return err
	}
	return nil
}

func GetSMBStatus() (*rpc.SMBStatusReply, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var reply *rpc.SMBStatusReply
	err := yousmb.ExecWithRPCClient(func(client rpc.YouSMBServiceClient) error {
		var err error
		reply, err = client.GetSMBStatus(timeout, &rpc.Empty{})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func GetSMBInfo() (*rpc.ServiceInfoReply, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var reply *rpc.ServiceInfoReply
	err := yousmb.ExecWithRPCClient(func(client rpc.YouSMBServiceClient) error {
		var err error
		reply, err = client.GetInfo(timeout, &rpc.Empty{})
		return err
	})
	if err != nil {
		return nil, err
	}
	return reply, nil
}
