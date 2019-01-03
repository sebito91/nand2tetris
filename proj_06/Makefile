BINARYNAME      := assembler 
BINARYENDING    :=
ifeq ($(OS),Windows_NT)
	BINARYENDING = .exe
endif

all: clean
	mkdir ./build
	/usr/bin/go build -v -o ./build/$(BINARYNAME)$(BINARYENDING) ./src/assembler/main.go

.PHONY: clean
clean:
	rm -rf ./build ./build/$(BINARYNAME)$(BINARYENDING)
