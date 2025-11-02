//go:build linux

package fet

import (
	"os"
)

func init() {
	fsinfo, err := os.Stat("/dev/shm")
	if err == nil && fsinfo.IsDir() {
		TEMPORARY_FOLDER = "/dev/shm/fetrunner"
	}
}
