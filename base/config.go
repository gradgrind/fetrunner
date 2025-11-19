package base

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

var (
	config     map[string]any
	configfile string
	FetCl      string
)

func test_fet() {
	var fetpath string
	fetpath0 := config["FET"]
	if fetpath0 == nil {
		fetpath = "fet-cl"
	} else {
		fetpath = fetpath0.(string)
	}
	cmd := exec.Command(fetpath, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if fetpath != "fet-cl" {
			// Try again, this time without path
			cmd := exec.Command("fet-cl", "--version")
			out, err = cmd.CombinedOutput()
			if err == nil {
				fetpath = "fet-cl"
				goto get_version
			}
		}
		Error.Println("FET not found")
		SetConfig("FET", nil)
		return
	}
get_version:
	version := regexp.MustCompile(`(?m)version +([0-9.]+)`)
	match := version.FindSubmatch(out)
	if match == nil {
		Result("FET_VERSION", "?")
	} else {
		Result("FET_VERSION", match[1])
	}
	SetConfig("FET", fetpath)
	FetCl = fetpath
}

func InitConfig() {
	config = map[string]any{} // an empty config
	dir, dirErr := os.UserConfigDir()
	if dirErr != nil {
		Error.Println("No config location!")
	} else {
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
					configfile = cfile
					saveconfig()
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
	}
	// Check path to `fet-cl`
	test_fet()
}

func SetConfig(key string, val any) {
	val0 := config[key]
	config[key] = val
	Result("CONFIG", fmt.Sprintf("%s=%v", key, val))
	if val0 != val {
		saveconfig()
	}
}

func saveconfig() {
	if len(configfile) == 0 {
		Error.Println("Write config failed: no config file")
		return
	}
	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(configfile, jsonBytes, 0644)
	if err != nil {
		Error.Println("Write config failed:", err)
		configfile = ""
	}
}
