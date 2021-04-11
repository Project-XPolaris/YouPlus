package service

import (
	"bufio"
	"errors"
	"fmt"
	. "github.com/ahmetb/go-linq/v3"
	"github.com/projectxpolaris/youplus/config"
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

func (m *UserManager) CheckPassword(username string, password string) (validate bool) {
	user := m.GetShadowByName(username)
	if user == nil {
		return false
	}
	return user.CheckPassword(password)
}

func (m *UserManager) CreateGroup(name string) error {
	cmd := exec.Command("groupadd", "-f", name)
	err := cmd.Run()
	if err != nil {
		return err
	}
	err = m.LoadUser()
	if err != nil {
		return err
	}
	return err
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
		err = yousmb.AddUser(username, password)
		if err != nil {
			return err
		}
	}
	config.Config.Users = append(config.Config.Users, username)
	err = config.Config.UpdateConfig()
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
	From(users).Where(func(i interface{}) bool {
		for _, user := range config.Config.Users {
			if i.(*SystemUser).Username == user {
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
