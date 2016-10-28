#Adding WaitGroup to test commands

###Why:
We want to have tests that use a series of commands, some of which pause, and others that don't.  (See below for an description of how we'll use this on the queueing runner test)

###Implementation approach:
Make runner.Command and execer.Command interfaces (runner.CommandI and execer.CommandI) so that the tests can extend the structures adding a WaitGroup object to each command.  (We don’t want the WaitGroup object to be on the production Command structures)

###First use
Test 1: volume test
generate run tests, where the run tests randomly have their own waitgroup (to simulate a mix of long and short running tests).  
Go routine 1: Add the run tests to the queue while randomly sleeping (< 10 ms) between each add 
Go routine 2: loop through runids up to the max number of run tests: 
Wait for the runid to be available in the queue
if the run test has a waitGroup, wait for it to become running, and send a done signal to the waitGroup, 
validate that the run test completes



[![Build Status](https://travis-ci.org/scootdev/scoot.svg?branch=caitie%2Fscheduler)](https://travis-ci.org/scootdev/scoot)
[![codecov.io](https://codecov.io/github/Kitware/candela/coverage.svg?branch=master)](https://codecov.io/gh/scootdev/scoot?branch=master)

## Try It Out
Setup a scheduler and worker nodes on your laptop by running
```
go run ./binaries/setup-cloud-scoot/main.go --strategy local.local
```

Run a series of randomly generated tests against Scoot
```
go run ./binaries/scootapi/main.go run_smoke_test
```
