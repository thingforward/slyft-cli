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

	"github.com/jawher/mow.cli"
	"golang.org/x/crypto/ssh/terminal"
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
	Auth SlyftAuth
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

	password := readSecret("Enter Password (min. 6 characters): ")
	if !confirm {
		return strings.TrimSpace(username), strings.TrimSpace(password), ""
	}

	passwordConfirmation := readSecret("Please confirm Password: ")
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

func RegisterUser() {
	fmt.Println("\nThank you for your interest in Slyft! Please provide us your email address and")
	fmt.Println("a password (min. 6 characters). Please make sure you have access to the email account given")
	fmt.Println("as we will send you a confirmation email to this address.")
	fmt.Println()
	err := authenticateUser("/auth", true)
	if err != nil {
		Log.Error("We're very sorry, but your registration failed. You may contact us at `info@slyft.io` for further details.")
	} else {
		fmt.Println("\nRegistration successful. We've sent you a confirmation email to the email address")
		fmt.Println("you given for this registration process. Please have a look at your inbox for")
		fmt.Println("a new message from `info@slyft.io` and follow the instructions presented there")
		fmt.Println("to activate your account.")
		fmt.Println()
	}
}

func LogUserIn() {
	err := authenticateUser("/auth/sign_in", false)
	if err != nil {
		Log.Error("Sorry, login failed")
	} else {
		fmt.Println("Login successful, have fun...")
	}
}

func makeDeleteCall(endpoint string) error {
	resp, err := Do(endpoint, "DELETE", nil)
	deactivateLogin()
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

func LogUserOut() {
	err := makeDeleteCall("/auth/sign_out")
	if err != nil {
		Log.Error("Sorry, logout failed.")
	} else {
		fmt.Println("Bye for now. Looking forward to seeing you soon...")
	}
}

func DeleteUser() {
	err := makeDeleteCall("/auth")
	if err != nil {
		Log.Error("Sorry, deletion failed")
	} else {
		fmt.Println("Deleted the account. We are sorry to see you go. Come back soon...")
	}
}

func RegisterUserRoutes(user *cli.Cmd) {
	user.Command("register r", "Register yourself", func(cmd *cli.Cmd) { cmd.Action = RegisterUser })
	user.Command("login l", "Login with your credentials", func(cmd *cli.Cmd) { cmd.Action = LogUserIn })
	user.Command("logout signout", "Log out from your session", func(cmd *cli.Cmd) { cmd.Action = LogUserOut })
	user.Command("destroy", "Delete your account", func(cmd *cli.Cmd) { cmd.Action = DeleteUser })
}
