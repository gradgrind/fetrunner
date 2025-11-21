package base

import (
	"encoding/json"
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

func (logger Logger) test_fet() {
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
		logger.Error("FET not found")
		logger.SetConfig("FET", nil)
		return
	}
get_version:
	version := regexp.MustCompile(`(?m)version +([0-9.]+)`)
	match := version.FindSubmatch(out)
	if match == nil {
		logger.Result("FET_VERSION", "?")
	} else {
		logger.Result("FET_VERSION", string(match[1]))
	}
	logger.SetConfig("FET", fetpath)
	FetCl = fetpath
}

func (logger Logger) InitConfig() {
	config = map[string]any{} // an empty config
	dir, dirErr := os.UserConfigDir()
	if dirErr != nil {
		logger.Error("No config location!")
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
					logger.Error("Config file not readable: %v", err)
				} else {
					configfile = cfile
					logger.saveconfig()
				}
			} else {
				// Read as JSON
				err = json.Unmarshal(origConfig, &config)
				if err != nil {
					logger.Error("Config file invalid JSON: %v", err)
				} else {
					configfile = cfile
				}
			}
		} else {
			logger.Error("Config location not accessible: %s", cdir)
		}
	}
	// Check path to `fet-cl`
	logger.test_fet()
}

func (logger Logger) SetConfig(key string, val any) {
	val0 := config[key]
	config[key] = val
	logger.Info("CONFIG %s=%v", key, val)
	if val0 != val {
		logger.saveconfig()
	}
}

func (logger Logger) saveconfig() {
	if len(configfile) == 0 {
		logger.Error("Write config failed: no config file")
		return
	}
	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(configfile, jsonBytes, 0644)
	if err != nil {
		logger.Error("Write config failed: %v", err)
		configfile = ""
	}
}
