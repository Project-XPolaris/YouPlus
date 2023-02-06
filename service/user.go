package service

import (
	"bufio"
	"errors"
	"fmt"
	. "github.com/ahmetb/go-linq/v3"
	"github.com/project-xpolaris/youplustoolkit/yousmb/rpc"
	"github.com/projectxpolaris/youplus/database"
	"github.com/projectxpolaris/youplus/utils"
	"github.com/projectxpolaris/youplus/yousmb"
	"os"
	"os/exec"
	"strings"
)

var (
	DefaultUserManager = UserManager{}
	UserNotFoundError  = errors.New("target user not found")
)

type UserManager struct {
	Users   []*SystemUser
	Groups  []*SystemUserGroup
	Shadows []*Shadow
}

func (m *UserManager) LoadUser() (err error) {
	m.Users, err = GetSystemUserList()
	if err != nil {
		return err
	}
	m.Groups, err = GetSystemUserGroupList()
	if err != nil {
		return err
	}
	m.Shadows, err = GetUserShadowList()
	if err != nil {
		return err
	}
	return nil
}
func (m *UserManager) GetUserByName(username string) *SystemUser {
	for _, systemUser := range m.Users {
		if systemUser.Username == username {
			return systemUser
		}
	}
	return nil
}
func (m *UserManager) GetUserGroupByName(name string) *SystemUserGroup {
	for _, group := range m.Groups {
		if group.Name == name {
			return group
		}
	}
	return nil
}
func (m *UserManager) GetShadowByName(username string) *Shadow {
	for _, shadow := range m.Shadows {
		if shadow.Username == username {
			return shadow
		}
	}
	return nil
}

func (m *UserManager) GetGroupByName(name string) *SystemUserGroup {
	for _, group := range m.Groups {
		if group.Name == name {
			return group
		}
	}
	return nil
}
func (m *UserManager) GetGroupById(gid string) *SystemUserGroup {
	for _, group := range m.Groups {
		if group.Gid == gid {
			return group
		}
	}
	return nil
}
func (m *UserManager) GetGroups() (groups []*SystemUserGroup, err error) {
	saveGroups := make([]database.UserGroup, 0)
	err = database.Instance.Find(&saveGroups).Error
	if err != nil {
		return nil, err
	}
	From(m.Groups).Where(func(sysGroup interface{}) bool {
		for _, sa := range saveGroups {
			if sa.Gid == sysGroup.(*SystemUserGroup).Gid {
				return true
			}
		}
		return false
	}).ToSlice(&groups)
	return
}
func (m *UserManager) CheckPassword(username string, password string) (validate bool) {
	user := m.GetShadowByName(username)
	if user == nil {
		return false
	}
	return user.CheckPassword(password)
}

func (m *UserManager) CreateGroup(name string) (*SystemUserGroup, error) {
	cmd := exec.Command("groupadd", "-f", name)
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	err = m.LoadUser()
	if err != nil {
		return nil, err
	}
	createdGroup := From(m.Groups).FirstWith(func(group interface{}) bool {
		return group.(*SystemUserGroup).Name == name
	}).(*SystemUserGroup)
	err = database.Instance.Save(&database.UserGroup{
		Gid: createdGroup.Gid,
	}).Error
	if err != nil {
		return nil, err
	}
	return createdGroup, err
}
func (m *UserManager) RemoveUserGroup(name string) error {
	group := m.GetGroupByName(name)
	if group == nil {
		return nil
	}
	cmd := exec.Command("groupdel", name)
	err := cmd.Run()
	if err != nil {
		return err
	}
	err = database.Instance.Model(&database.UserGroup{}).Unscoped().Where("gid = ?", group.Gid).Delete(&database.UserGroup{Gid: group.Gid}).Error
	if err != nil {
		return err
	}
	return nil
}
func (m *UserManager) ChangeUserPassword(username string, password string) error {
	user := m.GetUserByName(username)
	if user == nil {
		return UserNotFoundError
	}
	err := user.ChangePassword(password)
	if err != nil {
		return err
	}
	err = m.LoadUser()
	if err != nil {
		return err
	}
	return nil
}

type SystemUser struct {
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	Uid           string `json:"uid,omitempty"`
	Gid           string `json:"gid,omitempty"`
	Comment       string `json:"comment,omitempty"`
	HomeDirectory string `json:"home_directory,omitempty"`
	Shell         string `json:"shell,omitempty"`
}

func GetSystemUserList() ([]*SystemUser, error) {
	usersFile, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer usersFile.Close()
	result := make([]*SystemUser, 0)
	scanner := bufio.NewScanner(usersFile)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		result = append(result, &SystemUser{
			Username:      parts[0],
			Password:      parts[1],
			Uid:           parts[2],
			Gid:           parts[3],
			Comment:       parts[4],
			HomeDirectory: parts[5],
			Shell:         parts[6],
		})
	}
	return result, nil
}

func (u *SystemUser) ChangePassword(NewPassword string) error {
	err := utils.ChangePassword(u.Username, NewPassword)
	if err != nil {
		return err
	}
	return nil
}

type Shadow struct {
	Username string
	Password string
}

func GetUserShadowList() ([]*Shadow, error) {
	shadowFile, err := os.Open("/etc/shadow")
	if err != nil {
		return nil, err
	}
	defer shadowFile.Close()
	result := make([]*Shadow, 0)
	scanner := bufio.NewScanner(shadowFile)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		result = append(result, &Shadow{
			Username: parts[0],
			Password: parts[1],
		})
	}
	return result, nil
}

