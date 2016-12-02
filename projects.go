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

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
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

func ProjectCreationHandler(c *cli.Context) error {
	if ProjectName == "" {
		ProjectName = ReadUserInput("Please provide project name: ")
	}
	projectDetails := ReadUserInput("Deatils to the project: ")

	resp, err := Do("/v1/projects", "POST", createProjectParam(ProjectName, projectDetails))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return displayProjectsFromResponse(resp, http.StatusCreated, false)
}

func ProjectListHandler(c *cli.Context) error {
	resp, err := Do("/v1/projects", "GET", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return displayProjectsFromResponse(resp, http.StatusOK, true)
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

func ProjectShowHandler(c *cli.Context) error {
	resp, err := findProjectAndApplyMethod("GET", "Which project needs to be diplayed in detail: ")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return displayProjectsFromResponse(resp, http.StatusOK, false)
}

func ProjectDeletionHandler(c *cli.Context) error {
	resp, err := findProjectAndApplyMethod("DELETE", "Which project needs to be diplayed in detail: ")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		fmt.Println("The project was successfully deleted")
		return nil
	}

	Log.Infof("Deletion was no successful, try later? (more: expected %v received %v)\n", http.StatusNoContent, resp.StatusCode)
	return errors.New("Received wrong return value")
}
