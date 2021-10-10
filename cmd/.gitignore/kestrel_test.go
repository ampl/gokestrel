package main

import (
	"os"
	"testing"
)

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
