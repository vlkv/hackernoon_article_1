all: build

.PHONY: build
build: v1/api_generated.go

v1/api_generated.go:
	go run karmem.org/cmd/karmem build --golang -o "v1" api.km
