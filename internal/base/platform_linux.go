//go:build linux

package base

import (
	"os"
)

func init() {
	fsinfo, err := os.Stat("/dev/shm")
	if err == nil && fsinfo.IsDir() {
		TEMPORARY_BASEDIR = "/dev/shm"
	}
}
