package main

import (
	"testing"
)

func TestValidatePassword(t *testing.T) {
	valid := "123456"
	invalid := "1234"

	if validatePassword(valid) == false {
		t.Errorf("Must accept valid password %s", valid)
	}

	if validatePassword(invalid) == true {
		t.Errorf("Must reject invalid password %s", invalid)
	}
}

func TestValidateEmail(t *testing.T) {
	valid := "foo@bar.boo"
	invalid1 := "foobar.boo"
	invalid2 := "foobar@boo"

	if validateEmail(valid) == false {
		t.Errorf("Must accept valid email %s", valid)
	}

	if validateEmail(invalid1) == true {
		t.Errorf("Must reject invalid email %s", invalid1)
	}

	if validateEmail(invalid2) == true {
		t.Errorf("Must reject invalid email %s", invalid2)
	}
}
