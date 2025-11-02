//go:build linux

package fet

import (
	"fmt"
	"os"
)

func init() {
	fmt.Println("??? 1")
	fsinfo, err := os.Stat("/dev/shm")
	if err == nil && fsinfo.IsDir() {
		fmt.Println("??? 2")
		TEMPORARY_FOLDER = "/dev/shm/fetrunner"
	} else {
		fmt.Println("??? ", err)
	}
}
