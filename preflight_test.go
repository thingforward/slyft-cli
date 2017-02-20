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
	multilineYaml := []byte(`"foo":
- "bar"
- "foobar"
- "boofar"
- "roobar"
`)
	//multilineYamlConverted := []byte("{\"foo\":[\"bar\",\"foobar\",\"boofar\",\"roobar\"]}")
	multilineRaml := []byte(`#%RAML 1.0
  title: My API`)
	//multilineRamlConverted := []byte("{\"title\":\"My API\"}")

	//expect error
	mimetype, err := preflightAsset(&invalidUtf8, jsonFileMock)
	if err == nil {
		t.Error("Must reject invalid UTF8 with JSON filename")
	}
	if mimetype != "" {
		t.Error("mimetype must be empty for invalid input")
	}

	mimetype, err = preflightAsset(&longAsset, jsonFileMock)
	if err == nil {
		t.Errorf("Must reject assets longer than %d bytes", maxAssetLen)
	}
	if mimetype != "" {
		t.Error("mimetype must be empty for invalid input")
	}

	mimetype, err = preflightAsset(&invalidUtf8, yamlFileMock)
	if err == nil {
		t.Error("Must reject invalid UTF8 with YAML filename")
	}
	if mimetype != "" {
		t.Error("mimetype must be empty for invalid input")
	}

	mimetype, err = preflightAsset(&validYaml, jsonFileMock)
	if err == nil {
		t.Error("Must reject YAML with JSON filename")
	}
	if mimetype != "" {
		t.Error("mimetype must be empty for invalid input")
	}

	//much JSON is also valid YAML, so don't disallow JSON with YAML filename

	mimetype, err = preflightAsset(&xmlMarkup, "")
	if err == nil {
		t.Error("Must reject XML markup")
	}
	if mimetype != "" {
		t.Error("mimetype must be empty for invalid input")
	}

	//RAML 0.8/1.0 requires a version line up front
	//so reject any YAML file without it
	mimetype, err = preflightAsset(&multilineYaml, ramlFileMock)
	if err == nil {
		t.Error("Must reject RAML without initial RAML version line")
	}
	if mimetype != "" {
		t.Error("mimetype must be empty for invalid input")
	}

	//expect success
	mimetype, err = preflightAsset(&validYaml, yamlFileMock)
	if err != nil {
		t.Errorf("Must accept valid YAML: %v", err)
	}
	if mimetype == "" {
		t.Error("mimetype must not be empty for valid input")
	}

	mimetype, err = preflightAsset(&validJson, jsonFileMock)
	if err != nil {
		t.Errorf("Must accept valid JSON: %v", err)
	}
	if mimetype == "" {
		t.Error("mimetype must not be empty for valid input")
	}

	//in-place conversion must match predefined result
	mimetype, err = preflightAsset(&multilineYaml, yamlFileMock)
	if err != nil {
		t.Errorf("Must accept valid multiline YAML: %v", err)
	}
	if mimetype == "" {
		t.Error("mimetype must not be empty for valid input")
	}
	//if string(multilineYaml) != string(multilineYamlConverted) {
	//	t.Errorf("Expected %s to match %s", multilineYaml, multilineYamlConverted)
	//}

	//likewise for RAML input
	mimetype, err = preflightAsset(&multilineRaml, ramlFileMock)
	if err != nil {
		t.Errorf("Must accept valid multiline RAML: %v", err)
	}
	if mimetype == "" {
		t.Error("mimetype must not be empty for valid input")
	}
	//if string(multilineRaml) != string(multilineRamlConverted) {
	//	t.Errorf("Expected %s to match %s", multilineRaml, multilineRamlConverted)
	//}
}
