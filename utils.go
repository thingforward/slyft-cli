package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"

	"strconv"

	"github.com/siddontang/go/log"
)

func defaultConfigFile() string {
	return os.Getenv("HOME") + "/.slyftrc" // this will fail for windows
}

func readFile(fileName string) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		Log.Debug("Could not open %s: %v\n", fileName, err)
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Debugf("Reading %s failed: %v.\n", file.Name(), err)
		return nil, err
	}
	return bytes, nil
}

func readConfig() (*SlyftRC, error) {
	config, err := readFile(defaultConfigFile())

	var sr SlyftRC
	if err != nil {
		Log.Info("Cannot read config file" + defaultConfigFile() + ": " + err.Error())
		return &sr, err
	}

	if err := json.Unmarshal(config, &sr); err != nil {
		Log.Debugf("Cannot parse config file: " + defaultConfigFile() + ": " + err.Error())
		return &sr, err
	}

	return &sr, nil
}

func writeAuthToConfig(sa *SlyftAuth) error {
	sr, _ := readConfig()
	// note -- we are ignoring the error here.

	sr.Auth = *sa
	newConfig, err := json.MarshalIndent(sr, "", "	")
	if err != nil {
		Log.Error("Failure to update config file: " + defaultConfigFile())
		return err
	}

	err = ioutil.WriteFile(defaultConfigFile(), newConfig, 0640)
	if err != nil {
		Log.Error("Failure to write config file: " + defaultConfigFile())
		return err
	}
	return nil
}

func readAuthFromConfig() (*SlyftAuth, error) {
	sr, err := readConfig()
	if err != nil {
		Log.Error("You don't seem to be logged in. Failed to read your config: " + err.Error())
		return nil, err
	}

	return &sr.Auth, nil
}

func deactivateLogin() {
	var sa SlyftAuth
	writeAuthToConfig(&sa)
}

func TerminalWidth() int {
	defaultWidth := 96 // brave new world. Not any more 80x24

	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return defaultWidth
	}
	re := regexp.MustCompile("[0-9]+")
	all := re.FindAllString(string(out), -1)
	if len(all) < 2 {
		return defaultWidth
	}
	width, err := strconv.Atoi(all[1])
	if err != nil {
		return defaultWidth
	}

	return width
}

// looks in current directory for a file `.slyftproject`,
// reads the first line and returns it (as a replacename
// for --name parameter whereever a project name is required)
func ReadProjectLock() (string, error) {
	f, err := os.Open(".slyftproject")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", errors.New("NoProjectLock")
}
