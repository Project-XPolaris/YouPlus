package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/project-xpolaris/youplustoolkit/yousmb/rpc"
	"github.com/projectxpolaris/youplus/database"
	"github.com/projectxpolaris/youplus/utils"
	"github.com/projectxpolaris/youplus/yousmb"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strings"
	"time"
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
		"browseable":     "yes",
		"public":         utils.GetSmbBoolText(folder.Public),
		"valid users":    strings.Join(getSMBUserAndUserGroupList(folder.ValidUsers, folder.ValidGroups), ","),
		"invalid users":  strings.Join(getSMBUserAndUserGroupList(folder.InvalidUsers, folder.InvalidGroups), ","),
		"read list":      strings.Join(getSMBUserAndUserGroupList(folder.ReadUsers, folder.ReadGroups), ","),
		"write list":     strings.Join(getSMBUserAndUserGroupList(folder.WriteUsers, folder.WriteGroups), ","),
	}
	if folder.Public {
		properties["public"] = "yes"
	}
	response, err := yousmb.DefaultYouSMBRPCClient.Client.GetConfig(yousmb.GetRPCTimeoutContext(), &rpc.Empty{})
	if err != nil {
		return err
	}
	if response.Sections == nil {
		return errors.New("cannot get smb config")
	}
	for _, section := range response.Sections {
		// update
		if *section.Name == folder.Name {
			_, err = yousmb.DefaultYouSMBRPCClient.Client.UpdateFolderConfig(yousmb.GetRPCTimeoutContext(), &rpc.AddConfigMessage{Name: &folder.Name, Properties: properties})
			return err
		}
	}
	//create new
	createReply, err := yousmb.DefaultYouSMBRPCClient.Client.AddFolderConfig(yousmb.GetRPCTimeoutContext(), &rpc.AddConfigMessage{Name: &folder.Name, Properties: properties})
	if !createReply.GetSuccess() {
		return errors.New(createReply.GetReason())
	}
	return err
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
	err = SyncShareFolderOptionToSMB(&folder)
	if err != nil {
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
	reply, err := yousmb.DefaultYouSMBRPCClient.Client.RemoveFolderConfig(yousmb.GetRPCTimeoutContext(), &rpc.RemoveConfigMessage{Name: &shareFolder.Name})
	if err != nil {
		return err
	}
	if !reply.GetSuccess() {
		return errors.New(reply.GetReason())
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
	timeout, _ := context.WithTimeout(context.Background(), 10*time.Second)
	reply, err := yousmb.DefaultYouSMBRPCClient.Client.GetSMBStatus(timeout, &rpc.Empty{})
	if err != nil {
		return nil, err
	}
	return reply, nil
}
