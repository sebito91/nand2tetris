BINARYNAME      := spaces 
BINARYENDING    :=
ifeq ($(OS),Windows_NT)
	BINARYENDING = .exe
endif

all: clean
	mkdir ./build
	go build -v -o ./build/$(BINARYNAME)$(BINARYENDING) ./src/main.go

.PHONY: clean
clean:
	rm -rf ./build $(BINARYNAME)$(BINARYENDING)
