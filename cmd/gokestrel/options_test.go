package main

import (
	"fmt"
	"os"
	"testing"
)

func TestGetJobAndPassword(t *testing.T) {
	var tests = []struct {
		env         string
		value       string
		jobNumber   int
		jobPassword string
	}{
		{"kestrel_options", " job=2746671 password=AnVsgUKc ", 2746671, "AnVsgUKc"},
		{"kestrel_options", "password=AnVsgUKc", 0, "AnVsgUKc"},
		{"kestrel_options", "job=2746671", 2746671, ""},
		{"kestrel_options", "job  =  2746671  password  =  AnVsgUKc  ", 2746671, "AnVsgUKc"},
	}
	for i, tt := range tests {
		testname := fmt.Sprintf("test #%d", i)
		t.Run(testname, func(t *testing.T) {
			os.Setenv(tt.env, tt.value)
			jobNumber, jobPassword := getJobAndPassword()
			os.Unsetenv(tt.env)
			if jobNumber != tt.jobNumber {
				t.Errorf("got '%v', want '%v'", jobNumber, tt.jobNumber)
			}
			if jobPassword != tt.jobPassword {
				t.Errorf("got '%v', want '%v'", jobPassword, tt.jobPassword)
			}
		})
	}
}

func TestPriority(t *testing.T) {
	var tests = []struct {
		env      string
		value    string
		priority string
	}{
		{"kestrel_options", " priority = 1 ", "1"},
		{"kestrel_options", " priorit y = 1 ", ""},
		{"kestrel_options", " priority = abc1 ", "abc1"},
		{"kestrel_options", "", ""},
	}
	for i, tt := range tests {
		testname := fmt.Sprintf("test #%d", i)
		t.Run(testname, func(t *testing.T) {
			os.Setenv(tt.env, tt.value)
			priority := getPriority()
			os.Unsetenv(tt.env)
			if priority != tt.priority {
				t.Errorf("got '%v', want '%v'", priority, tt.priority)
			}
		})
	}
}

func TestGetNEOSServer(t *testing.T) {
	var tests = []struct {
		env   string
		value string
		host  string
		port  string
	}{
		{"neos_server", " neos-server.org:3333 ", "neos-server.org", "3333"},
		{"NEOS_SERVER", " 127.0.0.1:3333 ", "127.0.0.1", "3333"},
		{"neos_server", "   neos-server.org : 123  ", "neos-server.org", "123"},
		{"NEOS_SERVER", "  127.0.0.1  :  456  ", "127.0.0.1", "456"},
		{"neos_server", "neos-server.org:3333", "neos-server.org", "3333"},
		{"NEOS_SERVER", "127.0.0.1:3333", "127.0.0.1", "3333"},
		{"NEOS_SERVER", "127.0.0.1", "127.0.0.1", "3333"},
	}
	for i, tt := range tests {
		testname := fmt.Sprintf("test #%d", i)
		t.Run(testname, func(t *testing.T) {
			os.Setenv(tt.env, tt.value)
			host, port := getNEOSServer()
			os.Unsetenv(tt.env)
			if host != tt.host {
				t.Errorf("got '%v', want '%v'", host, tt.host)
			}
			if port != tt.port {
				t.Errorf("got '%v', want '%v'", port, tt.port)
			}
		})
	}
}

func TestGetEmail(t *testing.T) {
	var tests = []struct {
		env   string
		value string
		email string
	}{
		{"email", " test@test1.com", "test@test1.com"},
		{"email", "test@test2.com", "test@test2.com"},
		{"EMAIL", " test@test3.com  ", "test@test3.com"},
		{"EMAIL", "   test@test4.com ", "test@test4.com"},
	}
	for i, tt := range tests {
		testname := fmt.Sprintf("test #%d", i)
		t.Run(testname, func(t *testing.T) {
			os.Setenv(tt.env, tt.value)
			email := getEmail()
			os.Unsetenv(tt.env)
			if email != tt.email {
				t.Errorf("got '%v', want '%v'", email, tt.email)
			}
		})
	}
}

func TestGetAuthenticationOptions(t *testing.T) {
	var tests = []struct {
		env1, value1 string
		env2, value2 string
		username     string
		password     string
	}{
		{"neos_username", " test@test1.com  ", "neos_user_password", " pwd1 ", "test@test1.com", "pwd1"},
		{"NEOS_USERNAME", " test@test2.com", "neos_user_password", " pwd2", "test@test2.com", "pwd2"},
		{"neos_username", "test@test3.com  ", "NEOS_USER_PASSWORD", "pwd3 ", "test@test3.com", "pwd3"},
		{"NEOS_USERNAME", "test@test4.com", "NEOS_USER_PASSWORD", "pwd4", "test@test4.com", "pwd4"},
	}
	for i, tt := range tests {
		testname := fmt.Sprintf("test #%d", i)
		t.Run(testname, func(t *testing.T) {
			os.Setenv(tt.env1, tt.value1)
			os.Setenv(tt.env2, tt.value2)
			username, password := getAuthenticationOptions()
			os.Unsetenv(tt.env1)
			os.Unsetenv(tt.env2)
			if username != tt.username {
				t.Errorf("got '%v', want '%v'", username, tt.username)
			}
			if password != tt.password {
				t.Errorf("got '%v', want '%v'", password, tt.password)
			}
		})
	}
}

func TestGetSolverName(t *testing.T) {
	var tests = []struct {
		env1, value1 string
		env2, value2 string
		env3, value3 string
		solver       string
		failed       bool
	}{
		{
			"kestrel_options", " solver = gurobi ",
			"email", "test@test.com",
			"", "",
			"", true,
		},
		{
			"KESTREL_OPTIONS", " solver = CpLeX",
			"email", "test@test.com",
			"", "",
			"CPLEX", false,
		},
		{
			"KESTREL_OPTIONS", " solver = CpLeX",
			"email", "test@test.com",
			"neos_server", "127.0.0.1:3333",
			"", true,
		},
		{
			"KESTREL_OPTIONS", " solver = CpLeX",
			"email", "",
			"", "",
			"", true,
		},
		{
			"KESTREL_OPTIONS", "",
			"email", "test@test.com",
			"", "",
			"", true,
		},
	}
	for i, tt := range tests {
		testname := fmt.Sprintf("test #%d", i)
		t.Run(testname, func(t *testing.T) {
			os.Setenv(tt.env1, tt.value1)
			os.Setenv(tt.env2, tt.value2)
			os.Setenv(tt.env3, tt.value3)
			kestrel, err := NewKestrel()
			failed := err != nil
			solver := ""
			if !failed {
				solver, err = kestrel.getSolverName()
				failed = err != nil
			}
			os.Unsetenv(tt.env1)
			os.Unsetenv(tt.env2)
			os.Unsetenv(tt.env3)
			if solver != tt.solver {
				t.Errorf("got '%v', %v, want '%v'", solver, err, tt.solver)
			}
			if tt.failed != failed {
				t.Errorf("got '%v', %v, want '%v'", failed, err, tt.failed)
			}
		})
	}
}
