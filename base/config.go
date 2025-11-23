package base

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var (
	config     map[string]string
	configfile string
)

// Check path to `fet-cl`
func (logger Logger) TestFet() {
	var fetpath string
	fetpath0 := config["FET"]
	if fetpath0 == "" {
		fetpath = "fet-cl"
	} else {
		fetpath = fetpath0
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
		logger.Error("FET_NOT_FOUND %s", out)
		logger.SetConfig("FET", "")
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
}

func (logger Logger) InitConfig() {
	config = map[string]string{} // an empty config
	dir, dirErr := os.UserConfigDir()
	if dirErr != nil {
		logger.Error("ConfigFile_NoConfigLocation")
	} else {
		cdir := filepath.Join(dir, "gradgrind")
		err := os.MkdirAll(cdir, 0755)
		if err == nil {
			configfile = filepath.Join(cdir, "fetrunner.conf")
			logger.Result("CONFIG_FILE", configfile)
			// open file
			f, err := os.Open(configfile)
			if err == nil {
				// remember to close the file at the end of the function
				defer f.Close()
				// read the file line by line using scanner
				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					// do something with a line
					t := scanner.Text()
					kv := strings.SplitN(t, "=", 2)
					if len(kv) != 2 {
						logger.Error("ConfigFile_BadLine: %s", t)
					} else {
						config[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
					}
				}
				if err := scanner.Err(); err != nil {
					logger.Error("ConfigFile_ScanError: %s", err)
				}
			}
		} else {
			logger.Error("ConfigFile_LocationNotAccessible: %s", cdir)
		}
	}
}

func (logger Logger) SetConfig(key string, val string) {
	val0 := config[key]
	if val0 != val {
		logger.Info("SET_CONFIG %s=%s", key, val)
		config[key] = val
		logger.saveconfig()
	} else {
		logger.Info("(SET_CONFIG %s=%s)", key, val)
	}
}

func (logger Logger) GetConfig(key string) string {
	return config[key]
}

func (logger Logger) saveconfig() {
	if configfile == "" {
		logger.Error("WriteConfigFailed_NoConfigFile")
		return
	}
	lines := [][]string{}
	for k, v := range config {
		lines = append(lines, []string{k, v})
	}
	slices.SortFunc(lines, func(a, b []string) int {
		return strings.Compare(a[0], b[0])
	})
	// create file
	f, err := os.OpenFile(configfile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Error("WriteConfigFailed: %s", err)
		configfile = ""
	}
	// remember to close the file
	defer f.Close()
	for _, line := range lines {
		_, err := f.WriteString(line[0] + "=" + line[1] + "\n")
		if err != nil {
			logger.Error("WriteConfigLine: %s", err)
		}
	}
}
