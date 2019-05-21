# build script

# Example:
#   make build

.PHONY: build
build:
	go build -o dist/whalefs main.go
	chmod +x dist/whalefs
	mkdir -p dist/config dist/templates
	cp config/* dist/config
	cp templates/* dist/templates