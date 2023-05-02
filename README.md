# Passing and returning structured datatypes to and from Go WebAssembly module

[WebAssembly](https://webassembly.org/) is a great technology and has a lot of nice features: it is multiplatform, has
near native performance, it can work in browser and on server side, etc. But (maybe) due to the fact that WebAssembly is
rather young, some basic tasks are not as easy as they expected to be especially for the newcomers. One of such
surprisingly difficult tasks is passing and returning to and from WebAssembly module any complex objects, because it is
well known, that Wasm supports only primitive datatypes (int32, int64, float32 and float64). In fact, passing any
complex objects like arrays, strings, structs with named fields all could be reduced to one single problem of passing
arrays of bytes and applying some serialization/deserialization algorithm to the data.

The general approach to accomplish this is intuitively quite simple - allocate some memory on the guest side (the Wasm
module side) and copy request's data from the host to that memory buffer. Then pass the pointer to that memory + buffer
size to the guest, process the data on the guest side according to the guest's business logic and produce some result.
Then allocate some memory for the result, copy result's bytes into that buffer and return similar pair (pointer + buffer
size) from the Wasm module to the host. Finally don't forget to correctly free all previously allocated memory buffers.
When we start thinking about memory management in WebAssembly, it is very dependent on what language [^1] was used for
subsequent compilation to Wasm instructions. Some languages have garbage collector (GC) "out of the box", some have not.
It is also not a trivial task to choose serialization format and library for interpreting the passed bytes. In the end,
when it comes to writing the exact code that correctly works and could be safely deployed to production... things may
become a little bit complicated at least.

