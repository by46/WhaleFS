# build script

# Example:
#   make build

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -mod=vendor -o dist/whalefs main.go
	mkdir -p dist/config dist/templates dist/i18n
	cp config/* dist/config
	cp templates/* dist/templates
	cp i18n/* dist/i18n

build-win:
	GOOS=windows GOARCH=amd64 go build -o dist/whalefs.exe main.go
	mkdir -p dist/config dist/templates dist/i18n
	cp config/* dist/config
	cp templates/* dist/templates
	cp i18n/* dist/i18n

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o dist/whalefs main.go
	mkdir -p dist/config dist/templates dist/i18n
	cp config/* dist/config
	cp templates/* dist/templates
	cp i18n/* dist/i18n