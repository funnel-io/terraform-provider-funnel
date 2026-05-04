.PHONY: build format test docs all

build:
	go vet
	go build -o terraform-provider-funnel

format:
	go fmt
	terraform fmt -recursive

test:
	go test -v ./provider/**/

docs:
	cd tools
	go generate ./...

all: format build test docs
