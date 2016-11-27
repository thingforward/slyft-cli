package main

import (
	"io"
	"net/http"
	"os"

	"github.com/urfave/cli"
)

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
