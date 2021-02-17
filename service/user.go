package service

import (
	"bufio"
	"fmt"
	. "github.com/ahmetb/go-linq/v3"
	"os"
	"os/exec"
	"strings"
	"youplus/config"
	"youplus/utils"
	"youplus/yousmb"
)

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

func NewUser(username string, password string) error {
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
	err = yousmb.AddUser(username, password)
	if err != nil {
		return err
	}
	config.Config.Users = append(config.Config.Users, username)
	err = config.Config.UpdateConfig()
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
