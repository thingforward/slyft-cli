package main

import (
	"testing"
)

func TestMarkdownHeading(t *testing.T) {
	expectedH1 := `Foobar
======

`
	expectedH3 := `### Foobar

`

	h1 := markdownHeading("Foobar", 1)
	if h1 != expectedH1 {
		t.Errorf("Expected:\n%s\nGot:\n%s\n", expectedH1, h1)
	}

	h3 := markdownHeading("Foobar", 3)
	if h3 != expectedH3 {
		t.Errorf("Expected:\n%s\nGot:\n%s\n", expectedH3, h3)
	}
}

func TestMarkdownTable(t *testing.T) {
	rows := make([][]string, 2)
	rows[0] = []string{"a", "B"}
	rows[1] = []string{"C", "d"}

	//add padding
	expected := "| A     | B     | \n|:------|:------|\n| C     | d     | \n\n"

	md := markdownTable(&rows)

	if md != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s\n", expected, md)
	}
}
