package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func extractProjectFromResponse(resp *http.Response, expectedCode int, listExpected bool) ([]Project, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != expectedCode {
		return nil, errors.New(fmt.Sprintf("Failed with the wrong code: %v. (expected %v)\n", resp.StatusCode, expectedCode))
	}

	if listExpected {
		return extractProjectsFromBody(body)
	}

	return extractProjectFromBody(body)
}

func extractAssetFromResponse(resp *http.Response, expectedCode int, listExpected bool) ([]Asset, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != expectedCode {
		return nil, errors.New(fmt.Sprintf("Failed with the wrong code: %v. (expected %v)\n", resp.StatusCode, expectedCode))
	}

	if listExpected {
		return extractAssetsFromBody(body)
	}

	return extractAssetFromBody(body)
}

func (ass *Asset) Delete() {
	if ass == nil {
		return
	}
	confirm := ReadUserInput("Are you sure? [no]: ")
	if confirm == "yes" || confirm == "y" || confirm == "Y" || confirm == "YES" {
		resp, err := Do(ass.EndPoint(), "DELETE", nil)
		if err != nil || resp.StatusCode != http.StatusNoContent {
			Log.Error("Something went wrong. Please try again")
		} else {
			Log.Error("Was successfully deleted")
		}
	} else {
		Log.Error("Good decision!")
	}
}

func (p *Project) Delete() {
	if p == nil {
		return
	}
	confirm := ReadUserInput("Are you sure? [no]: ")
	if confirm == "yes" || confirm == "y" || confirm == "Y" || confirm == "YES" {
		resp, err := Do(p.EndPoint(), "DELETE", nil)
		if err != nil || resp.StatusCode != http.StatusNoContent {
			Log.Error("Something went wrong. Please try again")
		} else {
			Log.Error("Was successfully deleted")
		}
	} else {
		Log.Error("Good decision!")
	}
}
