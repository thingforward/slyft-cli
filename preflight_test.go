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
	multilineYamlConverted := []byte("{\"foo\":[\"bar\",\"foobar\",\"boofar\",\"roobar\"]}")
	multilineRaml := []byte(`#%RAML 1.0
  title: My API`)
	multilineRamlAfterPreflight := multilineRaml

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

	//much JSON is also valid YAML, so don't disallow JSON with YAML filename

	err = preflightAsset(&xmlMarkup, "")
	if err == nil {
		t.Error("Must reject XML markup")
	}

	//RAML 0.8/1.0 requires a version line up front
	//so reject any YAML file without it
	err = preflightAsset(&multilineYaml, ramlFileMock)
	if err == nil {
		t.Error("Must reject RAML without initial RAML version line")
	}

	//expect success
	err = preflightAsset(&validYaml, yamlFileMock)
	if err != nil {
		t.Errorf("Must accept valid YAML: %v", err)
	}

	err = preflightAsset(&validJson, jsonFileMock)
	if err != nil {
		t.Errorf("Must accept valid JSON: %v", err)
	}

	//in-place conversion must match predefined result
	err = preflightAsset(&multilineYaml, yamlFileMock)
	if err != nil {
		t.Errorf("Must accept valid multiline YAML: %v", err)
	}
	if string(multilineYaml) != string(multilineYamlConverted) {
		t.Errorf("Expected %s to match %s", multilineYaml, multilineYamlConverted)
	}

	//likewise for RAML input, only here the buffer mustn't change
	//as RAML isn't JSON-compatible (the initial comment doesn't translate for a start)
	err = preflightAsset(&multilineRaml, ramlFileMock)
	if err != nil {
		t.Errorf("Must accept valid multiline RAML: %v", err)
	}
	if string(multilineRaml) != string(multilineRamlAfterPreflight) {
		t.Errorf("Expected %s to match %s", multilineRaml, multilineRamlAfterPreflight)
	}
}