func (s *Shadow) CheckPassword(password string) bool {
	terminateIndex := strings.LastIndex(s.Password, "$")
	salt := s.Password[:terminateIndex]
	ok, _ := utils.CheckPassword(salt, password, s.Password)
	return ok
}

func (m *UserManager) NewUser(username string, password string, only bool) error {
	cmd := exec.Command("useradd", username)
	err := cmd.Run()
	if err != nil {
		return err
	}
	cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("echo %s:%s | chpasswd", username, password))
	err = cmd.Run()
	if err != nil {
		return err
	}
	// add smb user
	if !only {
		err = yousmb.ExecWithRPCClient(func(client rpc.YouSMBServiceClient) error {
			reply, err := client.AddUser(yousmb.GetRPCTimeoutContext(), &rpc.AddUserMessage{
				Username: &username,
				Password: &password,
			})
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

	}
	err = database.Instance.Save(&database.User{
		Username: username,
	}).Error
	if err != nil {
		return err
	}
	err = m.LoadUser()
	if err != nil {
		return err
	}
	return nil
}

func (m *UserManager) RemoveUser(username string) error {
	cmd := exec.Command("userdel", username)
	err := cmd.Run()
	if err != nil {
		return err
	}
	// add smb user
	err = yousmb.ExecWithRPCClient(func(client rpc.YouSMBServiceClient) error {
		reply, err := client.RemoveUser(yousmb.GetRPCTimeoutContext(), &rpc.RemoveUserMessage{Username: &username})
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
	err = database.Instance.Unscoped().Where("username = ?", username).Delete(&database.User{Username: username}).Error
	if err != nil {
		return err
	}
	err = m.LoadUser()
	if err != nil {
		return err
	}
	return nil
}
func GetUserList() ([]string, error) {
	users, err := GetSystemUserList()
	if err != nil {
		return nil, err
	}
	result := make([]string, 0)
	var saveUsers []database.User
	err = database.Instance.Find(&saveUsers).Error
	if err != nil {
		return nil, err
	}
	From(users).Where(func(i interface{}) bool {
		for _, user := range saveUsers {
			if i.(*SystemUser).Username == user.Username {
				return true
			}
		}
		return false
	}).Select(func(i interface{}) interface{} {
		return i.(*SystemUser).Username
	}).ToSlice(&result)
	return result, nil
}

type SystemUserGroup struct {
	Name  string
	Gid   string
	Users []string
}

func GetSystemUserGroupList() ([]*SystemUserGroup, error) {
	userGroupFile, err := os.Open("/etc/group")
	if err != nil {
		return nil, err
	}
	defer userGroupFile.Close()
	result := make([]*SystemUserGroup, 0)
	scanner := bufio.NewScanner(userGroupFile)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		group := &SystemUserGroup{
			Name: parts[0],
			Gid:  parts[2],
		}
		group.Users = strings.Split(parts[3], ",")
		result = append(result, group)
	}
	return result, nil
}
func (g *SystemUserGroup) AddUser(user *SystemUser) error {
	cmd := exec.Command("usermod", "-a", "-G", g.Name, user.Username)
	err := cmd.Run()
	if err != nil {
		return err
	}
	err = DefaultUserManager.LoadUser()
	if err != nil {
		return err
	}
	return nil
}
func (g *SystemUserGroup) DelUser(user *SystemUser) error {
	cmd := exec.Command("gpasswd", "-d", user.Username, g.Name)
	err := cmd.Run()
	if err != nil {
		return err
	}
	err = DefaultUserManager.LoadUser()
	if err != nil {
		return err
	}
	return nil
}
func (g *SystemUserGroup) HasUser(username string) bool {
	for _, user := range g.Users {
		if user == username {
			return true
		}
	}
	return false
}

func GetUserCount() (count int64, err error) {
	err = database.Instance.Model(&database.User{}).Count(&count).Error
	return
}

type UserShareFolder struct {
	Name   string                `json:"name"`
	Folder *database.ShareFolder `json:"-"`
	Access bool                  `json:"access"`
	Read   bool                  `json:"read"`
	Write  bool                  `json:"write"`
}

func GetUserShareList(user *database.User) ([]*UserShareFolder, error) {
	folders, err := GetShareFolders()
	if err != nil {
		return nil, err
	}
	// remove invalidate
	From(folders).Where(func(i interface{}) bool {
		for _, invalidUser := range i.(*database.ShareFolder).InvalidUsers {
			if invalidUser.Username == user.Username {
				return false
			}
		}
		return true
	}).ToSlice(&folders)
	userFolders := make([]*UserShareFolder, 0)
	for _, folder := range folders {
		inReadUsers := From(folder.ReadUsers).AnyWith(func(i interface{}) bool {
			return i.(*database.User).Username == user.Username
		})
		inWriteUsers := From(folder.WriteUsers).AnyWith(func(i interface{}) bool {
			return i.(*database.User).Username == user.Username
		})
		userFolder := &UserShareFolder{
			Name:   folder.Name,
			Read:   true,
			Access: true,
			Folder: folder,
		}
		if folder.Readonly {
			if inWriteUsers {
				userFolder.Write = true
			} else {
				userFolder.Write = false
			}
		} else {
			if inReadUsers {
				userFolder.Write = false
				if inWriteUsers {
					userFolder.Write = true
				}
			} else {
				userFolder.Write = true
			}
		}
		userFolders = append(userFolders, userFolder)
	}
	return userFolders, nil
}

func GetUserByUsernameWithPassword(username string, password string) (*database.User, error) {
	var user database.User
	err := database.Instance.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	if !DefaultUserManager.CheckPassword(user.Username, password) {
		return nil, errors.New("password error")
	}
	return &user, nil
}
