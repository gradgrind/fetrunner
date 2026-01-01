//go:build windows

package base

import (
	"os"
)

func init() {
	fsinfo, err := os.Stat("R:")
	if err == nil && fsinfo.IsDir() {
		TEMPORARY_BASEDIR0 = "R:"
	}
}
