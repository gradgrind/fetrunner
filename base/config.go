package base

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	config     map[string]any
	configfile string
	FetCl      string
)

func init() {
	config = map[string]any{} // an empty config
	dir, dirErr := os.UserConfigDir()
	if dirErr != nil {
		Error.Println("No config location!")
		return
	}
	cdir := filepath.Join(dir, "gradgrind")
	err := os.MkdirAll(cdir, 0755)
	if err == nil {
		cfile := filepath.Join(cdir, "fetrunner.json")
		origConfig, err := os.ReadFile(cfile)
		if err != nil {
			if !os.IsNotExist(err) {
				// The user has a config file but we couldn't read it.
				// Report the error and leave the file.
				Error.Println("Config file not readable: ", err)
			} else {
				emsg := saveconfig()
				if len(emsg) != 0 {
					Error.Println(emsg)
				} else {
					configfile = cfile
				}
			}
		} else {
			// Read as JSON
			err = json.Unmarshal(origConfig, &config)
			if err != nil {
				Error.Println("Config file invalid JSON:", err)
			} else {
				configfile = cfile
			}
		}
	} else {
		Error.Println("Config location not accessible: " + cdir)
	}

	// Check path to `fet-cl`
	var fetcl string
	fetcl0, ok := config["fet-cl"]
	if !ok {
		// Seek executable
		fetcl, err = exec.LookPath("fet-cl")
		if err == nil {
			goto have_fetcl
		}
	} else {
		fetcl, ok = fetcl0.(string)
		if ok {
			goto have_fetcl
		}
	}

have_fetcl:
	//TODO: Test fetcl, get the version?

	SetConfig("fet-cl", fetcl)
	FetCl = fetcl
}

func SetConfig(key string, val any) string {
	config[key] = val
	return saveconfig()
}

func saveconfig() string {
	emsg := "Write config failed"
	if len(configfile) != 0 {
		jsonBytes, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(configfile, jsonBytes, 0644)
		if err == nil {
			return ""
		}
		emsg = fmt.Sprint(emsg+":", err)
	}
	return emsg
}
