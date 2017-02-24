package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func Do(resource, method string, params interface{}) (*http.Response, error) {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(params)
	req, err := http.NewRequest(method, ServerURL(resource), b)

	if err != nil {
		Log.Critical("Failed to create a request: " + err.Error())
		return nil, err
	}

	auth, err := readAuthFromConfig()
	if err != nil {
		fmt.Println("You do not seem to be logged in. Please do a `slyft user login`")
		return nil, err
	}
	if !auth.GoodForLogin() {
		fmt.Println("You do not seem to be logged in. Please do a `slyft user login`")
		return nil, errors.New("Not logged in.")
	}

	addAuthToHeader(&req.Header, auth)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{}

	//Log.Debugf("auth=%#v", auth)
	Log.Debugf("req=%#v", req)

	resp, err := client.Do(req)
	Log.Debugf("resp=%#v", resp)

	if err != nil {
		Log.Debugf("err=%#v", err)
	}

	return resp, err
}

func DoNoAuth(resource, method string, params interface{}) (*http.Response, error) {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(params)
	req, err := http.NewRequest(method, ServerURL(resource), b)

	if err != nil {
		Log.Critical("Failed to create a request: " + err.Error())
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	Log.Debugf("req=%#v", req)
	Log.Debugf("b=%#v", b)
	client := &http.Client{}
	return client.Do(req)
}

func addAuthToHeader(hdr *http.Header, s *SlyftAuth) {
	hdr.Add("access-token", s.AccessToken)
	hdr.Add("client", s.Client)
	hdr.Add("uid", s.Uid)
}

func ServerURL(endpoint string) string {
	return BackendBaseUrl + endpoint
}
