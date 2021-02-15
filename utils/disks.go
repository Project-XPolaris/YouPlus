package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/shirou/gopsutil/disk"
	"io"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CmdRunner struct{}
type Disks map[string]map[string]string

func (c *CmdRunner) Run(cmd string, args []string) (io.Reader, error) {
	command := exec.Command(cmd, args...)
	resCh := make(chan []byte)
	errCh := make(chan error)
	go func() {
		out, err := command.CombinedOutput()
		if err != nil {
			errCh <- err
		}
		resCh <- out
	}()
	timer := time.After(2 * time.Second)
	select {
	case err := <-errCh:
		return nil, err
	case res := <-resCh:
		return bytes.NewReader(res), nil
	case <-timer:
		return nil, fmt.Errorf("time out (cmd:%v args:%v)", cmd, args)
	}
}

func (c *CmdRunner) Exec(cmd string, args []string) string {
	command := exec.Command(cmd, args...)
	outputBytes, _ := command.CombinedOutput()
	return string(outputBytes[:])
}

func ParserLsblk(r io.Reader) map[string]map[string]string {
	var lsblk = make(Disks)
	re := regexp.MustCompile("([A-Z]+)=(?:\"(.*?)\")")
	scan := bufio.NewScanner(r)
	for scan.Scan() {
		//pre := []string{"sd","hd"}
		var disk_name = ""
		disk := make(map[string]string)
		raw := scan.Text()
		sr := re.FindAllStringSubmatch(raw, -1)
		for i, k := range sr {
			k[1] = strings.ToLower(k[1])
			k[2] = strings.ToLower(k[2])
			if i == 0 {
				disk_name = k[2]
			}
			disk[k[1]] = k[2]
			if k[1] == "mountpoint" && strings.HasPrefix(k[2], "/") {
				usage := DiskUsage(k[2])
				disk["used"] = string(usage)
			}
		}
		if disk["type"] == "disk" {
			disk_path := fmt.Sprintf("/sys/block/%s/queue/rotational", disk_name)
			buf, err := ioutil.ReadFile(disk_path)
			if err != nil {
				fmt.Println(err)
			} else {
				disk["disk_rotational"] = strings.TrimSpace(string(buf))
			}
		}
		lsblk[disk_name] = disk
	}
	return lsblk

}

func Lsblk() Disks {
	var cmdrun = CmdRunner{}
	rr, err := cmdrun.Run("lsblk", []string{"-P", "-b", "-o", "NAME,KNAME,MODEL,PARTUUID,SIZE,ROTA,TYPE,MOUNTPOINT,MAJ:MIN,PKNAME,FSTYPE"})
	if err != nil {
		fmt.Println(err)
	}
	disks := ParserLsblk(rr)
	return disks
}

func DiskUsage(path string) string {
	usage, _ := disk.Usage(path)
	return strconv.Itoa(int(usage.Used))
}
