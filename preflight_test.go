package main

import (
	"testing"
)

func TestPreflightAsset(t *testing.T) {
	jsonFileMock := "mock.json"
	yamlFileMock := "mock.yaml"
	ramlFileMock := "mock.raml"

	//byte arrays
	invalidUtf8 := []byte{0xff, 0xfe, 0xfd}
	xmlMarkup := []byte("<?xml version='1.0' encoding='UTF-8' standalone='yes'?><root/>")
	validJson := []byte("{ \"foo\": [\"bar\", \"barfoo\"] }")
	validYaml := []byte("\"foo\": \"bar\"")
	longAsset := make([]byte, maxAssetLen+1)
	validMultilineYaml := []byte(`"foo":
- "bar"
- "foobar"
`)
	invalidMultilineYaml := []byte(`"foo":
- "bar"
  /------------\
	|random DITAA|
	\------------/
- "boofar"
`)
	multilineRaml := []byte(`#%RAML 1.0
  title: My API`)

	//expect error
	err := preflightAsset(&invalidUtf8, jsonFileMock)
	if err == nil {
		t.Error("Must reject invalid UTF8 with JSON filename")
	}

	err = preflightAsset(&longAsset, jsonFileMock)
	if err == nil {
		t.Errorf("Must reject assets longer than %d bytes", maxAssetLen)
	}

	err = preflightAsset(&invalidUtf8, yamlFileMock)
	if err == nil {
		t.Error("Must reject invalid UTF8 with YAML filename")
	}

	err = preflightAsset(&validYaml, jsonFileMock)
	if err == nil {
		t.Error("Must reject YAML with JSON filename")
	}

	err = preflightAsset(&invalidMultilineYaml, jsonFileMock)
	if err == nil {
		t.Error("Must reject invalid YAML")
	}

	//much JSON is also valid YAML, so don't disallow JSON with YAML filename

	err = preflightAsset(&xmlMarkup, "")
	if err == nil {
		t.Error("Must reject XML markup")
	}

	//RAML 0.8/1.0 requires a version line up front
	//so reject any YAML file without it
	err = preflightAsset(&validMultilineYaml, ramlFileMock)
	if err == nil {
		t.Error("Must reject RAML without initial RAML version line")
	}

	//faulty YAML

	//expect success
	err = preflightAsset(&validYaml, yamlFileMock)
	if err != nil {
		t.Errorf("Must accept valid YAML: %v", err)
	}

	err = preflightAsset(&validJson, jsonFileMock)
	if err != nil {
		t.Errorf("Must accept valid JSON: %v", err)
	}

	err = preflightAsset(&validMultilineYaml, yamlFileMock)
	if err != nil {
		t.Errorf("Must accept valid multiline YAML: %v", err)
	}

	err = preflightAsset(&multilineRaml, ramlFileMock)
	if err != nil {
		t.Errorf("Must accept valid multiline RAML: %v", err)
	}
}
