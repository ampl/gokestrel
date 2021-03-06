package main

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// setup
	_ = os.Remove(jobsFile())
	code := m.Run()
	// shutdown
	_ = os.Remove("kestresult.sol")
	os.Exit(code)
}

func TestSubmitKillRetrieve(t *testing.T) {
	stub := getEnvOption("TEST_STUB")
	email := getEnvOption("TEST_EMAIL")
	if stub == "" || email == "" {
		t.Skip("Skipping due to missing TEST_STUB and/or TEST_EMAIL")
	}
	os.Setenv("email", email)
	os.Setenv("kestrel_options", "solver=cplex")
	k, err := NewKestrel()
	if err != nil {
		t.Fatalf("NewKestrel failed with '%v'", err)
	}
	xml, err := k.formXML(stub)
	if err != nil {
		t.Fatalf("k.formXML failed with '%v'", err)
	}
	jobNumber, password, err := k.submit(xml)
	if err != nil {
		t.Fatalf("k.submit failed with '%v'", err)
	}
	err = k.kill(jobNumber, password)
	if err != nil {
		t.Fatalf("k.kill failed with '%v'", err)
	}
	err = k.retrieve(stub, jobNumber, password)
	if err != nil {
		t.Fatalf("k.retrieve failed with '%v'", err)
	}
}

func TestSolve(t *testing.T) {
	stub := getEnvOption("TEST_STUB")
	email := getEnvOption("TEST_EMAIL")
	if stub == "" || email == "" {
		t.Skip("Skipping due to missing TEST_STUB and/or TEST_EMAIL")
	}
	os.Setenv("email", email)
	os.Setenv("kestrel_options", "solver=cplex")
	sigint := make(chan os.Signal, 1)
	exit, err := solve(stub, sigint)
	if want := 0; exit != want {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	sigint <- os.Interrupt
	exit, err = solve(stub, sigint)
	if want := 1; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	go func() {
		time.Sleep(5 * time.Second)
		sigint <- os.Interrupt
	}()
	exit, err = solve(stub, sigint)
	if want := 1; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
}

func TestAuthenticated(t *testing.T) {
	stub := getEnvOption("TEST_STUB")
	email := getEnvOption("TEST_EMAIL")
	username := getEnvOption("TEST_USERNAME")
	password := getEnvOption("TEST_PASSWORD")
	if stub == "" || email == "" || username == "" || password == "" {
		t.Skip("Skipping due to missing TEST_STUB, TEST_EMAIL, TEST_USERNAME and/or TEST_PASSWORD")
	}
	defer unsetEnv("email", "kestrel_options", "neos_username", "neos_user_password")
	os.Setenv("email", email)
	os.Setenv("kestrel_options", "solver=cplex")
	os.Setenv("neos_username", username)
	os.Setenv("neos_user_password", password)
	sigint := make(chan os.Signal, 1)
	exit, err := solve(stub, sigint)
	if want := 0; exit != want {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	os.Setenv("neos_user_password", "wrong_password")
	exit, err = solve(stub, sigint)
	if want := 1; exit != want || err == nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
}

func TestPriority(t *testing.T) {
	stub := getEnvOption("TEST_STUB")
	email := getEnvOption("TEST_EMAIL")
	if stub == "" || email == "" {
		t.Skip("Skipping due to missing TEST_STUB and/or TEST_EMAIL")
	}
	defer unsetEnv("email", "kestrel_options", "cplex_options")
	os.Setenv("email", email)
	os.Setenv("kestrel_options", "solver=cplex priority=short")
	os.Setenv("cplex_options", "lpdisplay=1")
	sigint := make(chan os.Signal, 1)
	exit, err := solve(stub, sigint)
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	os.Setenv("kestrel_options", "solver=cplex priority=long")
	exit, err = solve(stub, sigint)
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
}

func TestSubmitRetrieve(t *testing.T) {
	stub := getEnvOption("TEST_STUB")
	email := getEnvOption("TEST_EMAIL")
	if stub == "" || email == "" {
		t.Skip("Skipping due to missing TEST_STUB and/or TEST_EMAIL")
	}
	os.Setenv("email", email)
	os.Setenv("kestrel_options", "solver=cplex")
	exit, err := submit(stub)
	if want := 0; exit != want {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = retrieve(stub)
	if want := 0; exit != want {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
}

func TestRunSolve(t *testing.T) {
	stub := getEnvOption("TEST_STUB")
	email := getEnvOption("TEST_EMAIL")
	if stub == "" || email == "" {
		t.Skip("Skipping due to missing TEST_STUB and/or TEST_EMAIL")
	}
	os.Setenv("email", email)
	os.Setenv("kestrel_options", "solver=cplex")
	exit, err := run([]string{"kestrel", stub, "-AMPL"})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
}

func TestRunFail(t *testing.T) {
	exit, err := run([]string{"kestrel"})
	if want := 1; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
}

func TestRunSubmitKill(t *testing.T) {
	stub := getEnvOption("TEST_STUB")
	email := getEnvOption("TEST_EMAIL")
	if stub == "" || email == "" {
		t.Skip("Skipping due to missing TEST_STUB and/or TEST_EMAIL")
	}
	defer unsetEnv("email", "kestrel_options")
	os.Setenv("email", email)
	os.Setenv("kestrel_options", "solver=cplex")
	exit, err := run([]string{"kestrel", "submit", stub})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	fname := jobsFile()
	jobs, err := listJobs(fname)
	if err != nil {
		t.Fatalf("listJobs failed with '%v'", err)
	}
	exit, err = run([]string{"kestrel", "kill"}) // should fail
	if want := 1; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	for _, job := range jobs {
		exit, err = run([]string{"kestrel", "kill", strconv.Itoa(job.jobNumber), job.password})
		if want := 0; exit != want || err != nil {
			t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
		}
		os.Setenv("kestrel_options", fmt.Sprintf("job=%d password=%s", job.jobNumber, job.password))
		exit, err = run([]string{"kestrel", "kill"})
		if want := 0; exit != want || err != nil {
			t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
		}
		exit, err = run([]string{"kestrel", "retrieve", stub})
		if want := 0; exit != want || err != nil {
			t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
		}
	}
}

func TestRunSubmitRetrieve(t *testing.T) {
	stub := getEnvOption("TEST_STUB")
	email := getEnvOption("TEST_EMAIL")
	if stub == "" || email == "" {
		t.Skip("Skipping due to missing TEST_STUB and/or TEST_EMAIL")
	}
	os.Setenv("email", email)
	os.Setenv("kestrel_options", "solver=cplex")
	exit, err := run([]string{"kestrel", "retrieve", stub}) // should fail
	if want := 1; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "submit", stub})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "retrieve", stub})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "submit", stub})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "submit", stub})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "retrieve", stub})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "retrieve", stub})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "retrieve", stub}) // should fail
	if want := 1; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
}
