# GoLab: Binary Search Trees

## Building
To build, run `make` and run each step individually.

## Running
To find which command line flags are available for each executable, run `./step#.out -h`. Each takes an amout of hash, data, and comparison workers, however, this data is only utilized in the later steps. You can also use the makefile rules to run a specific file and set the variables by hand.

## Timing
The timing harness takes 3 arguments: max hash workers, max data workers, max comparison workers. It will iterate through powers of 2 until it reaches or passes the max numbers provided. It will run the sequential solution for the first graph, the step 2 tests on fine.txt and coarse.txt, and the step 3 comparison testing on coarse.txt. It saves the resulting graphs as pdfs in the immediate directory.
