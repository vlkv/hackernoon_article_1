module host

go 1.19

require (
	github.com/bytecodealliance/wasmtime-go v1.0.0
	karmem.org v1.2.9
)

require api v0.0.1

replace api v0.0.1 => ../api
