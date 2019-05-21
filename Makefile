# build script

# Example:
#   make build

.PHONY: build
build:
	go build -o dist/whalefs main.go
	mkdir -p dist/config
	cp config/* dist/config