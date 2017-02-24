package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	version "github.com/mcuadros/go-version"
)

const CONFIG_JSON_URL = "https://s3-eu-west-1.amazonaws.com/io-slyft-config/slyft-config.json"

type configJson struct {
	APIVersion struct {
		Min     int `json:"min"`
		Max     int `json:"max"`
		Current int `json:"current"`
	} `json:"api_version"`
	EndPoints []struct {
		Num1 string `json:"1,omitempty"`
		Num2 string `json:"2,omitempty"`
		Num3 string `json:"3,omitempty"`
		Num4 string `json:"4,omitempty"`
	} `json:"end_points"`
	ClientVersion struct {
		Latest string   `json:"latest"`
		Update []string `json:"update"`
	} `json:"client_version"`
}

type UpdateCheckResult struct {
	ShouldUpdate bool
	MustUpdate   bool
}

func getConfigJson() (*configJson, error) {
	config := &configJson{}
	resp, err := getJson(CONFIG_JSON_URL)
	if err != nil {
		return config, err
	}
	defer resp.Body.Close()

	err = ensureValidResponse(resp)
	if err != nil {
		return config, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err := json.Unmarshal(body, &config); err != nil {
		return config, err
	}
	return config, nil
}

func UpdateCheck(appVersion string) error {
	config, err := getConfigJson()
	if err != nil {
		return err
	}
	res := &UpdateCheckResult{}
	if version.Compare(appVersion, config.ClientVersion.Latest, "<") {
		res.ShouldUpdate = true
	}
	if stringInSlice(appVersion, config.ClientVersion.Update) {
		res.MustUpdate = true
	}
	displayUpdateCheck(res)
	if res.MustUpdate == true {
		return errors.New(
			fmt.Sprintf("You need to update your application. Your version: %v, latest version: %v", appVersion, config.ClientVersion.Latest),
		)
	}
	return nil
}

func displayUpdateCheck(res *UpdateCheckResult) {
	if !res.ShouldUpdate && !res.MustUpdate {
		return
	}
	if res.MustUpdate {
		Log.Info("Your version is outdated, you need to update before you can continue.")
		return
	}
	if res.ShouldUpdate {
		Log.Info("A newer version is available, consider updating.")
	}
}

func getJson(url string) (*http.Response, error) {
	b := new(bytes.Buffer)
	req, err := http.NewRequest("GET", url, b)
	if err != nil {
		Log.Critical("Failed to create a request: " + err.Error())
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{}
	return client.Do(req)
}

func ensureValidResponse(resp *http.Response) error {
	if !(resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK) {
		return errors.New(fmt.Sprintf("Server returned no content and status code: %v", resp.StatusCode))
	}
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
