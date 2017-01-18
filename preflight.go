package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"regexp"
	"unicode/utf8"
)

const maxAssetLen int = 20000

func preflightAsset(a *[]byte, file string) error {
	if len(*a) == 0 {
		return errors.New("input must not be empty")
	}

	if len(*a) > maxAssetLen {
		return errors.New(fmt.Sprintf("input length must not exceed %d", maxAssetLen))
	}

	if utf8.Valid(*a) == false {
		return errors.New("input must be valid UTF-8")
	}

	//if extension indicates YAML, attempt conversion
	//(otherwise assume JSON)
	reYaml := regexp.MustCompile("(?i)\\.ya?ml$") //'a' is optional
	isYaml := reYaml.FindStringIndex(file) != nil
	reRaml := regexp.MustCompile("(?i)\\.raml$") //'a' is obligatory
	isRaml := reRaml.FindStringIndex(file) != nil

	if isRaml {
		if bytes.HasPrefix(*a, []byte("#%RAML")) == false {
			return errors.New(fmt.Sprint("invalid RAML: expected RAML comment line"))
		}
	}

	if isYaml || isRaml {
		json, err := yaml.YAMLToJSON(*a)
		if err != nil {
			return errors.New(fmt.Sprintf("invalid YAML: %v", err))
		}
		*a = json
	}

	//now parse the JSON
	var any interface{}
	err := json.Unmarshal(*a, &any)
	if err != nil {
		return errors.New(fmt.Sprintf("invalid JSON: %v", err))
	}

	return nil
}
