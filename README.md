# GoKestrel: Kestrel AMPL client for all platforms

This project replicates https://github.com/NEOS-Server/Kestrel-AMPL-Linux/ in Go in order to produce binaries that can be easily distributed for various platforms without requiring Python to be installed.

Protocol: [NEOS XML-RPC protocol](https://neos-server.org/neos/xml-rpc.html)

Downlaod from: https://ampl.com/dl/fdabrandao/gokestrel/

## Usage

### Solving directly

```bash
ampl: model diet.mod;
ampl: data diet.dat;
ampl: option solver kestrel;
ampl: option email "***@***.***";
ampl: option kestrel_options "solver=cplex";
ampl: option cplex_options "display=2";
ampl: solve;
Connecting to: neos-server.org:3333
Job XXXX submitted to NEOS, password='xxxx'
Check the following URL for progress report:
https://neos-server.org/neos/cgi-bin/nph-neos-solver.cgi?admin=results&jobnumber=XXXX&pass=xxxx
Job XXXX dispatched
password: xxxx
---------- Begin Solver Output -----------
Condor submit: 'neos.submit'
Condor submit: 'watchdog.submit'
Job submitted to NEOS HTCondor pool.
CPLEX 20.1.0.0: optimal solution; objective 88.2
1 dual simplex iterations (0 in phase I)
```

### Using commands for asyncronous submissions

The command files `kestrelsub`, `kestrelret`, and `kestrelkill` are available at [commands/](commands/). To insure that AMPL will find the scripts, place them in the directory (or folder) that will be current when you execute AMPL, or set option `ampl_include` to specify the directory where the script can be found.

```bash
$ ampl
ampl: model diet.mod;
ampl: data diet.dat;
ampl: option email "***@***.***";
ampl: option kestrel_options "solver=cplex";
ampl: commands kestrelsub;
Connecting to: neos-server.org:3333
Submitting model at kmodel.nl
Job XXXX submitted to NEOS, password='xxxx'
Check the following URL for progress report:
https://neos-server.org/neos/cgi-bin/nph-neos-solver.cgi?admin=results&jobnumber=XXXX&pass=xxxx
ampl: commands kestrelret;
Connecting to: neos-server.org:3333
Writting solution to kmodel.sol
CPLEX 20.1.0.0: optimal solution; objective 88.2
0 simplex iterations (0 in phase I)
ampl: option kestrel_options 'job=XXXX password=xxxx';
ampl: commands kestrelkill;
Connecting to: neos-server.org:3333
Job XXXX is finished
```

### Using shell for asyncronous submissions

If the folder containing AMPL and all solvers including kestrel is in the environment variable PATH,
it may be more convenient to submit/retrieve/kill jobs invoking kestrel directly instead of using commands as follows:

```bash
$ ampl
ampl: model diet.mod;
ampl: data diet.dat;
ampl: option ampl_id (_pid); # get a submission queue (first-in-first-out) for this AMPL session
ampl: option email "***@***.***"; # required by NEOS
ampl: option kestrel_options "solver=cplex"; # solver to use on NEOS
ampl: option kestrel_stub "kmodel"; # stub file used by kestrel submit/retrieve
ampl: write bkmodel; # write model to kmodel.nl
ampl: shell "kestrel submit"; # submit the job
Connecting to: neos-server.org:3333
Submitting model at kmodel.nl
Job XXXX submitted to NEOS, password='xxxx'
Check the following URL for progress report:
https://neos-server.org/neos/cgi-bin/nph-neos-solver.cgi?admin=results&jobnumber=XXXX&pass=xxxx
ampl: shell "kestrel retrieve"; # retrieve the job
Connecting to: neos-server.org:3333
Writting solution to kmodel.sol
ampl: solution kmodel.sol; # load the solution
CPLEX 20.1.0.0: optimal solution; objective 88.2
1 dual simplex iterations (0 in phase I)
ampl: shell "kestrel kill XXXX xxxx"; # to kill a job
Connecting to: neos-server.org:3333
Job XXXX is finished
```

### Authenticated submissions

For authenticated submissions set `neos_username` and `neos_user_password` as follows:
```bash
ampl: option neos_username 'username';
ampl: option neos_user_password 'password';
```

### Priority

Jobs submitted with the priority of long can run for at most 8 hours. Jobs submitted with the priority short can run for at most 5 minutes. Results for long jobs do not stream. You can control the priority as follows:
```bash
ampl: option kestrel_options "solver=xxx priority=short";
```
or
```bash
ampl: option kestrel_options "solver=xxx priority=long";
```
In this driver we set the default priority to short so that we can retrieve output from the solver.

## License

BSD-3

***
Copyright ?? 2021 AMPL Optimization inc. All rights reserved.
