package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jawher/mow.cli"
	"github.com/olekukonko/tablewriter"
	"strings"
)

type Project struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Details   string    `json:"details"`
	Settings  string    `json:"settings"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectParam struct {
	Project Project `json:"project"`
}

type SearchString struct {
	SearchString string `json:"search_string"`
}

func createProjectParam(name, details, settings string) *ProjectParam {
	return &ProjectParam{
		Project{
			Name:     name,
			Details:  details,
			Settings: settings,
		},
	}
}

func ReadUserIntInput(prompt string) (int, error) {
	input := ReadUserInput(prompt)
	asInteger, err := strconv.Atoi(input)
	if err != nil {
		return -1, err
	}
	return asInteger, nil
}

func ReadUserInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	resp, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}

	return strings.TrimSpace(resp)
}

func (p *Project) Display() { // String?
	if p == nil {
		return
	}

	var data [][]string
	data = append(data, []string{"Name", p.Name})
	data = append(data, []string{"Details", p.Details})
	data = append(data, []string{"CreatedAt", p.CreatedAt.String()})
	data = append(data, []string{"UpdatedAt", p.UpdatedAt.String()})
	data = append(data, []string{"Settings", string(p.Settings)})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(TerminalWidth())
	table.SetHeader([]string{"Key", "Value"})
	table.SetBorder(false)
	table.AppendBulk(data)
	fmt.Fprintf(os.Stdout, "\n---- Project Details ----\n")
	table.Render()
	fmt.Fprintf(os.Stdout, "\n")
}

func DisplayProjects(projects []Project) {
	if len(projects) == 0 {
		fmt.Println("No projects found")
		return
	}

	if len(projects) == 1 {
		projects[0].Display()
		return
	}

	var data [][]string
	for i := range projects {
		p := projects[i]
		data = append(data, []string{fmt.Sprintf("%d", i+1), p.Name, p.Details})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(TerminalWidth())
	table.SetHeader([]string{"Number", "Name", "Details"})
	table.SetBorder(false)
	table.AppendBulk(data)
	fmt.Fprintf(os.Stdout, "\n")
	table.Render()
	fmt.Fprintf(os.Stdout, "\n")
}

func extractProjectsFromBody(body []byte) ([]Project, error) {
	projects := make([]Project, 0)
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func extractProjectFromBody(body []byte) ([]Project, error) {
	p := &Project{}
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	return []Project{*p}, nil
}

func displayProjectsFromResponse(resp *http.Response, expectedCode int, listExpected bool) error {
	projects, err := extractProjectFromResponse(resp, expectedCode, listExpected)
	if err != nil {
		return err
	}

	if listExpected {
		DisplayProjects(projects)
	} else {
		if len(projects) == 1 {
			projects[0].Display()
		}
	}

	return nil
}

func createProject(cmd *cli.Cmd) {
	cmd.Spec = "[--name]"
	name := cmd.StringOpt("name", "", "Name for the project")

	cmd.Action = func() {
		*name = strings.TrimSpace(*name)
		if *name == "" {
			temp := ReadUserInput("Please provide project name: ")
			name = &temp
			if strings.TrimSpace(*name) == "" {
				fmt.Println("The project name cannot be empty")
				cli.Exit(1)
			}
		} else {
			fmt.Printf("Project Name: %s\n", *name)
		}

		projectDetails := ReadUserInput("Details to the project (optional): ")
		resp, err := Do("/v1/projects", "POST", createProjectParam(*name, projectDetails, ""))
		if err != nil {
			Log.Error(err)
			return
		}
		defer resp.Body.Close()
		displayProjectsFromResponse(resp, http.StatusCreated, false)
	}
}

func settingsProject(cmd *cli.Cmd) {
	cmd.Spec = "[--name] KEY VALUE"
	name := cmd.StringOpt("name", "", "Name for the project")
	key := cmd.StringArg("KEY", "", "Name of the setting")
	value := cmd.StringArg("VALUE", "", "Value of the setting")
	cmd.Action = func() {
		if *name == "" {
			*name, _ = ReadProjectLock()
		}
		p, err := chooseProject(*name, "Which project needs to be updated: ")
		if err == nil {
			resp, err := Do(p.EndPoint(), "PUT", createProjectParam("", "", fmt.Sprintf(`{"%s": "%s"}`, *key, *value)))
			defer resp.Body.Close()
			if err != nil || resp.StatusCode != http.StatusNoContent {
				Log.Error("Something went wrong. Please try again")
			} else {
				fmt.Println("Was successfully updated")
			}
			return
		}
		Log.Error(err)
	}
}

func FindProjects(portion string) (*http.Response, error) {
	if strings.TrimSpace(portion) == "" {
		return Do("/v1/projects", "GET", nil)
	}
	return Do("/v1/projects/search", "GET", &SearchString{portion})
}

func listProjects(cmd *cli.Cmd) {
	cmd.Spec = "[--name]"
	name := cmd.StringOpt("name", "", "Name for the project")

	cmd.Action = func() {
		resp, err := FindProjects(*name)
		if err != nil {
			Log.Error(err)
			return
		}
		defer resp.Body.Close()
		displayProjectsFromResponse(resp, http.StatusOK, true)
	}
}

func chooseProject(portion, message string) (*Project, error) {
	resp, err := FindProjects(portion)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	projects, err := extractProjectFromResponse(resp, http.StatusOK, true)
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, errors.New("No such project. Sorry")
	}

	if len(projects) == 1 {
		return &projects[0], nil
	}

	DisplayProjects(projects)

	choice, err := ReadUserIntInput(message)
	if err != nil {
		return nil, err
	}

	if choice > len(projects) {
		return nil, errors.New("Plese choose a number from the first column")
	}

	return &projects[choice-1], nil
}

func (p *Project) EndPoint() string {
	return fmt.Sprintf("/v1/projects/%d", p.ID)
}

func (p *Project) AssetsUrl() string {
	return p.EndPoint() + "/assets"
}

func (p *Project) AssetstoreUrl() string {
	return p.EndPoint() + "/assetstore"
}

func (p *Project) JobsUrl() string {
	return p.EndPoint() + "/jobs"
}

func showProject(cmd *cli.Cmd) {
	cmd.Spec = "[--name]"
	name := cmd.StringOpt("name", "", "Name of the project")

	cmd.Action = func() {
		if *name == "" {
			*name, _ = ReadProjectLock()
		}
		p, err := chooseProject(*name, "Which project needs to be displayed in detail: ")
		if err == nil {
			resp, err := Do(p.EndPoint(), "GET", nil)
			defer resp.Body.Close()
			if err == nil {
				displayProjectsFromResponse(resp, http.StatusOK, false)
				return
			}
		}
		Log.Error(err)
		return
	}
}

func deleteProject(cmd *cli.Cmd) {
	cmd.Spec = "[--name]"
	name := cmd.StringOpt("name", "", "Name (or part of it) of the project")

	cmd.Action = func() {
		if *name == "" {
			*name, _ = ReadProjectLock()
		}
		p, err := chooseProject(*name, "Please choose the project to be deleted: ")
		if err != nil {
			Log.Error(err)
			return
		}
		p.Delete()
	}
}

func RegisterProjectRoutes(proj *cli.Cmd) {
	proj.Command("create c", "Create a new project", createProject)
	proj.Command("settings", "Create a new project", settingsProject)
	proj.Command("list ls", "List all projects", listProjects)
	proj.Command("show sh", "Show an existing project", showProject)
	proj.Command("delete d", "Delete an existing project", deleteProject)

	proj.Command("build b", "Build a project", buildProject)
	proj.Command("validate v", "Validate a project", validateProject)
	proj.Command("status st", "Show job status of a projct", jobStatusProject)
}
