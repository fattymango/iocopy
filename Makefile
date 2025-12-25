.PHONY: build-linux build-windows run

build-linux:
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o build/ipscan-linux main.go

build-windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o build/ipscan-windows main.go

run:
	@go run .