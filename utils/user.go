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
	fmt.Println(sha512)
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