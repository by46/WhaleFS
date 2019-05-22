# build script

# Example:
#   make build

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o dist/whalefs main.go
	mkdir -p dist/config dist/templates
	cp config/* dist/config
	cp templates/* dist/templates