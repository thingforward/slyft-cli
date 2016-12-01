package main

import "net/http"

func Do(endpoint, method string) (*http.Response, error) {
	url := ServerURL(endpoint)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		Log.Critical("Failed to create a request: " + err.Error())
		return nil, err
	}

	auth, err := readAuthFromConfig()
	if err != nil {
		return nil, err
	}

	addAuthToHeader(&req.Header, auth)
	client := &http.Client{}
	return client.Do(req)
}

func addAuthToHeader(hdr *http.Header, s *SlyftAuth) {
	hdr.Add("access-token", s.AccessToken)
	hdr.Add("client", s.Client)
	hdr.Add("uid", s.Uid)
}
