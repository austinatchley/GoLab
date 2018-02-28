WRK=1
FILE=coarse
PRNT=
LCK=

all: build
	
run: build
	./step3.out -input=sample/$(FILE).txt -hash-workers=$(WRK) $(PRNT) $(LCK)

build:
	go build -o tree.out tree.go
	go build -o step2.out step2.go
	go build -o step3.out step3.go

fmt:
	go fmt *.go

time: build
	./timer_harness.py 1 1 1 sample/coarse.txt

clean:
	rm *.out

simple: build
simple: FILE=simple
simple: run

fine: build
fine: FILE=fine
fine: run

2wrk: WRK=2
3wrk: simple
