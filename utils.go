package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

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

func writeAuthToConfig(sa *SlyftAuth) error {
	config, err := readFile(defaultConfigFile())
	var sr SlyftRC
	if err != nil {
		Log.Info("A new config file will be created at " + defaultConfigFile())
	} else {
		if err2 := json.Unmarshal(config, &sr); err != nil {
			Log.Info("The config file will be updated" + defaultConfigFile())
			return err2
		}
	}

	sr.Auth = *sa
	newConfig, err := json.MarshalIndent(sr, "", "	")
	if err != nil {
		Log.Error("Failure to update config file: " + defaultConfigFile())
		return err
	}

	err = ioutil.WriteFile(defaultConfigFile(), newConfig, 0644)
	if err != nil {
		Log.Error("Failure to write config file: " + defaultConfigFile())
		return err
	}
	return nil
}
