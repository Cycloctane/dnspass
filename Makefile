VERSION=$(shell git describe --tags --always)
FLAGS=-ldflags="-s -w -X main.version=$(VERSION)" -trimpath
OUTPUT_DIR=build

.PHONY: all clean test windows_x64 linux_x64

default: all

test:
	go test -v ./internal/records/

windows_x64:
	GOOS=windows GOARCH=amd64 go build -v -o $(OUTPUT_DIR)/dnspass_$(VERSION)_$@.exe $(FLAGS) ./cmd/dnspass

linux_x64:
	GOOS=linux GOARCH=amd64 go build -v -o $(OUTPUT_DIR)/dnspass_$(VERSION)_$@ $(FLAGS) ./cmd/dnspass

all: windows_x64 linux_x64

clean:
	rm -f $(OUTPUT_DIR)/*
