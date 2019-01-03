BINARYNAME      := compiler 
BINARYENDING    :=
ifeq ($(OS),Windows_NT)
	BINARYENDING = .exe
endif

all: clean
	mkdir ./build
	go build -v -o ./build/$(BINARYNAME)$(BINARYENDING) ./src/uchicago/${BINARYNAME}/main.go

.PHONY: clean
clean:
	rm -rf ./build ./build/$(BINARYNAME)$(BINARYENDING)