This problem is exacerbated by the fact that the information that could be found in the Internet (on official Wasm
resources, in the documentation of various Wasm runtimes, software engineer's blogs etc.) about this task is vague
and/or uncomplete. For example, some recipies could be found for Rust or JavaScript, but not all of them are applicable
to Go (on which we focus here). In other cases somewhere we may find an example of how to pass into the Wasm module a
string (or array data) but there are no good examples about how to return similar string out from Wasm. Also, there are
some examples that illustrate the principles but miss the proper memory management which makes such examples useless and
not production ready.

In this article we will walk through the solution of the task described above. We cannot cover all the diversity of
languages and Wasm runtimes, so focus just on the following. We will write our guest application in Go, compile it to
Wasm with [TinyGo](https://tinygo.org/docs/guides/webassembly/) compiler and embed it with
[Wasmtime](https://github.com/bytecodealliance/wasmtime-go) runtime into the host application which will be written also
in Go. For serialization we will use [Karmem](https://github.com/inkeliz/karmem) [^2] which is a format and a library.


## API of the guest application
Our guest application will accept complex objects of `DataRequest` type, which in Karmem language could be described as
this:
```
struct DataRequest inline {
    Numbers []int32;
    K int32;
}
```
It has array of integers `Numbers` and a number `K`. Our guest application will do very simple following business logic: return only those numbers which are greater than the given `K` number. So, our guest application will return objects of
`DataResponse` type:
```
struct DataResponse inline {
    NumbersGreaterK []int32;
}
```

These datatype definitions are located in the [api.km](https://github.com/vlkv/hackernoon_article_1/blob/master/api/api.km) file. We need to call Karmem code generator with a command:
```sh
hackernoon_article_1/api$ go run karmem.org/cmd/karmem build --golang -o "v1" api.km
```

This command generates for us a file `api/v1/api_generated.go` which contains Go code for serialization and
deserialization of `DataRequest` and `DataResponse` struct types. Karmem has very intuitive API, for example, here is a piece of code that creates a `DataRequest` and serializes it to `[]byte`:
```go
import "api/v1"

req := v1.DataRequest{
    Numbers: []int32{10, 43, 13, 24, 56, 16},
    K: 42,
}
writer := karmem.NewWriter(4 * 1024)
if _, err := req.WriteAsRoot(writer); err != nil {
	panic(err)
}
reqBytes := writer.Bytes()
```

Deserialization could be accomplished in a similar (mirrored) manner:
```go
import "api/v1"

reader := karmem.NewReader(reqBytes)
req := new(v1.DataRequest)
req.ReadAsRoot(reader)
```

Now we are able to convert our requests and responses to and from arrays of bytes. Let's proceed to the next steps.


## Memory management on the guest side

Before we begin to directly pass the `[]byte` data to the Wasm module, let's look at some details of memory management
of our guest application. According to the description of our general approach, we need to
- allocate a buffer of guest's memory from the host side (need it to copy request's bytes there),
- allocate a buffer of guest's memory from the guest side (need it to copy response's bytes there),
- deallocate (free) previously allocated memory buffer from the host side (both buffers will be deallocated from the
 host side).

Thus all we need here is a pair of functions: `Malloc` and `Free` [^3] (which are very similar to those used in the C
language) exported by the guest application. Here is the `Malloc` function:
```go
var allocatedBytes = map[uintptr][]byte{}

//go:export Malloc
func Malloc(size uint32) uintptr {
	buf := make([]byte, size)
	ptr := &buf[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	allocatedBytes[unsafePtr] = buf
	return unsafePtr
}
```

Comment `//go:export Malloc` is not just a comment but a TinyGo way to mark the functions that should be exported out
from the resulting Wasm module. The `allocatedBytes` map holds all the references to all allocated memory buffers so GC
will not come and collect them. The only non-trivial part here could be this 'magic': `unsafePtr :=
uintptr(unsafe.Pointer(ptr))`. This is simply a way in Go to get raw (and thus unsafe) pointer to some object. We need
raw pointer because we should treat it like an integer number so we able to pass it to (and from) the Wasm.

Implementation of the `Free` function is trivial, it simply deletes references to previously allocated buffers from the `allocatedBytes` map:
```go
//go:export Free
func Free(ptr uintptr) {
	delete(allocatedBytes, ptr)
}
```

Also we have a helper function to access the allocated memory buffers from the guest side:
```go
func getBytes(ptr uintptr) []byte {
	return allocatedBytes[ptr]
}
```

See the [mem.go](https://github.com/vlkv/hackernoon_article_1/blob/master/guest/mem.go) for complete source code of the
memory management module of the guest application.

All this memory management code along with the rest of the guest application code will be compiled into Wasm
instructions with TinyGo compiler. The exact command will be presented a little bit later in this article.


## Prepare request and pass it from host to guest

All the required preparations are done so in this paragraph we are ready to see the details of the host application
code. We skip details of the Wasm runtime initialization, because this is not the main focus of the article. The
initialization is encapsulated in the function `newWasmInstance` and we call it in the very begining:
```go
instance, store, mem, err := newWasmInstance("../guest/guest.wasm")
if err != nil {
    panic(err)
}
```

The `newWasmInstance` does all the initialization needed according to the Wasmtime [Getting started documentation](https://docs.wasmtime.dev/lang-go.html#getting-started-and-simple-example) and returns references to the Wasm VM instance as well as to it's store and linear memory.

Next, we get references to the three exported guest function objects that we will need:
```go
malloc := instance.GetFunc(store, "Malloc")
free := instance.GetFunc(store, "Free")
processRequest := instance.GetFunc(store, "ProcessRequest")
```
They are the memory management `Malloc` and `Free` functions (those discussed in the previous paragraph) and the
`ProcessRequest` function which is the guest's function which implements the guest's API. Conceptually `ProcessRequest`
accepts an instance of a `DataRequest` struct type and returns an instance of a `DataResponse` struct type. But in fact,
it accepts two 32-bit integers and returns one 64-bit integer. Here is it's signature as declared in the guest's sources:
```go
//go:export ProcessRequest
func ProcessRequest(reqPtr uintptr, reqLen uint32) uint64 {
  // ...
}
```
The two integers that `ProcessRequest` function accepts are:
- `reqPtr` is the address to the begining of the memory buffer where the serialized `DataRequest` bytes are copied to
- `reqLen` is the size of that buffer.

The resulting 64-bit integer holds bit representation of the two 32-bit integers that represent the address of the
buffer and it's size where serialized bytes of `DataResponse` are copied to. The reason why it is a single 64-bit
integer instead of a tuple of two 32-bit integers is that it is super unclear how TinyGo treats data when function
return complex tuple-like result. It was not documented at the moment of the writing this article (or simply I could not
find it). So this should be considered as a workaround hack (that works wery well) to return a pair of 32-bit integers
from a function that is exported from a Wasm module.

```go
// Here `reqBytes` is a []byte array with the DataRequest serialized bytes
reqBytesLen := int32(len(reqBytes))
ptrReq, err := malloc.Call(store, reqBytesLen)
if err != nil {
    panic(err)
}

int32PtrReq := ptrReq.(int32)
copy(
    mem.UnsafeData(store)[int32PtrReq : int32PtrReq+reqBytesLen],
    reqBytes,
)

respPtrLen, err := processRequest.Call(store, int32PtrReq, reqBytesLen)
if err != nil {
    panic(err)
}

free.Call(store, int32PtrReq)

// `respPtrLen` points to the `DataResponse`, we will use it in the next steps. TO BE CONTINUED...
```
Here we call `Malloc` function that allocates the memory buffer of the exact size to fit the `reqBytes` data, copy that
bytes data to the buffer and call the `ProcessRequest` function. Right after that we call `Free` on the allocated memory
buffer.

There is a lot of type casting happen here and this deserves a bit of explanation. For example, `ProcessRequest` function as you may have noticed accepts two 32-bit integers of unsigned types: `uintptr` and `uint32`. But Wasm supports only signed int32 types. You may see it if you inspect (for example with `wasmer inspect guest.wasm` command, more info [here](https://docs.wasmer.io/ecosystem/wasmer/usage#wasmer-inspect) about their CLI tool) the Wasm module:
```
Type: wasm
Size: 86.3 KB
<...>
Exports:
  Functions:
    <...>
    "ProcessRequest": [I32, I32] -> [I64]
<...>
```

That is why we have to do all the type casting on the host side explicitly, it is not performed automatically because Go
does not allow this.

Here is the full source code of the host's
[main.go](https://github.com/vlkv/hackernoon_article_1/blob/master/host/main.go) module.


## The rest of the guest application implementation

So far so good, we have sent the address to the `DataRequest` serialized bytes and number of those bytes to the guest.
Let's see how guest handles this data:

```go
//go:export ProcessRequest
func ProcessRequest(reqPtr uintptr, reqLen uint32) uint64 {
    reader := karmem.NewReader(getBytes(reqPtr))
    req := new(v1.DataRequest)
    req.ReadAsRoot(reader)

    resp := doProcessRequest(req)

    writer := karmem.NewWriter(4 * 1024)
    if _, err := resp.WriteAsRoot(writer); err != nil {
        panic(err)
    }
    respBytes := writer.Bytes()
    respBytesLen := uint32(len(respBytes))
    ptrResp := Malloc(respBytesLen)
    respBuf := getBytes(ptrResp)
    copy(respBuf, respBytes)
    return packPtrAndSize(ptrResp, respBytesLen) // NOTE: It is the host's responsibility to free this memory!
}
```

First, guest asks it's memory management "subsystem" to find a corresponding `[]bytes` array by calling
`getBytes(reqPtr)` function. Second, Karmem library is used to deserialize the `DataRequest` struct instance and as a
result the `req` variable references to it. After that the `doProcessRequest` function is called which contains all the
"business logic" of our guest application. Here it is (it is trivial, but the main point here is that it produces an
instance of a `DataResponse` struct):
```go
func doProcessRequest(req *v1.DataRequest) *v1.DataResponse {
    result := make([]int32, 0)
    for _, number := range req.Numbers {
        if number > req.K {
            result = append(result, number)
        }
    }
    resp := v1.DataResponse{
        NumbersGreaterK: result,
    }
    return &resp
}
```

In the end of the `ProcessRequest` function some interesting things happen. We take the `respBytes`, call the `Malloc` function to allocate the memory buffer (which is Wasm's linear memory, but host also is perfectly able to read from it) and copy the `DataResponse` serialized (with Karmem) bytes data into that buffer. Then, we call the `packPtrAndSize` function that "joins" two 32-bit integers into one 64-bit integer and return it to the host.

The implementation of the `packPtrAndSize` has some bit manipulation:
```go
func packPtrAndSize(ptr uintptr, size uint32) (ptrAndSize uint64) {
    return uint64(ptr)<<uint64(32) | uint64(size)
}
```

It takes a 64-bit integer and writes the 32-bits of the `ptr` variable to the high bits of it. It also writes the
32-bits of the `size` variable to the low bits of the 64-bit integer. Then returns the 64-bit integer as a result. The full source code of the guest's `main.go` module could be found [here](https://github.com/vlkv/hackernoon_article_1/blob/master/guest/main.go).

One very important thing in this part of the article is that the memory buffer that was **allocated by the guest** for
the `DataResponse` result should be **deallocated by the host** at some point in time when it will be not needed
anymore. How this is implemented in code will be seen right in the next paragraph. The reasoning behind this way of
managing memory is that Wasm module (compiled by TinyGo) has it's own GC onboard. If we simply return the address of the
byte buffer that returned Karmem to us (the `respBytes` variable), there is no guarantee that GC will not collect this
memory after the `ProcessRequest` function completes. And this could be fatal if garbage collection of that memory will
happen before the host will use the response.

After all the sources of the guest application is in their place, we can call the TinyGo compiler and build the Wasm module with a command:
```sh
hackernoon_article_1/guest$ tinygo build -target=wasi -o guest.wasm .
```

This should produce a `hackernoon_article_1/guest/guest.wasm` binary file which is used by the host application.


## Accept the result on the host side

Let's look what happens on the host side right after the `ProcessRequest` function call:

```go
// `respPtrLen` points to the `DataResponse`, we will use it in the next steps. TO BE CONTINUED...
respPtr, respLen := unpackPtrAndSize(uint64(respPtrLen.(int64)))

resp := new(v1.DataResponse)
respBytes := mem.UnsafeData(store)[int32(respPtr) : int32(respPtr)+int32(respLen)]
resp.ReadAsRoot(karmem.NewReader(respBytes))

fmt.Printf("NumbersGreaterK=%v\n", resp.NumbersGreaterK)

free.Call(store, respPtr) // This memory was allocated on the guest side, we free it on the host side here
```

The resulting 64-bit variable `respPtrLen` is being "unpacked" into two 32-bit integers which are the address of the
memory buffer and the size of that buffer. The `unpackPtrAndSize` function that does the opposite thing that was made by `packPtrAndSize`:
```go
func unpackPtrAndSize(ptrSize uint64) (ptr uintptr, size uint32) {
    ptr = uintptr(ptrSize >> 32)
    size = uint32(ptrSize)
    return
}
```

Tnen, we call Karmem to deserialize the bytes in the resulting buffer into the `DataResponse` struct instance `resp` and
use the data from there (in our example this usage is a simple printing of the `resp.NumbersGreaterK` to the standard
output). After we used the resulting data we call `Free` on the `respPtr` because it is the host's responsibility to
free that memory buffer which was allocated by the guest code on the guest side. End of story!


## How to run the complete example
The complete sources of the example discussed in this article are in my GitHub [repo](https://github.com/vlkv/hackernoon_article_1/tree/master). What you need to do as a prerequisite is:
- Install the latest Go (I used `go version go1.19.5 linux/amd64`) according to their official [instructions](https://go.dev/doc/install).
- Install the latest TinyGo (I used `tinygo version 0.27.0 linux/amd64 (using go version go1.19.5 and LLVM version
  15.0.0)`) according to their official [instructions](https://tinygo.org/getting-started/install/).
- Make sure you have both `go` and `tinygo` executables in your `PATH`.

I have created a bunch of `Makefile`s in the git repo, so theoretically all you have to do is to execute:
```sh
hackernoon_article_1$ make
```

Which should do the following:
- generate the Karmem serialization/deserialization code from the [api.km](https://github.com/vlkv/hackernoon_article_1/blob/master/api/api.km) definitions
- call the TinyGo compiler and build the guest.wasm Wasm module
- run the host Go application

Here is the output on my machine:
```
vitvlkv@littlepea:~/hackernoon_article_1$ make
cd ./host && make
make[1]: Entering directory '/home/vitvlkv/hackernoon_article_1/host'
cd ../api && make
make[2]: Entering directory '/home/vitvlkv/hackernoon_article_1/api'
go run karmem.org/cmd/karmem build --golang -o "v1" api.km
make[2]: Leaving directory '/home/vitvlkv/hackernoon_article_1/api'
cd ../guest && make
make[2]: Entering directory '/home/vitvlkv/hackernoon_article_1/guest'
cd ../api && make
make[3]: Entering directory '/home/vitvlkv/hackernoon_article_1/api'
make[3]: Nothing to be done for 'all'.
make[3]: Leaving directory '/home/vitvlkv/hackernoon_article_1/api'
tinygo build -target=wasi -o guest.wasm .
make[2]: Leaving directory '/home/vitvlkv/hackernoon_article_1/guest'
WASMTIME_BACKTRACE_DETAILS=1 go run .
Numbers=[10 43 13 24 56 16], K=42
NumbersGreaterK=[43 56]
make[1]: Leaving directory '/home/vitvlkv/hackernoon_article_1/host'
```

As you can see, we passed to the guest array of `Numbers=[10 43 13 24 56 16]` and integer `K=42` and received back array
`NumbersGreaterK=[43 56]` with two integers which are larger than the given `K` in the input `Numbers` array. This is
exactly what we wanted our guest application to do!


## Conclusion

This article proposes a solution to one common problem which WebAssembly's software engineers face very often: passing
non-primitive datatypes to and from Wasm module. WebAssembly has a steep learning curve but it also has a great
potential as a technology. I hope this article will help somebody in their work or personal project.


## Footnotes

[^1] If we'd write our guest application in Rust, then the whole task could be solved in a much simpler manner, using
their [wasm-bindgen](https://rustwasm.github.io/docs/wasm-bindgen/) library.

[^2] JSON format could be used too, but it should be noted that Go's standard `encoding/json` library doesn't work in
TinyGo, because TinyGo does not support reflection. People who need to use JSON without schema in Wasm usually go with
[gson](https://github.com/tidwall/gjson) library. [tinyjson](https://github.com/CosmWasm/tinyjson) seems to be a good
alternative for cases where the schema of all JSON messages is known. For this article I was looking for something like [Protobuf](https://protobuf.dev) but unfortunately, their Go's implementation does not work with TinyGo. Karmem is very close to Protobuf conceptually, that is why I decided to use it.

[^3] When TinyGo compiles Go sources into Wasm module it automatically adds some own implementations of `malloc` and
`free` functions to the exports list (they could be observed if you inspect the Wasm module with some corresponding
tool, like [wasmer inspect](https://docs.wasmer.io/ecosystem/wasmer/usage#wasmer-inspect)). We do not use them by two
reasons: 1) TinyGo promises nothing about stability of exported `malloc` and `free` 2) we have to call `Malloc` for
storing the result somehow from the guest side, it is unclear how to call standard `malloc` this way but very straightforward
with our own custom `Malloc`.
