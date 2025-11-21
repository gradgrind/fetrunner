package main

import (
	"os"
	"strings"
)

func Dispatch(cmd0 string) string {
	var result string
	cmdsplit := strings.Fields(cmd0)
	switch cmd := cmdsplit[0]; cmd {

	case "CONFIG_DIR":
		dir, dirErr := os.UserConfigDir()
		if dirErr == nil {
			result = "> config dir: " + dir
		} else {
			result = "! No config dir"
		}

	case "CONFIG_INIT":
		//TODO: Needs adapting, the call is now
		// logger.InitConfig()
		//was base.InitConfig()

	default:
		result = "! Invalid command: " + cmd0

	}
	return result
}
