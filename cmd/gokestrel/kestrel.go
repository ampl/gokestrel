package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"alexejk.io/go-xmlrpc"
)

type Kestrel struct {
	Host, Port             string
	Username, UserPassword string
	Email                  string
	Client                 *xmlrpc.Client
}

func NewKestrel() (*Kestrel, error) {
	host, port := getNEOSServer()
	username, password := getAuthenticationOptions()
	email := getEmail()
	if email == "" {
		return nil, fmt.Errorf("An email address is required for NEOS submissions.\n" +
			"To set: option email \"<address>\";\n\n")
	}
	client, err := xmlrpc.NewClient(fmt.Sprintf("https://%s:%s", host, port))
	if err == nil {
		err = client.Call("ping", nil, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("Error, NEOS solver is temporarily unavailable. Error: %v", err)
	}
	return &Kestrel{
		Host:         host,
		Port:         port,
		Username:     username,
		UserPassword: password,
		Email:        email,
		Client:       client,
	}, nil
}

func (k *Kestrel) submit(xml string) (int, string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return 0, "", err
	}
	user := fmt.Sprintf("%s on %s", getEnvOption("LOGNAME"), hostname)
	result := struct {
		Results []interface{}
	}{}
	if k.Username == "" || k.UserPassword == "" {
		request := struct {
			Xml  string
			User string
		}{xml, user}
		if err := k.Client.Call("submitJob", &request, &result); err != nil {
			return 0, "", err
		}
	} else {
		request := struct {
			Xml      string
			Username string
			Password string
		}{xml, k.Username, k.UserPassword}
		if err := k.Client.Call("authenticatedSubmitJob", &request, &result); err != nil {
			return 0, "", err
		}
	}
	var jobNumber int
	if len(result.Results) == 2 {
		if v, ok := result.Results[0].(int); ok {
			jobNumber = v
		}
	}
	if jobNumber == 0 {
		return 0, "", fmt.Errorf("Error: %v\nJob not submitted.\n", result.Results[1])
	}

	password := ""
	if v, ok := result.Results[1].(string); ok {
		password = v
	}
	fmt.Printf("Job %d submitted to NEOS, password='%s'\n", jobNumber, password)
	fmt.Printf("Check the following URL for progress report:\n")
	fmt.Printf(
		"https://%s/neos/cgi-bin/nph-neos-solver.cgi?admin=results&jobnumber=%d&pass=%s\n",
		k.Host, jobNumber, password)

	return jobNumber, password, nil
}

func (k *Kestrel) retrieve(stub string, jobNumber int, password string) error {
	stub = strings.TrimSuffix(stub, ".nl")
	request := struct {
		JobNumber int
		Password  string
	}{jobNumber, password}
	result := struct {
		Solution string
	}{}
	if err := k.Client.Call("getFinalResults", &request, &result); err != nil {
		return err
	}
	_, err := writeToFile(result.Solution, stub+".sol")
	return err
}

func (k *Kestrel) kill(jobNumber int, password string) error {
	request := struct {
		JobNumber int
		Password  string
	}{jobNumber, password}
	result := struct {
		Response string
	}{}
	err := k.Client.Call("killJob", &request, &result)
	if err != nil {
		return err
	}
	fmt.Println(result.Response)
	return nil
}

func (k *Kestrel) getIntermediateResults(jobNumber int, password string, offset int) (string, int, error) {
	request := struct {
		JobNumber int
		Password  string
		Offset    int
	}{jobNumber, password, offset}
	result := struct {
		Results []interface{}
	}{}
	if err := k.Client.Call("getIntermediateResults", &request, &result); err != nil {
		return "", 0, err
	}
	output := ""
	if len(result.Results) == 2 {
		if v, ok := result.Results[0].([]byte); ok {
			output = string(v)
		}
		if v, ok := result.Results[1].(int); ok {
			offset = v
		}
	}
	return output, offset, nil
}

func (k *Kestrel) getJobStatus(jobNumber int, password string) (string, error) {
	request := struct {
		JobNumber int
		Password  string
	}{jobNumber, password}
	result := struct {
		Status string
	}{}
	if err := k.Client.Call("getJobStatus", &request, &result); err != nil {
		return "", err
	}
	return result.Status, nil
}

var solverRgx = regexp.MustCompile(`(?i)solver\s*=*\s*(\S+)`)

