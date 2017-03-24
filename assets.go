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
	data = append(data, []string{"Number", "Name", "UpdatedAt", "Project Name", "Origin"})
	for i := range assets {
		a := assets[i]
		data = append(data, []string{fmt.Sprintf("%d", i+1), a.Name, a.UpdatedAt.String(), a.ProjectName, a.Origin})
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
		return nil, errors.New("No assets found.")
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

func putAsset(id int, assetParam *AssetParam, p *Project) error {
	resp, err := Do(p.AssetUrl(id), "PUT", assetParam)
	if err != nil {
		ReportError("Contacting server", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		Log.Debugf("resp.Code=%#v / expected=%d", resp.StatusCode, http.StatusNoContent)
		return errors.New(fmt.Sprintf("Failed with the wrong code: %v. (expected %v)\n", resp.StatusCode, http.StatusNoContent))
	}

	return nil
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
		if len(assets) == 1 {
			// we have a duplicate.
			if askForConfirmation("The asset already exists. Do you want to overwrite it?") {
				assets[0].Display()
				return putAsset(assets[0].ID, assetParam, p)
			}
			return nil
		}
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

func getAllAssets(p *Project) ([]Asset, error) {
	resp, err := Do(p.AssetsUrl(), "GET", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	assets, err := extractAssetFromResponse(resp, http.StatusOK, true)

	if err != nil {
		return nil, err
	}

	if len(assets) == 0 {
		return nil, errors.New("No assets found.")
	}

	return assets, nil
}

func removeSingleFileFromAsset(assets []Asset, file string, p *Project) {

	for _, asset := range assets {
		if asset.Name == file {
			fmt.Printf("Deleting asset %s\n", file)

			resp, err := Do(asset.EndPoint(), "DELETE", &AssetNameString{file})
			Log.Debugf("resp=%#v", resp)
			if err != nil {
				Log.Debugf("err=%#v", err)
				ReportError("Removing asset", err)
				return
			}
			defer resp.Body.Close()

			if err != nil || resp.StatusCode != http.StatusNoContent {
				if resp.StatusCode == http.StatusNotFound {
					fmt.Printf("Unable to delete asset with name %s\n", file)
				} else {
					fmt.Printf("Something went wrong. Please try again. (ResponseCode: %d)\n", resp.StatusCode)
				}
			} else {
				fmt.Println("Was successfully deleted")
			}

			return
		}
	}

	fmt.Printf("Unable to delete asset with name %s\n", file)

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
	cmd.Spec = "[--project] [--file] [INPUTFILES...]"
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

		didProcessSomething := false
		if file != nil && *file != "" {
			*file = strings.TrimSpace(*file)
			err := readFileAndPostAsset(*file, p)
			if err == nil {
				didProcessSomething = true
			}
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
					err := readFileAndPostAsset(singleFile, p)
					if err == nil {
						didProcessSomething = true
					}
				}
			}
		}

		if didProcessSomething == false {
			fmt.Println("Need to specify --file or give valid files as arguments. Did not upload anything")
		}

	}
}

func getAsset(cmd *cli.Cmd) {
	cmd.Spec = "[--project] [--file] [FILES...]"
	name := cmd.StringOpt("project p", "", "Name (or part of it) of a project")
	file := cmd.StringOpt("file f", "", "name of the asset to be downloaded")
	files := cmd.StringsArg("FILES", nil, "Multiple assets to download")

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

		didProcessSomething := false
		if file != nil && *file != "" {
			*file = strings.TrimSpace(*file)
			getAssetAndSaveToFile(*file, p)
			didProcessSomething = true
		}
		if files != nil {
			for _, singleFile := range *files {
				getAssetAndSaveToFile(singleFile, p)
				didProcessSomething = true
			}
		}
		if didProcessSomething == false {
			fmt.Println("Need to specify --file or give valid files as arguments. Did not download anything")
		}
	}
}

func (ass *Asset) EndPoint() string {
	return fmt.Sprintf("/v1/projects/%d/assets/%d", ass.ProjectId, ass.ID)
}

func removeAsset(cmd *cli.Cmd) {
	cmd.Spec = "[--project] [--count] [FILES...]"
	name := cmd.StringOpt("project p", "", "Name (or part of it) of a project")
	count := cmd.IntOpt("count n", 0, "Choose from the last 'count' assets of the project (if project is not specified, select from all)")
	files := cmd.StringsArg("FILES", nil, "Name(s) of files to delete from asset list")

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

			if files != nil && len(*files) > 0 {
				assets, err := getAllAssets(p)
				if err == nil {
					// locate and delete files
					for _, singleFile := range *files {
						removeSingleFileFromAsset(assets, singleFile, p)
					}
				} else {
					Log.Debugf("%#v", err)
					fmt.Println("Error querying assets.")
				}
			} else {
				// choose interactive
				ass, err = chooseAsset(p.AssetsUrl(), true, "Which one shall be removed: ", *count)

				if err != nil {
					ReportError("Choosing the asset", err)
					return
				}
				Log.Debugf("Choosen asset %#v", ass)

				DeleteApiModel(ass)
			}
		}

	}
}

func RegisterAssetRoutes(proj *cli.Cmd) {
	SetupLogger()

	proj.Command("add a", "Add asset to a project", addAsset)
	proj.Command("list ls", "List your assets", listAssets)
	proj.Command("get g", "Download a single asset", getAsset)
	proj.Command("delete d", "Remove and asset from a project", removeAsset)
}
