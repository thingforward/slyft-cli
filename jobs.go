package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	cli "github.com/jawher/mow.cli"
	"github.com/olekukonko/tablewriter"
)

type Job struct {
	ID          int        `json:"id"`
	Kind        string     `json:"kind"`
	Status      string     `json:"status"`
	Results     JobResults `json:"results"`
	ProjectId   int        `json:"project_id"`
	ProjectName string     `json:"project_name"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type JobResults struct {
	ResultMessage string   `json:"resultMessage"`
	ResultStatus  int      `json:"resultStatus"`
	ResultAssets  []string `json:"resultAssets"`
}

func (j *Job) Display() { // String?
	if j == nil {
		return
	}

	var data [][]string
	data = append(data, []string{"Id", fmt.Sprintf("%d", j.ID)})
	data = append(data, []string{"Kind", j.Kind})
	data = append(data, []string{"Status", j.Status})
	data = append(data, []string{"ProjectId", fmt.Sprintf("%d", j.ProjectId)})
	data = append(data, []string{"ProjectName", j.ProjectName})
	data = append(data, []string{"CreatedAt", j.CreatedAt.String()})
	data = append(data, []string{"UpdatedAt", j.UpdatedAt.String()})
	data = append(data, []string{"ResultMessage", j.Results.ResultMessage})
	data = append(data, []string{"ResultStatus", fmt.Sprintf("%d", j.Results.ResultStatus)})
	for index, asset := range j.Results.ResultAssets {
		data = append(data, []string{fmt.Sprintf("ResultAssets[%d]", index), asset})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(TerminalWidth())
	table.SetHeader([]string{"Key", "Value"})
	table.SetBorder(false)
	table.AppendBulk(data)
	fmt.Fprintf(os.Stdout, "\n---- Job Details ----\n")
	table.Render()
	fmt.Fprintf(os.Stdout, "\n")
}

func DisplayJobs(jobs []Job) {
	if len(jobs) == 0 {
		fmt.Println("No jobs found")
		return
	}

	if len(jobs) == 1 {
		jobs[0].Display()
		return
	}

	var data [][]string
	for i := range jobs {
		j := jobs[i]
		data = append(data, []string{fmt.Sprintf("%d", i+1), fmt.Sprintf("%d", j.ID), j.Kind, j.Status, j.ProjectName})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(TerminalWidth())
	table.SetHeader([]string{"Number", "ID", "Kind", "Status", "Project Name"})
	table.SetBorder(false)
	table.AppendBulk(data)
	fmt.Fprintf(os.Stdout, "\n")
	table.Render()
	fmt.Fprintf(os.Stdout, "\n")
}

func extractJobsFromBody(body []byte) ([]Job, error) {
	jobs := make([]Job, 0)
	if err := json.Unmarshal(body, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

func extractJobFromBody(body []byte) ([]Job, error) {
	a := &Job{}
	if err := json.Unmarshal(body, &a); err != nil {
		return nil, err
	}
	return []Job{*a}, nil
}

func chooseJob(endpoint string, askUser bool, message string) (*Job, error) {
	resp, err := Do(endpoint, "GET", nil)
	if err != nil {
		Log.Error(err)
		return nil, err
	}
	defer resp.Body.Close()
	jobs, err := extractJobFromResponse(resp, http.StatusOK, true)

	if err != nil {
		Log.Error(err)
		return nil, err
	}

	if len(jobs) == 0 {
		return nil, errors.New("No job. Sorry")
	}

	DisplayJobs(jobs)

	if len(jobs) == 1 {
		return &jobs[0], nil
	}

	if !askUser {
		return nil, nil
	}

	choice, err := ReadUserIntInput(message)
	if err != nil {
		return nil, err
	}

	if choice > len(jobs) {
		return nil, errors.New("Plese choose a number from the first column")
	}

	return &jobs[choice-1], nil
}

type JobParam struct {
	Job Job `json:"job"`
}

func creatJobParam(kind string, p *Project) *JobParam {
	return &JobParam{
		Job{
			Kind:      kind,
			ProjectId: p.ID,
		},
	}
}

func postNewJOB(kind, name string) {
	p, err := chooseProject(name, fmt.Sprintf("%s project: ", kind))
	if err != nil {
		Log.Error(err)
		return
	}

	resp, err := Do(p.JobsUrl(), "POST", creatJobParam(kind, p))
	if err != nil {
		Log.Error(err)
		return
	}

	defer resp.Body.Close()
	jobs, err := extractJobFromResponse(resp, http.StatusCreated, false)
	if err != nil {
		Log.Error(err)
		return
	}

	if len(jobs) == 1 {
		jobs[0].Display()
	}
}

func jobStatusProject(cmd *cli.Cmd) {
	cmd.Spec = "[--project] | [--all]"
	name := cmd.StringOpt("project p", "", "Name (or part of it) of a project")
	all := cmd.BoolOpt("all a", false, "Fetch details of all your jobs (do not combine with -p)")

	cmd.Action = func() {
		*name = strings.TrimSpace(*name)
		if *all || *name == "" {
			chooseJob("/v1/jobs", false, "")
			return
		}

		// first get the project, then get the pid, and make the call.
		p, err := chooseProject(*name, "Which project's jobs would you like to see: ")
		if err != nil {
			Log.Error(err)
			return
		}
		job, err := chooseJob(p.JobsUrl(), true, "Select a job id to show more details: ")
		if err != nil {
			Log.Error(err)
			return
		}
		job.Display()
	}
}

func buildProject(cmd *cli.Cmd) {
	cmd.Spec = "--project"
	name := cmd.StringOpt("project p", "", "Name (or part of it) of a project")

	cmd.Action = func() {
		postNewJOB("build", strings.TrimSpace(*name))
	}
}

func validateProject(cmd *cli.Cmd) {
	cmd.Spec = "--project"
	name := cmd.StringOpt("project p", "", "Name (or part of it) of a project")

	cmd.Action = func() {
		postNewJOB("validate", strings.TrimSpace(*name))
	}
}

func (job *Job) EndPoint() string {
	return fmt.Sprintf("/v1/projects/%d/jobs/%d", job.ProjectId, job.ID)
}
