package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Project struct {
	Name      string    `json:"name"`
	Details   string    `json:"details"`
	Settings  string    `json:settings"`
	UserId    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectParam struct {
	Project Project `json:"project"`
}

func createProjectParam(name, details string) *ProjectParam {
	return &ProjectParam{
		Project{
			Name:    name,
			Details: details,
		},
	}
}

func ReadUserInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(prompt)
	resp, err := reader.ReadString('\n')
	if err != nil {
		Log.Debug(err.Error())
		return ""
	}

	return resp
}

func DisplayProjects(projects []Project) {
	if len(projects) == 0 {
		fmt.Println("No projects found")
		return
	}

	var data [][]string
	for i := range projects {
		p := projects[i]
		data = append(data, []string{p.Name, p.Details, p.Settings, p.CreatedAt.String(), p.UpdatedAt.String()})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(TerminalWidth())
	table.SetHeader([]string{"Name", "Details", "Settings", "Created At", "Updated At"})
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()
}

func handleProjectResponse(resp *http.Response, expectedCode int, listExpected bool) error {
	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == expectedCode {
		if listExpected {

			projects := make([]Project, 0)
			if err := json.Unmarshal(body, &projects); err != nil {
				Log.Error("Failed to read the server response, please try again: " + err.Error())
				return err
			}
			DisplayProjects(projects)

		} else {
			p := &Project{}
			if err := json.Unmarshal(body, &p); err != nil {
				Log.Error("Failed to read the server response, please try again: " + err.Error())
				return err
			}
			DisplayProjects([]Project{*p})
		}
		return nil
	}

	if err != nil {
		return errors.New(fmt.Sprintf("Failed with following error: %v\n", err))
	}

	return nil
}

func ProjectCreationHandler(c *cli.Context) error {
	Log.Error(ProjectName)
	if ProjectName == "" {
		ProjectName = ReadUserInput("Please provide project name: ")
	}
	projectDetails := ReadUserInput("Deatils to the project: ")

	resp, err := Do("/v1/projects", "POST", createProjectParam(ProjectName, projectDetails))
	if err != nil {
		Log.Error("Creation failed: " + err.Error())
	}
	defer resp.Body.Close()

	return handleProjectResponse(resp, http.StatusCreated, false)
}

func ProjectListHandler(c *cli.Context) error {
	resp, err := Do("/v1/projects", "GET", nil)
	if err != nil {
		Log.Error("Creation failed: " + err.Error())
	}
	defer resp.Body.Close()

	return handleProjectResponse(resp, http.StatusOK, true)
	return nil
}

func ProjectShowHandler(c *cli.Context) error {
	return nil
}

func ProjectDeletionHandler(c *cli.Context) error {
	return nil
}
