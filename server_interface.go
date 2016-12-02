package main

import (
	"bytes"
	"encoding/json"
	"github.com/urfave/cli"
	"io"
	"net/http"
	"os"
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
		return nil, err
	}

	addAuthToHeader(&req.Header, auth)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
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

// Server check functions. Perhaps not to be included in the final build.
func PingServer(c *cli.Context) error {

	url := ServerURL("/_ping")

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		Log.Fatal(err)
	}
	return nil
}
