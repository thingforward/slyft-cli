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

	"strings"

	"github.com/jawher/mow.cli"
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

func (p *Project) getName() string {
	return p.Name
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

	data := [][]string{
		[]string{"Key", "Value"},
		[]string{"Name", p.Name},
		[]string{"Details", p.Details},
		[]string{"CreatedAt", p.CreatedAt.String()},
		[]string{"UpdatedAt", p.UpdatedAt.String()},
		[]string{"Settings", string(p.Settings)}}

	fmt.Fprintf(os.Stdout, "%s%s",
		markdownHeading("Project Details", 1),
		markdownTable(&data))
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

	data := [][]string{
		{"Number", "Name", "Details"}}
	for i := range projects {
		p := projects[i]
		data = append(data, []string{fmt.Sprintf("%d", i+1), p.Name, p.Details})
	}

	fmt.Fprint(os.Stdout, markdownTable(&data))
}

func extractProjectsFromBody(body []byte) ([]Project, error) {
	//Log.Debugf("body=%v", string(body))
	projects := make([]Project, 0)
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func extractProjectFromBody(body []byte) ([]Project, error) {
	//Log.Debugf("body=%v", string(body))
	p := &Project{}

	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	return []Project{*p}, nil
}

func displayProjectsFromResponse(resp *http.Response, expectedCode int, listExpected bool) error {
	projects, err := extractProjectFromResponse(resp, expectedCode, listExpected)
	if err != nil {
		Log.Fatalf("Unable to list projects: %s", err)
		return err
	}
	Log.Debugf("projects=%+v", projects)

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
	cmd.Spec = "[--name] [--remember]"
	name := cmd.StringOpt("name n", "", "Name for the project")
	remember := cmd.BoolOpt("remember r", false, "Remember project name in the current directory")

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
			ReportError("Contacting the server", err)
			return
		}
		defer resp.Body.Close()
		displayProjectsFromResponse(resp, http.StatusCreated, false)

		if remember != nil && *remember {
			_, err := os.Open(".slyftproject")
			if err == nil {
				fmt.Println("--remember was chosen, but there is already a .slyftproject file. Leaving as-is")
			} else {
				fmt.Println("Remembering this project in file .slyftproject")

				slyftProjectFile, err := os.Create(".slyftproject")
				defer slyftProjectFile.Close()
				if err != nil {
					fmt.Println("--remember was chosen, but was unable to create a .slyftproject here.")
					fmt.Println("Please create manually or use --project/--name parameters")
					Log.Debug(err)
				} else {
					w := bufio.NewWriter(slyftProjectFile)
					_, err := w.WriteString(*name)
					if err != nil {
						fmt.Println("--remember was chosen, but was unable to write to .slyftproject here.")
						fmt.Println("Please check manually or use --project/--name parameters")
						Log.Debug(err)
					}
					w.Flush()
				}
			}
		}
	}
}

func settingsProject(cmd *cli.Cmd) {
	cmd.Spec = "[--name] KEY VALUE"
	name := cmd.StringOpt("name", "", "Name for the project")
	key := cmd.StringArg("KEY", "", "Name of the setting")
	value := cmd.StringArg("VALUE", "", "Value of the setting")
	cmd.Action = func() {
		if *key == "" {
			fmt.Println("KEY must not be empty.")
			return
		}

		if *name == "" {
			*name, _ = ReadProjectLock()
		}
		p, err := chooseProject(*name, "Which project needs to be updated: ")
		if err == nil {
			resp, err := Do(p.EndPoint(), "PUT", createProjectParam("", "", fmt.Sprintf(`{"%s": "%s"}`, *key, *value)))
			defer resp.Body.Close()
			if err != nil || resp.StatusCode != http.StatusNoContent {
				fmt.Printf("Something went wrong: %s\n", err)
			} else {
				fmt.Println("Successfully updated")
				resp, err := Do(p.EndPoint(), "GET", nil)
				defer resp.Body.Close()
				if err == nil {
					displayProjectsFromResponse(resp, http.StatusOK, false)
					return
				}
			}
			return
		}
		ReportError("Choosing a project", err)
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
		Log.Debugf("err=%#v", err)
		Log.Debugf("resp=%#v", resp)
		if err != nil {
			ReportError("Listing the projects", err)
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
		ReportError("Showing project", err)
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
			ReportError("Deleting project", err)
			return
		}
		DeleteApiModel(p)
	}
}

func RegisterProjectRoutes(proj *cli.Cmd) {
	SetupLogger()

	proj.Command("create c", "Create a new project", createProject)
	proj.Command("settings", "Create a new project", settingsProject)
	proj.Command("list ls", "List all projects", listProjects)
	proj.Command("show sh", "Show an existing project", showProject)
	proj.Command("delete d", "Delete an existing project", deleteProject)

	proj.Command("build b", "Build a project", buildProject)
	proj.Command("validate v", "Validate a project", validateProject)
	proj.Command("status st", "Show job status of a projct", jobStatusProject)
}
