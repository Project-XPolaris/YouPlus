package utils

import (
	"fmt"
	"os/exec"
)

func WipeDiskFS(device string) error {
	cmd := exec.Command("wipefs", "-a", device)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func CreateAppendDiskPartition(device string, partType int, size string) error {
	script := fmt.Sprintf("echo ,%s,%d, | sfdisk %s", size, partType, device)
	cmd := exec.Command("/bin/sh", "-c", script)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
func DeletePartition(device string, partitionId int) error {
	cmd := exec.Command("sfdisk", device, "--delete", fmt.Sprintf("%d", partitionId))
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
