package service

import (
	"strings"

	librpc "github.com/project-xpolaris/youplustoolkit/yousmb/rpc"
	"github.com/projectxpolaris/youplus/database"
	"github.com/projectxpolaris/youplus/yousmb"
)

type SyncResult struct {
	CreatedStorages int      `json:"createdStorages"`
	UpdatedStorages int      `json:"updatedStorages"`
	CreatedShares   int      `json:"createdShares"`
	UpdatedShares   int      `json:"updatedShares"`
	Errors          []string `json:"errors"`
}

func parseSMBBool(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	return value == "yes" || value == "true" || value == "1"
}

func splitUserAndGroups(list string) (users []string, groups []string) {
	items := strings.Split(list, ",")
	for _, raw := range items {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		if strings.HasPrefix(name, "@") {
			groups = append(groups, strings.TrimPrefix(name, "@"))
			continue
		}
		users = append(users, name)
	}
	return
}

func ensureZFSStorageByPool(poolName string) (string, error) {
	// try find by Name first
	var exist database.ZFSStorage
	result := database.Instance.Where("name = ?", poolName).First(&exist)
	if result.Error == nil && exist.ID != "" {
		return exist.ID, nil
	}
	// try find by MountPoint (legacy stored as dataset path/pool name)
	result = database.Instance.Where("mount_point = ?", poolName).First(&exist)
	if result.Error == nil && exist.ID != "" {
		return exist.ID, nil
	}
	// create new storage with pool root as dataset path
	storage, err := CreateZFSStorage(poolName)
	if err != nil {
		return "", err
	}
	return storage.GetId(), nil
}

// SyncSmbSharesToDB scans current SMB configuration and ensures corresponding
// ZFS storages (by pool) and ShareFolder records exist and are updated.
func SyncSmbSharesToDB() (*SyncResult, error) {
	res := &SyncResult{Errors: make([]string, 0)}
	var cfg *librpc.ConfigReply
	err := yousmb.ExecWithRPCClient(func(client librpc.YouSMBServiceClient) error {
		var e error
		cfg, e = client.GetConfig(yousmb.GetRPCTimeoutContext(), &librpc.Empty{})
		return e
	})
	if err != nil {
		return nil, err
	}
	if cfg == nil || cfg.Sections == nil {
		return res, nil
	}
	for _, section := range cfg.Sections {
		if section == nil || section.Name == nil || section.Fields == nil {
			continue
		}
		name := strings.TrimSpace(*section.Name)
		if name == "" || strings.ToLower(name) == "global" {
			continue
		}
		path := strings.TrimSpace(section.Fields["path"])
		if path == "" || !strings.HasPrefix(path, "/") {
			// skip non-path shares
			continue
		}
		// derive pool name from "/pool/..."
		parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			continue
		}
		poolName := parts[0]
		storageId, e := ensureZFSStorageByPool(poolName)
		if e != nil {
			res.Errors = append(res.Errors, e.Error())
			continue
		}
		// upsert share folder by name
		var share database.ShareFolder
		database.Instance.Where("name = ?", name).First(&share)
		isCreate := share.ID == 0
		share.Name = name
		share.Path = path
		share.ZFSStorageId = storageId
		share.PartStorageId = ""
		share.PathStorageId = ""
		// flags
		if v, ok := section.Fields["available"]; ok {
			share.Enable = parseSMBBool(v)
		}
		if v, ok := section.Fields["browseable"]; ok {
			// if available not set, use browseable as a hint
			if !share.Enable {
				share.Enable = parseSMBBool(v)
			}
		}
		if v, ok := section.Fields["public"]; ok {
			share.Public = parseSMBBool(v)
		}
		if v, ok := section.Fields["read only"]; ok {
			share.Readonly = parseSMBBool(v)
		}

		// persist basic fields first to ensure ID
		if isCreate {
			if e = database.Instance.Save(&share).Error; e != nil {
				res.Errors = append(res.Errors, e.Error())
				continue
			}
			res.CreatedShares++
		} else {
			if e = database.Instance.Save(&share).Error; e != nil {
				res.Errors = append(res.Errors, e.Error())
				continue
			}
			res.UpdatedShares++
		}

		// sync ACL lists
		if v, ok := section.Fields["valid users"]; ok {
			users, groups := splitUserAndGroups(v)
			_ = putFolderUserList(&share, users, "ValidUsers")
			_ = putFolderGroupList(&share, groups, "ValidGroups")
		}
		if v, ok := section.Fields["invalid users"]; ok {
			users, groups := splitUserAndGroups(v)
			_ = putFolderUserList(&share, users, "InvalidUsers")
			_ = putFolderGroupList(&share, groups, "InvalidGroups")
		}
		if v, ok := section.Fields["read list"]; ok {
			users, groups := splitUserAndGroups(v)
			_ = putFolderUserList(&share, users, "ReadUsers")
			_ = putFolderGroupList(&share, groups, "ReadGroups")
		}
		if v, ok := section.Fields["write list"]; ok {
			users, groups := splitUserAndGroups(v)
			_ = putFolderUserList(&share, users, "WriteUsers")
			_ = putFolderGroupList(&share, groups, "WriteGroups")
		}
	}
	return res, nil
}

// SyncZFSMountsToStorage scans ZFS pools and ensures a corresponding ZFSStorage exists per pool.
func SyncZFSMountsToStorage() (int, int, error) {
	created := 0
	updated := 0
	// open all pools via libzfs in existing manager helpers
	pools, err := DefaultZFSManager.GetPoolList(&ZFSPoolListFilter{})
	if err != nil {
		return 0, 0, err
	}
	for _, pool := range pools {
		name, e := pool.Name()
		if e != nil {
			return created, updated, e
		}
		var exist database.ZFSStorage
		find := database.Instance.Where("mount_point = ?", name).First(&exist)
		if find.Error == nil && exist.ID != "" {
			// already exists
			continue
		}
		// not exist -> create
		_, e = CreateZFSStorage(name)
		if e != nil {
			return created, updated, e
		}
		created++
	}
	return created, updated, nil
}
