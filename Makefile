# build script

# Example:
#   make build

.PHONY: build
build:
	go build -o dist/whalefs main.go
	cp config/* dist