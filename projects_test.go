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

	if param.Project.Name != name ||
		param.Project.Details != details ||
		param.Project.Settings != settings ||
		param.Project.CreatedAt.After(time.Now()) ||
		param.Project.UpdatedAt.After(time.Now()) {
		t.Errorf("Broken project parameter: %v", param)
	}
}
