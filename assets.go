package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	cli "github.com/jawher/mow.cli"
)

type Asset struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	ProjectId   int       `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Origin      string    `json:"origin"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (a *Asset) getName() string {
	return a.Name
}

func (a *Asset) Display() { // String?
	if a == nil {
		return
	}

	data := [][]string{
		[]string{"Key", "Value"},
		[]string{"Name", a.Name},
		[]string{"ProjectId", fmt.Sprintf("%d", a.ProjectId)},
		[]string{"ProjectName", a.ProjectName},
		[]string{"Origin", a.Origin},
		[]string{"CreatedAt", a.CreatedAt.String()},
		[]string{"UpdatedAt", a.UpdatedAt.String()},
	}

	fmt.Fprintf(os.Stdout, "%s%s",
		markdownHeading("Asset Details", 1),
		markdownTable(&data))
}

func DisplayAssets(assets []Asset) {
	if len(assets) == 0 {
		fmt.Println("No assets found")
		return
	}

	if len(assets) == 1 {
		assets[0].Display()
		return
	}

	var data [][]string
	data = append(data, []string{"Number", "Name", "Project Name", "Origin"})
	for i := range assets {
		a := assets[i]
		data = append(data, []string{fmt.Sprintf("%d", i+1), a.Name, a.ProjectName, a.Origin})
	}

	fmt.Fprintf(os.Stdout, markdownTable(&data))
}

func extractAssetsFromBody(body []byte) ([]Asset, error) {
	Log.Debugf("body=%v", string(body))
	assets := make([]Asset, 0)
	if err := json.Unmarshal(body, &assets); err != nil {
		return nil, err
	}
	return assets, nil
}

func extractAssetFromBody(body []byte) ([]Asset, error) {
	Log.Debugf("body=%v", string(body))
	a := &Asset{}
	if err := json.Unmarshal(body, &a); err != nil {
		return nil, err
	}
	return []Asset{*a}, nil
}

