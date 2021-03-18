package service

import (
	"bufio"
	"fmt"
	"github.com/d-tux/go-fstab"
	"github.com/projectxpolaris/youplus/config"
	"os"
	"os/exec"
)

var DefaultFstab Fstab = Fstab{}

type Fstab struct {
	Mounts fstab.Mounts
}

func LoadFstab() error {
	mounts, err := fstab.ParseFile(config.Config.Fstab)
	if err != nil {
		return err
	}
	DefaultFstab.Mounts = mounts
	return nil
}

type AddMountOption struct {
	Spec    string            `json:"spec"`
	File    string            `json:"file"`
	VfsType string            `json:"vfs_type"`
	MntOps  map[string]string `json:"mnt_ops"`
	Freq    int               `json:"freq"`
	PassNo  int               `json:"pass_no"`
}

func (f *Fstab) AddMount(option *AddMountOption) {
	for _, mount := range f.Mounts {
		if mount.File == option.File {
			return
		}
	}
	f.Mounts = append(f.Mounts, &fstab.Mount{
		Spec:    option.Spec,
		File:    option.File,
		VfsType: option.VfsType,
		MntOps:  option.MntOps,
		Freq:    option.Freq,
		PassNo:  option.PassNo,
	})
}
func (f *Fstab) RemoveMount(file string) error {
	index := -1
	for mindex, mount := range f.Mounts {
		if mount.File == file {
			index = mindex
		}
	}
	if index != -1 {
		f.Mounts[index] = f.Mounts[len(f.Mounts)-1]
		f.Mounts = f.Mounts[:len(f.Mounts)-1]
		err := UmountFS(file, "-l")
		if err != nil {
			return err
		}
	}
	return nil
}
func (f *Fstab) Save() error {
	file, err := os.Create(config.Config.Fstab)
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	for _, mount := range f.Mounts {
		_, err = fmt.Fprintln(w, fmt.Sprintf(mount.String()))
		if err != nil {
			return err
		}
	}
	return w.Flush()
}

func (f *Fstab) Reload() error {
	out, err := exec.Command("mount", "-a").Output()
	if err != nil {
		return err
	}
	fmt.Println("fstab is reloaded")
	output := string(out[:])
	fmt.Println(output)
	return nil
}

func UmountFS(dir string, extra ...string) error {
	params := make([]string, 0)
	params = append(params, dir)
	params = append(params, extra...)
	out, err := exec.Command("umount", params...).Output()
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("umount %s", dir))
	output := string(out[:])
	fmt.Println(output)
	return nil
}
