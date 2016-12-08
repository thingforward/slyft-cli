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
	Settings  string    `json:settings"`
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

func createProjectParam(name, details string) *ProjectParam {
	return &ProjectParam{
		Project{
			Name:    name,
			Details: details,
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
	fmt.Printf(prompt)
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
	data = append(data, []string{"Settings", p.Settings})

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
		resp, err := Do("/v1/projects", "POST", createProjectParam(*name, projectDetails))
		if err != nil {
			Log.Error(err)
			return
		}
		defer resp.Body.Close()
		displayProjectsFromResponse(resp, http.StatusCreated, false)
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

func chooseProject(portion, message string) (int, error) {
	resp, err := FindProjects(portion)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	projects, err := extractProjectFromResponse(resp, http.StatusOK, true)
	if err != nil {
		return -1, err
	}

	if len(projects) == 0 {
		return -1, errors.New("No such project. Sorry")
	}

	if len(projects) == 1 {
		return projects[0].ID, nil
	}

	DisplayProjects(projects)

	choice, err := ReadUserIntInput(message)
	if err != nil {
		return -1, err
	}

	if choice > len(projects) {
		return -1, errors.New("Plese choose a number from the first column")
	}

	return projects[choice-1].ID, nil
}

func projectUrl(pid int) string {
	return fmt.Sprintf("/v1/projects/%d", pid)
}

func findProjectAndApplyMethod(portion, method, message string) (*http.Response, error) {
	chosenId, err := chooseProject(portion, message)
	if err != nil {
		return nil, err
	}
	// if method == DELETE --- then confirm.
	return Do(projectUrl(chosenId), method, nil)
}

func showProject(cmd *cli.Cmd) {
	cmd.Spec = "[--name]"
	name := cmd.StringOpt("name", "", "Name of the project")

	cmd.Action = func() {
		resp, err := findProjectAndApplyMethod(*name, "GET", "Which project needs to be diplayed in detail: ")
		if err != nil {
			Log.Error(err)
			return
		}
		defer resp.Body.Close()
		displayProjectsFromResponse(resp, http.StatusOK, false)
	}
}

func deleteProject(cmd *cli.Cmd) {
	cmd.Spec = "[--name]"
	name := cmd.StringOpt("name", "", "Name (or part of it) of the project")

	cmd.Action = func() {
		resp, err := findProjectAndApplyMethod(*name, "DELETE", "Which project needs to be diplayed in detail: ")
		if err != nil {
			Log.Error(err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNoContent {
			fmt.Println("The project was successfully deleted")
			return
		}

		Log.Infof("Deletion was no successful, try later? (more: expected %v received %v)\n", http.StatusNoContent, resp.StatusCode)
	}
}

func RegisterProjectRoutes(proj *cli.Cmd) {
	proj.Command("create c", "Create a new project", createProject)
	proj.Command("list ls", "List all projects", listProjects)
	proj.Command("show sh", "Show an existing project", showProject)
	proj.Command("delete d", "Delete an existing project", deleteProject)
}
