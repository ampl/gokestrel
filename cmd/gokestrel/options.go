package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

func getEnv(alternatives ...string) string {
	for _, env := range alternatives {
		if value, ok := os.LookupEnv(env); ok {
			return value
		}
	}
	return ""
}

func getEnvOption(option string) string {
	return getEnv(option, strings.ToLower(option), strings.ToUpper(option))
}

func getOptions() string {
	return getEnvOption("kestrel_options")
}

var jobNumberRgx = regexp.MustCompile(`job\s*=\s*(\d+)`)
var jobPasswordRgx = regexp.MustCompile(`password\s*=\s*(\S+)`)

func getJobAndPassword() (int, string) {
	/*
		If kestrel_options is set to job/password, then return the job and password values
	*/
	jobNumber := 0
	password := ""
	options := getOptions()
	if options != "" {
		if match := jobNumberRgx.FindStringSubmatch(options); len(match) == 2 {
			if v, err := strconv.ParseInt(match[1], 10, 32); err == nil {
				jobNumber = int(v)
			}
		}
		if match := jobPasswordRgx.FindStringSubmatch(options); len(match) == 2 {
			password = match[1]
		}
	}
	return jobNumber, password
}

var priorityRgx = regexp.MustCompile(`priority\s*=\s*(\S+)`)

func getPriority() string {
	options := getOptions()
	if options == "" {
		return ""
	}
	if match := priorityRgx.FindStringSubmatch(options); len(match) == 2 {
		return match[1]
	}
	return ""
}

var neosServerPortRgx = regexp.MustCompile(`(\S+)\s*:\s*(\d+)`)
var neosServerRgx = regexp.MustCompile(`(\S+)`)

func getNEOSServer() (string, string) {
	/*
		If neos_server is set to host[:port], then return the job and password values
	*/
	host := "neos-server.org"
	port := "3333"
	options := getEnvOption("neos_server")
	if options != "" {
		if match := neosServerPortRgx.FindStringSubmatch(options); len(match) == 3 {
			return match[1], match[2]
		} else if match := neosServerRgx.FindStringSubmatch(options); len(match) == 2 {
			return match[1], port
		}
	}
	return host, port
}

func getEmail() string {
	/*
		Get email provided by user.
	*/
	email := getEnvOption("email")
	return strings.TrimSpace(email)
}

func getAuthenticationOptions() (string, string) {
	/*
		If 'authenticate' is set to '"username", "password"', then return the authentication options
	*/
	username := getEnvOption("neos_username")
	password := getEnvOption("neos_user_password")
	return strings.TrimSpace(username), strings.TrimSpace(password)
}
