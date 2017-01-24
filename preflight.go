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

	//UTF-16 or UTF-32?
	utfEndiannessSet := bytes.HasPrefix(*a, []byte{0xff, 0xfe}) || //UTF-16LE
		bytes.HasPrefix(*a, []byte{0xfe, 0xff}) || //UTF-16BE
		bytes.HasPrefix(*a, []byte{0x00, 0x00, 0xfe, 0xff}) || //UTF-32LE
		bytes.HasPrefix(*a, []byte{0x00, 0x00, 0xff, 0xfe}) //UTF-32BE

	//for UTF-16 or UTF-32, let JSON/YAML parsers test validity
	//otherwise ensure UTF-8 validity
	if utfEndiannessSet == false && utf8.Valid(*a) == false {
		return errors.New("invalid UTF-8")
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
		jsonbytes, err := yaml.YAMLToJSON(*a)
		if err != nil {
			return errors.New(fmt.Sprintf("invalid YAML: %v", err))
		}
		var anyjson interface{}
		err = json.Unmarshal(jsonbytes, &anyjson)
		if err != nil {
			return errors.New(fmt.Sprintf("invalid YAML(2): %v", err))
		}
		return nil
	}

	//now parse the JSON
	var any interface{}
	err := json.Unmarshal(*a, &any)
	if err != nil {
		return errors.New(fmt.Sprintf("invalid JSON: %v", err))
	}

	return nil
}
