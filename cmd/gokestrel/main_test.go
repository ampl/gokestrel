package main

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// setup
	_ = os.Remove(jobsFile())
	code := m.Run()
	// shutdown
	_ = os.Remove("kestresult.sol")
	os.Exit(code)
}

func TestSolve(t *testing.T) {
	stub := getEnvOption("TEST_STUB")
	email := getEnvOption("TEST_EMAIL")
	if stub == "" || email == "" {
		t.Skip("Skipping due to missing TEST_STUB and/or TEST_EMAIL")
	}
	os.Setenv("email", email)
	os.Setenv("kestrel_options", "solver=cplex")
	exit, err := solve(stub)
	if want := 0; exit != want {
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
	os.Setenv("email", email)
	os.Setenv("kestrel_stub", stub)
	os.Setenv("kestrel_options", "solver=cplex")
	exit, err := run([]string{"kestrel", "submit"})
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
		os.Setenv("kestrel_options", fmt.Sprintf("job=%d password=%s", job.jobNumber, job.password))
		exit, err = run([]string{"kestrel", "kill"})
		if want := 0; exit != want || err != nil {
			t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
		}
		exit, err = run([]string{"kestrel", "retrieve"})
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
	os.Setenv("kestrel_stub", stub)
	os.Setenv("kestrel_options", "solver=cplex")
	exit, err := run([]string{"kestrel", "retrieve"}) // should fail
	if want := 1; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "submit"})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "retrieve"})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "submit"})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "submit"})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "retrieve"})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "retrieve"})
	if want := 0; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
	exit, err = run([]string{"kestrel", "retrieve"}) // should fail
	if want := 1; exit != want || err != nil {
		t.Fatalf("got '%v', '%v', want '%v'", exit, err, want)
	}
}
