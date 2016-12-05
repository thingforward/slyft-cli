package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	cli "github.com/jawher/mow.cli"
	"github.com/olekukonko/tablewriter"
)

type Asset struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	ProjectId   int       `json:"project_id"`
	ProjectName string    `json:"project_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Asset       string    `json:"asset"`
	Url         string    `json:"url"`
}

func (a *Asset) Display() { // String?
	if a == nil {
		return
	}

	var data [][]string
	data = append(data, []string{"Name", a.Name})
	data = append(data, []string{"ProjectId", fmt.Sprintf("%d", a.ProjectId)})
	data = append(data, []string{"ProjectName", a.ProjectName})
	data = append(data, []string{"CreatedAt", a.CreatedAt.String()})
	data = append(data, []string{"UpdatedAt", a.UpdatedAt.String()})
	data = append(data, []string{"Asset", a.Asset})
	data = append(data, []string{"Url", a.Url})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(TerminalWidth())
	table.SetHeader([]string{"Key", "Value"})
	table.SetBorder(false)
	table.AppendBulk(data)
	fmt.Fprintf(os.Stdout, "\n---- Asset Details ----\n")
	table.Render()
	fmt.Fprintf(os.Stdout, "\n")
}

func DisplayAssets(assets []Asset) {
	if len(assets) == 0 {
		fmt.Println("No assets found")
		return
	}

	var data [][]string
	for i := range assets {
		a := assets[i]
		data = append(data, []string{fmt.Sprintf("%d", i+1), a.Name, a.ProjectName, a.Url})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(TerminalWidth())
	table.SetHeader([]string{"Number", "Name", "Project Name", "URL"})
	table.SetBorder(false)
	table.AppendBulk(data)
	fmt.Fprintf(os.Stdout, "\n")
	table.Render()
	fmt.Fprintf(os.Stdout, "\n")
}

func extractAssetsFromBody(body []byte) ([]Asset, error) {
	assets := make([]Asset, 0)
	if err := json.Unmarshal(body, &assets); err != nil {
		return nil, err
	}
	return assets, nil
}

func extractAssetFromBody(body []byte) ([]Asset, error) {
	a := &Asset{}
	if err := json.Unmarshal(body, &a); err != nil {
		return nil, err
	}
	return []Asset{*a}, nil
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

func displayAssetsFromResponse(resp *http.Response, expectedCode int, listExpected bool) error {
	assets, err := extractAssetFromResponse(resp, expectedCode, listExpected)
	if err != nil {
		return err
	}

	if listExpected {
		DisplayAssets(assets)
	} else {
		if len(assets) == 1 {
			assets[0].Display()
		}
	}

	return nil
}

func getAndDisplayAssets(endpoint string) {
	resp, err := Do(endpoint, "GET", nil)
	if err != nil {
		Log.Error(err)
		return
	}
	defer resp.Body.Close()
	displayAssetsFromResponse(resp, http.StatusOK, true)
}

func assetEndPointForProjectId(pid int) string {
	return fmt.Sprintf("/v1/projects/%d/assets", pid)
}

func listAssets(cmd *cli.Cmd) {
	cmd.Spec = "[--project] | [--all]"
	name := cmd.StringOpt("project p", "", "Name (or part of it) of a project")
	all := cmd.BoolOpt("all a", false, "Fetch details of all your assets (do not combine with -p)")

	cmd.Action = func() {
		*name = strings.TrimSpace(*name)
		if *all || *name == "" {
			getAndDisplayAssets("/v1/assets")
			return
		}

		// first get the project, then get the pid, and make the call.
		projectId, err := chooseProject(*name, "Which project's assets would you like to see?")
		if err != nil {
			Log.Error(err)
			return
		}
		getAndDisplayAssets(assetEndPointForProjectId(projectId))
	}
}

func RegisterAssetRoutes(proj *cli.Cmd) {
	//proj.Command("add a", "Add asset to a project", addAsset)
	proj.Command("list ls", "List your assets", listAssets)
	//proj.Command("show sh", "Show details of an existing project", func(cmd *cli.Cmd) { cmd.Action = showAsset })
	//proj.Command("delete d", "Remove and asset from a project", func(cmd *cli.Cmd) { cmd.Action = removeAsset })
}
