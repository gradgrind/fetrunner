//go:build windows

package fet

import (
	"os"
)

func init() {
	FET_CL = "fet-clw.exe"

	fsinfo, err := os.Stat("R:")
	if err == nil && fsinfo.IsDir() {
		TEMPORARY_FOLDER = "R:/fetrunner"
	}
}