func (k *Kestrel) getSolverName() (string, error) {
	/*
		Read in the kestrel_options to pick out the solver name.
			The tricky parts:
				we don't want to be case sensitive, but NEOS is.
				we need to read in options variable
	*/
	// Get a list of available kestrel solvers from NEOS
	request := struct {
		Category string
	}{"kestrel"}
	result := struct {
		Solvers []string
	}{}
	if err := k.Client.Call("listSolversInCategory", &request, &result); err != nil {
		return "", err
	}
	allKestrelSolvers := result.Solvers
	kestrelAmplSolvers := []string{}
	for _, s := range allKestrelSolvers {
		if strings.HasSuffix(s, ":AMPL") {
			kestrelAmplSolvers = append(kestrelAmplSolvers, strings.TrimSuffix(s, ":AMPL"))
		}
	}
	chooseFrom := "Choose from:\n"
	for _, s := range kestrelAmplSolvers {
		chooseFrom += fmt.Sprintf("\t%s\n", s)
	}
	chooseFrom += "\nTo choose: option kestrel_options \"solver=xxx\";\n\n"

	// Read kestrel_options to get solver name
	options := getOptions()

	solverName := ""
	if match := solverRgx.FindStringSubmatch(options); len(match) == 2 {
		solverName = match[1]
	}

	if options == "" || solverName == "" {
		return "", fmt.Errorf("No solver name selected. %s", chooseFrom)
	}

	neosSolverName := ""
	for _, s := range kestrelAmplSolvers {
		if strings.EqualFold(s, solverName) {
			neosSolverName = s
		}
	}

	if neosSolverName == "" {
		return "", fmt.Errorf("%s is not available on NEOS. %s", solverName, chooseFrom)
	}

	return neosSolverName, nil
}

func (k *Kestrel) formXML(stub string) (string, error) {
	/*
		Create xml file for this problem
	*/
	stub = strings.TrimSuffix(stub, ".nl")
	solver, err := k.getSolverName()
	if err != nil {
		return "", err
	}

	// Get priority
	priority := getPriority()
	if priority != "" {
		priority = fmt.Sprintf("<priority>%s</priority>\n", priority)
	}

	// Collect AMPL-created environment variables
	solverOptions := fmt.Sprintf("kestrel_options:solver=%s\n", strings.ToLower(solver))
	solverOptionsKey := fmt.Sprintf("%s_options", solver)

	solverOptionsValue := getEnvOption(solverOptionsKey)
	if solverOptionsValue != "" {
		solverOptions += fmt.Sprintf("%s_options:%s\n", strings.ToLower(solver), solverOptionsValue)
	}

	source, err := os.Open(stub + ".nl")
	if err != nil {
		return "", err
	}
	defer source.Close()
	buf := new(bytes.Buffer)
	destination := gzip.NewWriter(buf)
	if _, err := io.Copy(destination, source); err != nil {
		return "", err
	}
	if err := destination.Close(); err != nil {
		return "", err
	}

	xml := fmt.Sprintf(`
	<document>
	<category>kestrel</category>
	<solver>%s</solver>
	<inputType>AMPL</inputType>
	<email>%s</email>
	%s
	<solver_options>%s</solver_options>
	<nlfile><base64>%s</base64></nlfile>\n`, solver, k.Email, priority,
		solverOptions, base64.StdEncoding.EncodeToString(buf.Bytes()))

	for _, key := range []string{"adj", "col", "env", "fix", "spc", "row", "slc", "unv"} {
		if content, err := ioutil.ReadFile(stub + key); err == nil && len(content) != 0 {
			xml += fmt.Sprintf("<%s><![CDATA[%s]]></%s>\n", key, content, key)
		}
	}

	for _, option := range []string{"kestrel_auxfiles", "mip_priorities", "objective_precision"} {
		if v, ok := os.LookupEnv(option); ok {
			xml += fmt.Sprintf("<%s><![CDATA[%s]]></%s>\n", option, v, option)
		}
	}

	xml += "</document>"
	return xml, nil
}

func writeToFile(content string, fname string) (int, error) {
	f, err := os.Create(fname)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	n, err := f.WriteString(content)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func jobsFile() string {
	return path.Join(os.TempDir(), fmt.Sprintf("at%s.jobs", getEnvOption("ampl_id")))
}
