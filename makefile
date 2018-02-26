all: build
	
run: build
	./tree.out -input=sample/coarse.txt

build:
	go build -o tree.out tree.go

fmt:
	go fmt *.go

time: build
	./timer_harness.py 1 1 1 sample/coarse.txt

clean:
	rm *.out

simple: build
	./tree.out -input=sample/simple.txt

fine: build
	./tree.out -input=sample/fine.txt
