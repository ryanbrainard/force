package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var cmdLogin = &Command{
	Run:   runLogin,
	Usage: "login",
	Short: "Log in to force.com",
	Long: `
Log in to force.com

Examples:

  force login
`,
}

func runLogin(cmd *Command, args []string) {
	creds, err := resolveForceCredentials(args)
	if err != nil {
		ErrorAndExit(err.Error())
	}

	_, err = SaveForceCredentials(creds)
	if err != nil {
		ErrorAndExit(err.Error())
	}
}

func resolveForceCredentials(args []string) (creds ForceCredentials, err error) {
	var endpoint ForceEndpoint
	switch len(args) {
	default:
		err = errors.New("invaild number of arguments")
	case 2:
		creds.InstanceUrl = args[0]
		if !strings.HasPrefix(creds.InstanceUrl, "http") {
			creds.InstanceUrl = "https://" + creds.InstanceUrl
		}
		creds.AccessToken = args[1]
		if !strings.HasPrefix(creds.AccessToken, "00D") {
			err = errors.New("invaild access token")
		}
		force := NewForce(creds)
		var services map[string]string
		services, err = force.GetServices()
		if err != nil {
			return
		}
		creds.Id = services["identity"]
	case 1:
		switch args[0] {
		case "test":
			endpoint = EndpointTest
		case "pre":
			endpoint = EndpointPrerelease
		default:
			err = errors.New(fmt.Sprintf("no such endpoint: %s", args[0]))
		}
		fallthrough
	case 0:
		creds, err = ForceLogin(endpoint)
	}
	return
}

var cmdLogout = &Command{
	Run:   runLogout,
	Usage: "logout <account>",
	Short: "Log out from force.com",
	Long: `
Log out from force.com

Examples:

  force logout user@example.org
`,
}

func runLogout(cmd *Command, args []string) {
	if len(args) != 1 {
		ErrorAndExit("must specify account to log out")
	}
	account := args[0]
	Config.Delete("accounts", account)
	if active, _ := Config.Load("current", "account"); active == account {
		Config.Delete("current", "account")
		SetActiveAccountDefault()
	}
}

func ForceLoginAndSave(endpoint ForceEndpoint) (username string, err error) {
	creds, err := ForceLogin(endpoint)
	if err != nil {
		return
	}
	return SaveForceCredentials(creds)
}

func SaveForceCredentials(creds ForceCredentials) (username string, err error) {
	force := NewForce(creds)
	login, err := force.Get(creds.Id)
	if err != nil {
		return
	}
	body, err := json.Marshal(creds)
	if err != nil {
		return
	}
	username = login["username"].(string)
	Config.Save("accounts", username, string(body))
	Config.Save("current", "account", username)
	return
}