func chooseAsset(endpoint string, askUser bool, message string, count int) (*Asset, error) {
	resp, err := Do(endpoint, "GET", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	assets, err := extractAssetFromResponse(resp, http.StatusOK, true)

	if err != nil {
		return nil, err
	}

	if len(assets) == 0 {
		return nil, errors.New("No asset. Sorry")
	}
	Log.Debugf("assets=%+v", assets)

	if count != 0 {
		start := len(assets) - count
		if start < 0 {
			start = 0
		}
		assets = assets[start:]
	}

	DisplayAssets(assets)

	if len(assets) == 1 {
		return &assets[0], nil
	}

	if !askUser {
		return nil, nil
	}

	choice, err := ReadUserIntInput(message)
	if err != nil {
		return nil, err
	}

	if choice > len(assets) {
		return nil, errors.New("Plese choose a number from the first column")
	}

	return &assets[choice-1], nil
}

type AssetPost struct {
	Name  string `json:"name"`
	Asset string `json:"asset"` // note: this will be base64 string
}

type AssetParam struct {
	Asset AssetPost `json:"asset"`
}

type AssetNameString struct {
	AssetNameString string `json:"asset_name"`
}

func creatAssetParam(file string) (*AssetParam, error) {
	// read the file content (use ioutil)
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	mimeType, err := preflightAsset(&bytes, file)
	if err != nil {
		return nil, err
	}

	return &AssetParam{
		AssetPost{
			Name:  file,
			Asset: "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(bytes),
		},
	}, nil
}

func readFileAndPostAsset(file string, p *Project) error {
	assetParam, err := creatAssetParam(file)
	if err != nil {
		ReportError("Creating request", err)
		return err
	}

	resp, err := Do(p.AssetsUrl(), "POST", assetParam)
	if err != nil {
		ReportError("Contacting server", err)
		return err
	}
	defer resp.Body.Close()
	assets, err := extractAssetFromResponse(resp, http.StatusCreated, false)
	if err != nil {
		ReportError("Creating asset", err)
		return err
	}

	if len(assets) == 1 {
		assets[0].Display()
	}
	return nil
}

func getAssetAndSaveToFile(file string, p *Project) {
	resp, err := Do(p.AssetstoreUrl(), "GET", &AssetNameString{file})
	if err != nil {
		ReportError("Downloading asset", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		// stream body to file of this name
		out, err := os.Create(file)
		if err != nil {
			ReportError("Creating asset file", err)
			return
		}
		defer out.Close()
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			ReportError("Writing asset file", err)
			return
		}
		fmt.Printf("Downloaded %s\n", file)
	} else {
		ReportError("Downloading asset", nil)
	}
}

func listAssets(cmd *cli.Cmd) {
	cmd.Spec = "[--project] | [--all]"
	name := cmd.StringOpt("project p", "", "Name (or part of it) of a project")
	all := cmd.BoolOpt("all a", false, "Fetch details of all your assets (do not combine with -p)")

	cmd.Action = func() {
		*name = strings.TrimSpace(*name)
		if *all {
			chooseAsset("/v1/assets", false, "", 0)
			return
		} else {
			if *name == "" {
				*name, _ = ReadProjectLock()
			}
		}

		// first get the project, then get the pid, and make the call.
		p, err := chooseProject(*name, "Which project's assets would you like to see: ")
		if err != nil {
			ReportError("Choosing the project", err)
			return
		}
		if _, err = chooseAsset(p.AssetsUrl(), false, "", 0); err != nil {
			ReportError("Choosing the asset", err)
		}
	}
}

func addAsset(cmd *cli.Cmd) {
	cmd.Spec = "[--project] [--file] INPUTFILES..."
	name := cmd.StringOpt("project p", "", "Name (or part of it) of a project")
	// --file is kept as documentation relates on it, but will be deprecated
	file := cmd.StringOpt("file f", "", "path to the file which you want as an asset")
	files := cmd.StringsArg("INPUTFILES", nil, "Multiple files to upload as assets")

	cmd.Action = func() {
		*name = strings.TrimSpace(*name)

		if *name == "" {
			*name, _ = ReadProjectLock()
		}

		// first get the project, then get the pid, and make the call.
		p, err := chooseProject(*name, "Add asset to: ")
		if err != nil {
			ReportError("Choosing the project", err)
			return
		}

		if file != nil && *file != "" {
			*file = strings.TrimSpace(*file)
			readFileAndPostAsset(*file, p)
		}
		if files != nil {
			for _, singleFile := range *files {
				f, err := os.Open(singleFile)
				fi, err := f.Stat()
				defer f.Close()
				switch {
				case err != nil:
					fmt.Printf("Unable to read from %s, skipping\n", singleFile)
				case fi.IsDir():
					fmt.Printf("Is a directory: %s, skipping\n", singleFile)
				default:
					fmt.Printf("Uploading %s ...\n", singleFile)
					readFileAndPostAsset(singleFile, p)
				}
			}
		}
	}
}

func getAsset(cmd *cli.Cmd) {
	cmd.Spec = "[--project] --file"
	name := cmd.StringOpt("project p", "", "Name (or part of it) of a project")
	file := cmd.StringOpt("file f", "", "name of the asset to be downloaded")

	cmd.Action = func() {
		*name = strings.TrimSpace(*name)

		if *name == "" {
			*name, _ = ReadProjectLock()
		}
		// first get the project, then get the pid, and make the call.
		p, err := chooseProject(*name, "Download asset from: ")
		if err != nil {
			ReportError("Choosing the project", err)
			return
		}

		*file = strings.TrimSpace(*file)
		getAssetAndSaveToFile(*file, p)
	}
}

func (ass *Asset) EndPoint() string {
	return fmt.Sprintf("/v1/projects/%d/assets/%d", ass.ProjectId, ass.ID)
}

func removeAsset(cmd *cli.Cmd) {
	cmd.Spec = "[--project] [--count]"
	name := cmd.StringOpt("project p", "", "Name (or part of it) of a project")
	count := cmd.IntOpt("count n", 1, "Choose from the last 'count' assets of the project (if project is not specified, select from all)")

	cmd.Action = func() {
		*name = strings.TrimSpace(*name)
		if *name == "" {
			*name, _ = ReadProjectLock()
		}

		var ass *Asset
		var err error
		if *name == "" {
			ass, err = chooseAsset("/v1/assets", true, "Which one shall be deleted: ", *count)
		} else {
			// first get the project, then get the pid, and make the call.
			p, err2 := chooseProject(*name, "Which project's assets would you like to see: ")
			if err2 != nil {
				ReportError("Choosing the project", err)
				return
			}
			ass, err = chooseAsset(p.AssetsUrl(), true, "Which one shall be removed: ", *count)
		}

		if err != nil {
			ReportError("Choosing the asset", err)
			return
		}

		DeleteApiModel(ass)
	}
}

func updateAssets(cmd *cli.Cmd) {
	cmd.Spec = "[--project] | [--all]"
	proj := cmd.StringOpt("project p", "", "Name (or part of it) of a project")
	all := cmd.BoolOpt("all a", false, "Fetch details of all your assets (do not combine with -p)")

	cmd.Action = func() {
		name := *proj
		name = strings.TrimSpace(name)

		//only worry if --all is not set and --project is empty
		if *all == false && name == "" {
			var err error
			name, err = ReadProjectLock()
			if err != nil {
				ReportError("--all flag not set, no --project specified, no project lock found", err)
				return
			}
		}

		endpoint := "/v1/assets"
		resp, err := Do(endpoint, "GET", nil)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		assets, err := extractAssetFromResponse(resp, http.StatusOK, true)

		if err != nil || len(assets) == 0 {
			return
		}

		assetTable := [][]string{[]string{"ID", "Name", "ProjectId", "ProjectName", "Origin", "CreatedAt", "UpdatedAt", "Status"}}
		rows := 0

		for _, a := range assets {

			//project selection
			if *all == false && strings.Contains(a.ProjectName, name) == false {
				continue
			}

			if updateAvailable(a.Name, a.UpdatedAt) == false {
				continue
			}

			p, err := FindProjectById(a.ProjectId)
			if err != nil {
				fmt.Printf("Project ID %d not found\n", a.ProjectId)
				continue
			}

			err = readFileAndPostAsset(a.Name, p)
			if err != nil {
				fmt.Println("Error on readFileAndPostAsset")
				continue
			}

			row := []string{strconv.Itoa(a.ID), a.Name, strconv.Itoa(a.ProjectId), a.ProjectName, a.Origin, a.CreatedAt.String(), a.UpdatedAt.String(), "Update"}
			assetTable = append(assetTable, row)
			rows++
		}

		if rows == 0 {
			return
		}

		plural := "s"
		if rows == 1 {
			plural = ""
		}

		fmt.Fprintf(os.Stdout, "%s%s",
			markdownHeading(fmt.Sprintf("Updated %d asset%s", rows, plural), 1),
			markdownTable(&assetTable))
	}
}

func updateAvailable(name string, updatedAt time.Time) bool {
	info, err := os.Stat(name)
	if err != nil || info.ModTime().After(updatedAt) == false {
		return false
	}
	return true
}

func RegisterAssetRoutes(proj *cli.Cmd) {
	SetupLogger()

	proj.Command("add a", "Add asset to a project", addAsset)
	proj.Command("list ls", "List your assets", listAssets)
	proj.Command("get g", "Download a single asset", getAsset)
	proj.Command("delete d", "Remove and asset from a project", removeAsset)
	proj.Command("update u", "Update assets for a project", updateAssets)
}
