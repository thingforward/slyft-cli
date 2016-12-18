package main

import (
	"testing"
	"time"
)

func TestCreateProjectParam(t *testing.T) {
	name := "TestName"
	details := "TestDetails"
	settings := "TestSettings"
	param := createProjectParam(name, details, settings)

	now := time.Now()
	if param.Project.Name != name ||
		param.Project.Details != details ||
		param.Project.Settings != settings ||
		param.Project.CreatedAt.After(now) ||
		param.Project.UpdatedAt.After(now) {
		t.Errorf("Broken project parameter: %v", param)
	}
}
