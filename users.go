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
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/jawher/mow.cli"
	"golang.org/x/crypto/ssh/terminal"
)

type SlyftAuth struct {
	AccessToken string `json:"access_token"`
	Client      string `json:"client"`
	Uid         string `json:"uid"`
}

type SlyftAuthResult struct {
	Success bool
	Errors  []string
}

func (sa SlyftAuth) String() string {
	bytes, err := json.Marshal(sa)
	if err != nil {
		return "Couldn't convert the Auth to string. Sorry about that"
	}
	return string(bytes)
}

func (sa SlyftAuth) GoodForLogin() bool {
	return sa.AccessToken != "" && sa.Client != "" && sa.Uid != ""
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
	Email                string          `json:"email"`
	Password             string          `json:"password"`
	PasswordConfirmation string          `json:"password_confirmation"`
	TermsAcceptance      TermsAcceptance `json:"terms"`
}

type Terms struct {
	Url       string `json:"url"`
	StartedAt string `json:"started_at"`
}

type TermsAcceptance struct {
	Accepted  bool   `json:"accepted"`
	Timestamp string `json:"timestamp"`
}

type PwdResetRequest struct {
	Code                 string `json:"code,omitempty"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

func readSecret(ask string) string {
	pwd_from_env := os.Getenv("SLYFT_USER_REGISTRATION_PWD")
	if len(pwd_from_env) == 0 {
		fmt.Print(ask)
		byteSecret, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println("")
		if err != nil {
			Log.Critical("Reading secrect failed: " + err.Error())
		}
		return string(byteSecret)
	} else {
		fmt.Print(ask)
		fmt.Print(" <<SUPLIED BY ENV VARIABLE>>")
		fmt.Println("")
		return pwd_from_env
	}
}

func termsUri() (string, error) {
	// get T&C JSON from endpoint to get the URL to the latest terms document
	resp, err := DoNoAuth("/terms", "GET", nil)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	t := &Terms{}
	if err := json.Unmarshal(body, &t); err != nil {
		return "", err
	}
	return t.Url, nil
}

func getTerms() (string, error) {
	// get the terms content as string from the referenced terms document
	termsUri, err := termsUri()
	if err != nil {
		return "", err
	}
	response, err := http.Get(termsUri)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	responseString := string(responseData)
	return responseString, nil
}

func displayTermsAndConditions() error {
	// display terms file contents
	terms, err := getTerms()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Print("\n-------------------------------------------------\n")
	lines := strings.Split(terms, "\n")
	term_height := TerminalHeight() - 2
	reader := bufio.NewReader(os.Stdin)
	for idx, line := range lines {
		fmt.Print(line)
		fmt.Print("\n")

		if (idx+1)%term_height == 0 {
			fmt.Print("-- Hit [ENTER] for next page >")
			reader.ReadString('\n')
		}
	}

	fmt.Print("-------------------------------------------------\n")
	return nil
}

func acceptTermsAndConditions() (bool, error) {
	/*
		- ask user for acceptance
		- return boolean true/false based on user input
	*/
	err := displayTermsAndConditions()
	if err != nil {
		return false, nil
	}
	accept := askForConfirmation("Do you accept the Terms and Conditions?")
	if accept == false {
		return false, nil
	}
	return accept, nil
}

func readCredentials(confirm bool) (string, string, string) {
	reader := bufio.NewReader(os.Stdin)

	var email, password string
	proceed := false
	for proceed == false {
		fmt.Print("Enter Email: ")
		email, _ = reader.ReadString('\n')
		proceed = validateEmail(strings.TrimSpace(email))
		if proceed == false {
			fmt.Println("Not a valid email address. Please try again.")
		}
	}

	proceed = false
	for proceed == false {
		password = readSecret("Enter Password (min. 6 characters): ")
		proceed = validatePassword(strings.TrimSpace(password))
		if proceed == false {
			fmt.Println("Not a valid password. Please try again.")
		}
	}

	if !confirm {
		return strings.TrimSpace(email), strings.TrimSpace(password), ""
	}

	passwordConfirmation := readSecret("Please confirm Password: ")
	return strings.TrimSpace(email), strings.TrimSpace(password), strings.TrimSpace(passwordConfirmation)
}

func validatePassword(s string) bool {
	return len(s) >= 6
}

func validateEmail(s string) bool {
	// regexp for email doesn't work, so check the bare minimum; adapted from:
	// https://davidcel.is/posts/stop-validating-email-addresses-with-regex/
	re := regexp.MustCompile(`^.+@.+\..+$`)
	return re.FindStringIndex(s) != nil
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
	// if the user wants to register, show T&C to the user, and ask for acceptance
	if register {
		fmt.Print("\nFor a successful registration, we kindly ask you to read and accept our\n")
		fmt.Print("Terms and Conditions. Please press [ENTER] to view and accept. >")
		reader := bufio.NewReader(os.Stdin)
		_, err := reader.ReadString('\n')

		accept, err := acceptTermsAndConditions()
		if !accept {
			return errors.New(fmt.Sprintf("You need to accept the terms first. %v\n", err))
		}
		creds.TermsAcceptance.Accepted = accept
		creds.TermsAcceptance.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05-0700")
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(creds)

	Log.Debugf("url=%#v", url)

	resp, err := http.Post(url, "application/json; charset=utf-8", b)

	Log.Debugf("err=%#v", err)
	Log.Debugf("resp=%#v", resp)

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
		if register {
			fmt.Print("\nWe're sorry, but your registration failed due to the following errors:\n")
			var f interface{}
			err := json.Unmarshal(body, &f)
			if err == nil {
				m := f.(map[string]interface{})
				e := (m["errors"]).(map[string]interface{})
				fm := (e["full_messages"]).([]interface{})
				for _, msg := range fm {
					fmt.Printf("* %s\n", msg)
				}

			}
		} else {
			fmt.Print("\nWe're sorry, but your login failed due to the following errors:\n")
			var sar SlyftAuthResult
			err := json.Unmarshal(body, &sar)
			if err == nil {
				for _, msg := range sar.Errors {
					fmt.Printf("* %s\n", msg)
				}
			} else {
				// Unable to parse it, log as-is
				Log.Critical(string(body))
			}
		}
		return errors.New(fmt.Sprintf("Server returned failure: %v\nBye", resp.Status))
	} else {
		Log.Critical("Error reading/parsing API output.")
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
		fmt.Println("We're very sorry, but your registration failed.")
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
		fmt.Println("Sorry, login failed")
	} else {
		fmt.Println("Login successful, have fun! For documentation, please have a look at www.slyft.io/docs")
	}
}

func makeDeleteCall(endpoint string) error {
	resp, err := Do(endpoint, "DELETE", nil)
	deactivateLogin()

	Log.Debugf("err=%#v", err)
	Log.Debugf("resp=%#v", resp)

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
	Log.Debugf("err=%#v", err)
	Log.Debugf("body=%#v", body)
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
	auth, err := readAuthFromConfig()
	if err != nil {
		fmt.Println("You do not seem to be logged in. Please do a `slyft user login`")
		return
	}
	if !auth.GoodForLogin() {
		fmt.Println("You do not seem to be logged in. Please do a `slyft user login`")
		return
	}

	fmt.Println("You may choose to delete your Slyft account at any time. Please be aware")
	fmt.Println("that all previously processed data under your account will be deleted.")
	fmt.Println()

	confirm := askForConfirmation("Are you sure to delete your user account?")
	if confirm {
		err := makeDeleteCall("/auth")
		if err != nil {
			Log.Error("Sorry, deletion failed")
		} else {
			fmt.Println("Deleted the account. We are sorry to see you go. Come back soon...")
		}
	} else {
		fmt.Println("Account left unchanged.")
	}
}

func ForgotPassword() {
	var email string
	type forgotPwd struct {
		Email       string `json:"email"`
		RedirectUrl string `json:"redirect_url"`
	}
	auth, err := readAuthFromConfig()
	if err != nil || !auth.GoodForLogin() {
		fmt.Print("Please provide the email address you have used to register: ")
		email, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		email = strings.TrimSpace(email)
		if !validateEmail(email) {
			fmt.Println("Not a valid email address. Please try again.")
			return
		}
	} else {
		email = auth.Uid
		fmt.Printf("You are registered with the email %s\n", email)
	}
	fmt.Printf("We will send an email to %s containing a reset code.\nThen use the change password function to reset your password using the token. ", email)
	cont := askForConfirmation("continue?")
	if !cont {
		return
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(&forgotPwd{Email: email, RedirectUrl: ""})

	resp, err := http.Post(ServerURL("/auth/password"), "application/json; charset=utf-8", b)
	Log.Debugf("resp=%#v", resp)
	if err != nil || resp.StatusCode != http.StatusOK {
		Log.Debugf("err=%#v", err)
		fmt.Println("Sorry, password could not be reset. Please try again")
		return
	}
	fmt.Println("Check your inbox for an email with the token, then use the change password fucntion")
}

func ChangePassword() {
	var email string
	auth, err := readAuthFromConfig()
	var changePasswordForLoggedInUser bool
	if err != nil || !auth.GoodForLogin() {
		fmt.Print("Please provide the email address you have used to register: ")
		email, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		email = strings.TrimSpace(email)
		if !validateEmail(email) {
			fmt.Println("Not a valid email address. Please try again.")
			return
		}
	} else {
		email = auth.Uid
		fmt.Printf("You are registered with the email %s\n", email)
		changePasswordForLoggedInUser = true
	}

	resetRequest, err := askUserForNewPasswordAndConfirmation()
	if err != nil {
		fmt.Printf("invalid input: %s\n", err)
		return
	}

	resetWithToken := askForConfirmation("If you have a reset token, answer (y)")
	if resetWithToken {
		fmt.Print("Please provide your password reset token: ")
		token, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		token = strings.TrimSpace(token)
		resetRequest.Code = token
		err = updatePasswordWithResetToken(email, resetRequest)
	} else {
		err = updatePasswordForAuthenticatedUser(email, resetRequest)
	}
	if err != nil {
		fmt.Printf("Could not reset password: %s\n", err)
		return
	}

	fmt.Println("The password has been reset for account: ", email)
	if changePasswordForLoggedInUser {
		deactivateLogin() // We leave it to the server whether to clean up the tokens.
	}
	fmt.Println("Please login with your new credentials.")
}

func askUserForNewPasswordAndConfirmation() (*PwdResetRequest, error) {
	password := readSecret("Please provide your new password (min. 6 characters): ")
	if !validatePassword(password) {
		err := fmt.Errorf("Not a valid password. Please try again.")
		return nil, err
	}

	passwordConfirmation := readSecret("Please repeat your new password: ")

	if password != passwordConfirmation {
		err := fmt.Errorf("Passwords do not match. Please try again.")
		return nil, err
	}
	return &PwdResetRequest{Password: password, PasswordConfirmation: passwordConfirmation}, nil
}

func updatePasswordWithResetToken(email string, resetRequest *PwdResetRequest) error {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(resetRequest)

	resp, err := http.Post(ServerURL("/password_reset"), "application/json; charset=utf-8", b)
	Log.Debugf("resp=%#v", resp)
	if err != nil || resp.StatusCode != http.StatusOK {
		Log.Debugf("err=%#v", err)
		return fmt.Errorf("Sorry, password could not be reset. Please try again")
	}
	return nil
}

func updatePasswordForAuthenticatedUser(email string, resetRequest *PwdResetRequest) error {
	password := readSecret("Please provide your current password: ")
	password = strings.TrimSpace(password)

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(&Credentials{Email: email, Password: password})

	resp, err := http.Post(ServerURL("/auth/sign_in"), "application/json; charset=utf-8", b)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Sorry, the credentials are not correct. Please try again")
	}

	newAuth := extractAuthFromHeader(&resp.Header)

	resp, err = DoAuth("/auth/password", "PUT", resetRequest, &newAuth)
	Log.Debugf("resp=%#v", resp)

	if err != nil || (resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent) {
		Log.Debugf("err=%#v", err)
		Log.Debugf("err=%#v", resp)
		return fmt.Errorf("Sorry, changing password failed.")
	}
	return nil
}

func RegisterUserRoutes(user *cli.Cmd) {
	SetupLogger()

	user.Command("register r", "Register yourself", func(cmd *cli.Cmd) { cmd.Action = RegisterUser })
	user.Command("login l", "Login with your credentials", func(cmd *cli.Cmd) { cmd.Action = LogUserIn })
	user.Command("logout", "Log out from your session", func(cmd *cli.Cmd) { cmd.Action = LogUserOut })
	user.Command("delete", "Delete your account", func(cmd *cli.Cmd) { cmd.Action = DeleteUser })
	user.Command("change-password cp", "Change your password", func(cmd *cli.Cmd) { cmd.Action = ChangePassword })
	user.Command("forgot-password fp", "Request password reset token, forgot password function", func(cmd *cli.Cmd) { cmd.Action = ForgotPassword })
}
