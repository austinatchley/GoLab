WRK=1

all: build
	
run: build
	./step2.out -input=sample/coarse.txt -hash-workers=$(WRK)

build:
	go build -o tree.out tree.go
	go build -o step2.out step2.go

fmt:
	go fmt *.go

time: build
	./timer_harness.py 1 1 1 sample/coarse.txt

clean:
	rm *.out

simple: build
	./step2.out -input=sample/simple.txt -hash-workers=$(WRK)

fine: build
	./step2.out -input=sample/fine.txt -hash-workers=$(WRK)

2wrk: WRK=2
2wrk: simple
