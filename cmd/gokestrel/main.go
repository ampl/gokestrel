package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var Version = "development"

type Job = struct {
	jobNumber int
	password  string
}

func listJobs(jobsFile string) ([]Job, error) {
	f, err := os.Open(jobsFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	jobs := []Job{}
	for {
		job := Job{}
		_, err := fmt.Fscanf(f, "%d %s", &job.jobNumber, &job.password)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func writeJobs(jobs []Job, jobsFile string) error {
	if len(jobs) == 0 {
		if err := os.Remove(jobsFile); err != nil {
			return err
		}
	}
	f, err := os.Create(jobsFile)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, job := range jobs {
		if _, err := fmt.Fprintf(f, "%d %s\n", job.jobNumber, job.password); err != nil {
			defer os.Remove(jobsFile)
			return err
		}
	}
	return nil
}

func submit(stub string) (int, error) {
	stub = strings.TrimSuffix(stub, ".nl")
	k, err := NewKestrel()
	if err != nil {
		return 1, err
	}
	fmt.Printf("Submitting model at %s\n", stub+".nl")
	xml, err := k.formXML(stub)
	if err != nil {
		return 1, err
	}
	jobNumber, password, err := k.submit(xml)
	if err != nil {
		return 1, err
	}
	// Add the job, pass to the stack
	fname := jobsFile()
	jobs, err := listJobs(fname)
	if err != nil {
		return 1, err
	}
	jobs = append(jobs, Job{jobNumber, password})
	err = writeJobs(jobs, fname)
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func retrieve(stub string) (int, error) {
	fname := jobsFile()
	jobs, err := listJobs(fname)
	if err != nil {
		return 1, err
	}
	if len(jobs) == 0 {
		fmt.Printf("Error, could not open file %s.\n", fname)
		fmt.Printf("Did you use kestrelsub?\n")
		return 1, nil
	}
	k, err := NewKestrel()
	if err != nil {
		return 1, err
	}
	if err := k.retrieve(stub, jobs[0].jobNumber, jobs[0].password); err != nil {
		return 1, err
	}
	if len(jobs) > 1 {
		fmt.Println("restofstack: ")
		for _, job := range jobs[1:] {
			fmt.Printf("%d %s\n", job.jobNumber, job.password)
		}
	}
	if err := writeJobs(jobs[1:], fname); err != nil {
		return 1, err
	}
	return 0, nil
}

func kill(jobNumber int, password string) (int, error) {
	k, err := NewKestrel()
	if err != nil {
		return 1, err
	}
	err = k.kill(jobNumber, password)
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func solve(stub string, sigint chan os.Signal) (int, error) {
	k, err := NewKestrel()
	if err != nil {
		return 1, err
	}
	jobNumber := 0
	password := ""
	errors := make(chan error)
	go func() {
		// See if kestrel_options has job=.. password=..
		jobNumber, password = getJobAndPassword()
		// otherwise, submit current problem to NEOS
		if jobNumber == 0 {
			xml, err := k.formXML(stub)
			if err != nil {
				errors <- err
				return
			}
			jobNumber, password, err = k.submit(xml)
			if err != nil {
				errors <- err
				return
			}
		}
		errors <- nil
	}()
	select {
	case <-sigint:
		fmt.Println("Keyboard Interrupt while submitting problem.")
		return 1, nil
	case err := <-errors:
		if err != nil {
			return 1, err
		}
	}
	offset := 0
	output := ""
	status := "Running"
	time.Sleep(1 * time.Second)
	for status == "Running" || status == "Waiting" {
		output, offset, err = k.getIntermediateResults(jobNumber, password, offset)
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("%s", output)
		status, err = k.getJobStatus(jobNumber, password)
		if err != nil {
			log.Println(err)
		}
		select {
		case <-sigint:
			fmt.Printf("Keyboard Interrupt\n")
			fmt.Printf("Job is still running on remote machine\n")
			fmt.Printf("To stop job:\n")
			fmt.Printf("\tampl: option kestrel_options \"job=%d password=%s\";\n", jobNumber, password)
			fmt.Printf("\tampl: commands kestrelkill;\n")
			fmt.Printf("To retrieve results:\n")
			fmt.Printf("\tampl: option kestrel_options \"job=%d password=%s\";\n", jobNumber, password)
			fmt.Printf("\tampl: solve;\n")
			return 1, nil
		case <-time.After(5 * time.Second):
		}
	}
	err = k.retrieve(stub, jobNumber, password)
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func run(args []string) (int, error) {
	if len(args) == 2 && (args[1] == "-v" || args[1] == "version") {
		fmt.Printf("kestrel version %v %v/%v\n", Version, runtime.GOOS, runtime.GOARCH)
		return 0, nil
	} else if len(args) >= 2 && len(args) <= 3 && args[1] == "submit" {
		stub := getEnvOption("kestrel_stub")
		if stub == "" {
			stub = "kmodel"
		}
		if len(args) == 3 {
			stub = args[2]
		}
		return submit(stub)
	} else if len(args) >= 2 && len(args) <= 3 && args[1] == "retrieve" {
		stub := getEnvOption("kestrel_stub")
		if stub == "" {
			stub = "kmodel"
		}
		if len(args) == 3 {
			stub = args[2]
		}
		return retrieve(stub)
	} else if (len(args) == 2 || len(args) == 4) && args[1] == "kill" {
		jobNumber, password := getJobAndPassword()
		if len(args) == 4 {
			n, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil {
				return 1, err
			}
			jobNumber = int(n)
			password = args[3]
		} else if jobNumber == 0 {
			fmt.Println("To kill a NEOS job, first set kestrel_options variable:")
			fmt.Println("\tampl: option kestrel_options \"job=#### password=xxxx\";")
			return 1, nil
		}
		return kill(jobNumber, password)
	} else if len(args) == 3 && args[2] == "-AMPL" {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		return solve(args[1], sigint)
	}
	fmt.Println("kestrel should be called from inside AMPL.")
	return 1, nil
}

func main() {
	code, err := run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}
