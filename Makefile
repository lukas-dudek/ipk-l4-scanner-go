.PHONY: all build clean test NixDevShellName

all: build

build:
	go build -o ipk-L4-scan .
	chmod +x ipk-L4-scan

NixDevShellName:
	@echo go

test:
	go test -v ./...

clean:
	rm -f ipk-L4-scan
