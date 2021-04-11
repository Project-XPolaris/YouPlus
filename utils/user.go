package utils

import (
	"fmt"
	"github.com/amoghe/go-crypt"
	"os/exec"
)

func CheckPassword(salt string, password string, savedPassword string) (bool, error) {
	sha512, err := crypt.Crypt(password, salt)
	if err != nil {
		return false, err
	}
	return savedPassword == sha512, nil
}

func GeneratePassword(password string) (string, error) {
	args := []string{
		"passwd",
		"-6",
		password,
	}

	cmd := exec.Command("openssl", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), err
}

func ChangePassword(username string, password string) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("echo %s:%s | chpasswd", username, password))
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
