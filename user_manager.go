package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/urfave/cli"
)

type SlyftAuth struct {
	AccessToken string `json:"access_token"`
	Client      string `json:"client"`
	Uid         string `json:"uid"`
}

func (sa SlyftAuth) String() string {
	bytes, err := json.Marshal(sa)
	if err != nil {
		return "Couldn't convert the Auth to string. Sorry about that"
	}
	return string(bytes)
}

type SlyftRC struct {
	Auth     SlyftAuth
	More     string
	Settings string
	To       string
	Come     string
}

func (sr SlyftRC) String() string {
	bytes, err := json.Marshal(sr)
	if err != nil {
		return "Couldn't convert the config to string. Sorry about that"
	}
	return string(bytes)
}

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

func extractAuthFromHeader(hdr *http.Header) SlyftAuth {
	return SlyftAuth{
		AccessToken: hdr.Get("access-token"),
		Client:      hdr.Get("client"),
		Uid:         hdr.Get("uid"),
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

	// Verify if the resp was ok
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		slyftAuth := extractAuthFromHeader(&resp.Header)
		return writeAuthToConfig(&slyftAuth)
	}

	// handle the error
	body, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		Log.Critical(string(body)) // TODO -- parse and print it beautifully. Extract errors/full_messages/etc.
		return errors.New(fmt.Sprintf("Server returned failure: %v\nBye", resp.Status))
	}

	return errors.New(fmt.Sprintf("Reading server resonse failed: %v\n", err))
}

func RegisterUser(c *cli.Context) error {
	err := authenticateUser("/auth", true)
	if err != nil {
		Log.Error("Sorry, registration failed")
	} else {
		fmt.Println("Successfully registered, have fun...")
	}
	return err
}

func LogUserIn(c *cli.Context) error {
	err := authenticateUser("/auth/sign_in", false)
	if err != nil {
		Log.Error("Sorry, login failed")
	} else {
		fmt.Println("Login successful, have fun...")
	}
	return err
}

func addAuthToHeader(hdr *http.Header, s *SlyftAuth) {
	hdr.Add("access-token", s.AccessToken)
	hdr.Add("client", s.Client)
	hdr.Add("uid", s.Uid)
}

func makeDeleteCall(endpoint string) error {
	url := ServerURL(endpoint)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		Log.Critical("Failed to create a request: " + err.Error())
		return err
	}

	auth, err := readAuthFromConfig()
	deactivateLogin()
	if err != nil {
		return err
	}

	Log.Errorf("%s", auth)

	addAuthToHeader(&req.Header, auth)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Verify if the resp was ok
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		slyftAuth := extractAuthFromHeader(&resp.Header)
		return writeAuthToConfig(&slyftAuth)
	}

	// handle the error
	body, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		Log.Critical(string(body)) // TODO -- parse and print it beautifully. Extract errors/full_messages/etc.
		return errors.New(fmt.Sprintf("Server returned failure: %v\nBye", resp.Status))
	}

	return errors.New(fmt.Sprintf("Reading server resonse failed: %v\n", err))
}

func LogUserOut(c *cli.Context) error {
	err := makeDeleteCall("/auth/sign_out")
	if err != nil {
		Log.Error("Sorry, logout failed.")
	} else {
		fmt.Println("Bye for now. Looking forward to seeing you soon...")
	}
	return err
}

func DeleteUser(c *cli.Context) error {
	err := makeDeleteCall("/auth")
	if err != nil {
		Log.Error("Sorry, deletion failed")
	} else {
		fmt.Println("Deleted the account. We are sorry to see you go. Come back soon...")
	}
	return err
}
