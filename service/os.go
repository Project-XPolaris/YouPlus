package service

import "os/exec"

func Shutdown() error {
	cmd := exec.Command("shutdown", "-h", "now")
	err := cmd.Start()
	return err
}

func Reboot() error {
	cmd := exec.Command("reboot")
	err := cmd.Start()
	return err
}
