package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

func DisplayProjects(projects []Project) {
	if len(projects) == 0 {
		fmt.Println("No projects found")
		return
	}

	var data [][]string
	for i := range projects {
		p := projects[i]
		data = append(data, []string{fmt.Sprintf("%d", i+1), p.Name, p.Details, p.Settings, p.CreatedAt.String(), p.UpdatedAt.String()})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(TerminalWidth())
	table.SetHeader([]string{"Number", "Name", "Details", "Settings", "Created At", "Updated At"})
	table.SetBorder(false)
	table.AppendBulk(data)
	fmt.Fprintf(os.Stdout, "\n")
	table.Render()
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

func extractProjectFromResponse(resp *http.Response, expectedCode int, listExpected bool) ([]Project, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != expectedCode {
		return nil, errors.New(fmt.Sprintf("Failed with the wrong code: %v. (expected %v)\n", resp.StatusCode, expectedCode))
	}

	if listExpected {
		return extractProjectsFromBody(body)
	}

	return extractProjectFromBody(body)
}

func displayProjectsFromResponse(resp *http.Response, expectedCode int, listExpected bool) error {
	projects, err := extractProjectFromResponse(resp, expectedCode, listExpected)
	if err != nil {
		return err
	}

	DisplayProjects(projects)
	return nil
}

func createProject(cmd *cli.Cmd) {
	cmd.Spec = "[--name]"
	name := cmd.StringOpt("name n", "", "Name for the project")

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

func listProjects() {
	resp, err := Do("/v1/projects", "GET", nil)
	if err != nil {
		Log.Error(err)
		return
	}
	defer resp.Body.Close()
	displayProjectsFromResponse(resp, http.StatusOK, true)
}

func chooseProject(message string) (int, error) {
	resp, err := Do("/v1/projects", "GET", nil)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	projects, err := extractProjectFromResponse(resp, http.StatusOK, true)
	if err != nil {
		return -1, err
	}

	DisplayProjects(projects)
	if len(projects) == 0 {
		return -1, errors.New("Nothing to process")
	}

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

func findProjectAndApplyMethod(method string, message string) (*http.Response, error) {
	chosenId, err := chooseProject(message)
	if err != nil {
		return nil, err
	}
	return Do(projectUrl(chosenId), method, nil)
}

func showProject() {
	resp, err := findProjectAndApplyMethod("GET", "Which project needs to be diplayed in detail: ")
	if err != nil {
		Log.Error(err)
		return
	}
	defer resp.Body.Close()
	displayProjectsFromResponse(resp, http.StatusOK, false)
}

func deleteProject() {
	resp, err := findProjectAndApplyMethod("DELETE", "Which project needs to be diplayed in detail: ")
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

func RegisterProjectRoutes(proj *cli.Cmd) {
	proj.Command("create c", "Create a new project", createProject)
	proj.Command("list ls", "List all projects", func(cmd *cli.Cmd) { cmd.Action = listProjects })
	proj.Command("show sh", "Show an existing project", func(cmd *cli.Cmd) { cmd.Action = showProject })
	proj.Command("delete d", "Delete an existing project", func(cmd *cli.Cmd) { cmd.Action = deleteProject })
}
