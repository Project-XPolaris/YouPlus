package service

import (
	"syscall"
)

func Shutdown() error {
	err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
	return err
}

func Reboot() error {
	err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	return err
}
