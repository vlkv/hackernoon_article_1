all: run

.PHONY: run
run: api ../guest/guest.wasm
	WASMTIME_BACKTRACE_DETAILS=1 go run .

api:
	cd ../api && $(MAKE)

../guest/guest.wasm:
	cd ../guest && $(MAKE)
