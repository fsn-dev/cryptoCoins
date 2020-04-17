# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: all coins clean fmt

all:
	./build.sh coins 
	@echo "Done building."

coins:
	./build.sh coins
	@echo "Done building."

clean:
	rm -fr ./bin/cmd/* 

fmt:
	./gofmt.sh
