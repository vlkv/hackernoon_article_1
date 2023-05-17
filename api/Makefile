all: build

build: v1/api_generated.go

v1/api_generated.go: api.km
	go run karmem.org/cmd/karmem build --golang -o "v1" api.km
