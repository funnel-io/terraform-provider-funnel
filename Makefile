.PHONY: build format test docs

build:
	go vet
	go build -o terraform-provider-funnel

format:
	go fmt

test:
	go test -v ./provider/**/

docs:
	cd tools
	go generate ./...
