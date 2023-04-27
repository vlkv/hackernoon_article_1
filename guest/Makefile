all: build

.PHONY: build
build: guest.wasm

api:
	cd ../api && $(MAKE)

guest.wasm: main.go mem.go api
	tinygo build -target=wasi -o guest.wasm .
