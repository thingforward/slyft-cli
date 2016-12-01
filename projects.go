package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/urfave/cli"
)

// GET    /v1/projects(.:format)            v1/projects#index
// POST   /v1/projects(.:format)            v1/projects#create
// GET    /v1/projects/new(.:format)        v1/projects#new
// GET    /v1/projects/:id/edit(.:format)   v1/projects#edit
// GET    /v1/projects/:id(.:format)        v1/projects#show
// PATCH  /v1/projects/:id(.:format)        v1/projects#update
// PUT    /v1/projects/:id(.:format)        v1/projects#update
// DELETE /v1/projects/:id(.:format)        v1/projects#destroy

type Project struct {
	Name    string `json:"name"`
	Details string `json:"details"`
}

type ProjectParam struct {
	Project Project `json:"project"`
}

func createProjectParam(name, details string) *ProjectParam {
	return &ProjectParam{
		Project{
			name,
			details,
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

func handleProjectResponse(resp *http.Response, expectedCode int, listExpected bool) error {
	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == expectedCode {
		Log.Critical(string(body)) // TODO -- parse and print it beautifully. Extract errors/full_messages/etc.
	}

	Log.Error(resp.StatusCode)

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

	return handleProjectResponse(resp, http.StatusCreated, true)
}

func ProjectListHandler(c *cli.Context) error     { return nil }
func ProjectShowHandler(c *cli.Context) error     { return nil }
func ProjectDeletionHandler(c *cli.Context) error { return nil }
