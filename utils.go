package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/siddontang/go/log"
)

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			ReportError("Getting confirmation", err)
			return false
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

func portableGetUsersHome() string {
	// works on Linux, OSX, Windows cmd and Windows gitbash
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return home
}

func defaultConfigFile() string {
	return filepath.FromSlash(portableGetUsersHome() + "/.slyftrc")
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

func TerminalHeight() int {
	defaultWidth := 24

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
	height, err := strconv.Atoi(all[0])
	if err != nil {
		return defaultWidth
	}

	return height
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
		res := scanner.Text()
		Log.Debugf("Operating on project=%s (from .slyftproject)", res)
		return res, nil
	}
	return "", errors.New("NoProjectLock")
}

func ReportError(context string, err error) {
	fmt.Printf("%s: failed.\n", context)
	if err != nil {
		fmt.Printf("Details: %s\n", err.Error())
		Log.Debugf("%s - failed - %s\n", context, err)
	}
}
