all: build
	
run: build
	./tree.out -input=sample/simple.txt

build:
	go build -o tree.out tree.go

time: build
	./timer_harness.py 1 1 1 sample/simple.txt

