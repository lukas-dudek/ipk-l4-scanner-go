.PHONY: all build clean test NixDevShellName

all: build

build:
	go build -o ipk-L4-scan .
	chmod +x ipk-L4-scan

NixDevShellName:
	@echo go

test:
	go test -v ./...
	chmod +x test_scanner.sh
	./test_scanner.sh

clean:
	rm -f ipk-L4-scan
