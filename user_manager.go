package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/urfave/cli"
)

type Credentials struct {
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

func readSecret(ask string) string {
	fmt.Print(ask)
	byteSecret, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	if err != nil {
		Log.Critical("Reading secrect failed: " + err.Error())
	}
	return string(byteSecret)
}

func readCredentials(confirm bool) (string, string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Email: ")
	username, _ := reader.ReadString('\n')

	password := readSecret("Enter Password: ")
	if !confirm {
		return strings.TrimSpace(username), strings.TrimSpace(password), ""
	}

	passwordConfirmation := readSecret("Confirm Password: ")
	return strings.TrimSpace(username), strings.TrimSpace(password), strings.TrimSpace(passwordConfirmation)
}

func getCredentials(confirm bool) *Credentials {
	email, password, confirmation := readCredentials(confirm)
	return &Credentials{
		Email:                email,
		Password:             password,
		PasswordConfirmation: confirmation,
	}
}

func authenticateUser(endpoint string, register bool) error {
	url := ServerURL(endpoint)
	creds := getCredentials(register)
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(creds)
	resp, err := http.Post(url, "application/json; charset=utf-8", b)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		Log.Fatal(err)
	}

	return nil
}

func RegisterUser(c *cli.Context) error {
	return authenticateUser("/auth", true)
}

func LogUserIn(c *cli.Context) error {
	return authenticateUser("/auth/sign_in", false)
}

func LogUserOut(c *cli.Context) error {
	fmt.Println("TODO TODO TODO")
	return nil
}

func DeleteUser(c *cli.Context) error {
	fmt.Println("TODO TODO TODO")
	return nil
}
