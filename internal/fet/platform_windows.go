//go:build windows

package fet

import (
	"os"
)

func init() {
	fsinfo, err := os.Stat("R:")
	if err == nil && fsinfo.IsDir() {
		TEMPORARY_FOLDER = "R:/fetrunner"
	}
}
