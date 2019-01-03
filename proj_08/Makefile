BINARYNAME      := bigvm 
BINARYENDING    :=
ifeq ($(OS),Windows_NT)
	BINARYENDING = .exe
endif

CURRENTDIR      := $(shell pwd)

all: clean
	mkdir ./build
	export GOPATH=$(CURRENTDIR) && go build -v -o ./build/$(BINARYNAME)$(BINARYENDING) ./src/bigvm/main.go

.PHONY: clean
clean:
	rm -rf ./build ./build/$(BINARYNAME)$(BINARYENDING)
