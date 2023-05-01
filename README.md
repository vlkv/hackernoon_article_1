# Passing and returning arrays to and from Go WebAssembly module

[WebAssembly](https://webassembly.org/) is a great technology and has a lot of nice features: it is multiplatform, it
can work in browser and on server side, etc. But (maybe) due to the fact that WebAssembly is rather young, some basic
tasks are not as easy as they expected to be especially for the newcomers. One of such surprisingly difficult tasks is
passing and returning to and from WebAssembley module any complex objects, because it is well known, that Wasm supports
only primitive datatypes (int32, int64, float32 and float64). In fact, passing any complex objects like arrays, strings,
structs with named fields all could be reduced to one single problem of passing arrays of bytes and applying some
serialization/deserialization algorithm to the data.

The general approach to accomplish this is intuitively quite simple - allocate memory, pass pointer to that memory +
buffer size to Wasm module, return similar pair from Wasm module and don't forget to correctly free all previously
allocated memory. When we start thinking about memory management in WebAssembly, it is very dependent on what language
[^1] was used for subsequent compilation to Wasm instructions. Some languages have garbage collector (GC) "out of the
box", some have not. It is also not a trivial task to choose serialization format and library for interpreting the
passed bytes. In the end, when it comes to writing the exact code that correctly works and could be safely deployed to
production... things may become a little bit complicated at least.

This problem is exacerbated by the fact that the information that could be found in the Internet (on official Wasm
resources, in the documentation of various Wasm runtimes, software engineer's blogs etc.) about this task is vague
and/or uncomplete. For example, some recipies could be found for Rust or JavaScript, but not all of them are applicable
to Go (on which we focus here). In other cases somewhere we may find an example of how to pass into the Wasm module a
string (or array data) but there are no good examples about how to return similar string out from Wasm. Also, there are
some examples that illustrate the principles but miss the proper memory management which makes such examples useless and
not production ready.

In this article we will walk through the solution of the task described above. We cannot cover all the diversity of
languages and Wasm runtimes, so focus on just a few. We will write our guest application in Go, compile it to Wasm with
[TinyGo](https://tinygo.org/docs/guides/webassembly/) compiler and embed it with
[Wasmtime](https://github.com/bytecodealliance/wasmtime-go) runtime into the host application which will be written also
in Go. For serialization we will use [Karmem](https://github.com/inkeliz/karmem) [^2] which is a format and a library.

[^1] If we'd write our guest application in Rust, then the whole task could be solved in a much simpler manner, using
their [wasm-bindgen](https://rustwasm.github.io/docs/wasm-bindgen/) library.

[^2] JSON format could be used too, but it should be noted that Go's standard `encoding/json` library doesn't work in
TinyGo, because TinyGo does not support reflection. People who need to use JSON without schema in Wasm usually go with
[gson](https://github.com/tidwall/gjson) library. [tinyjson](https://github.com/CosmWasm/tinyjson) seems to be a good
alternative for cases where the schema of all JSON messages is known. For this article I was looking for something like [Protobuf](https://protobuf.dev) but unfortunately, their Go's implementation does not work with TinyGo. Karmem is very close to Protobuf conceptually, that is why I decided to use it.

## Schema of the guest's API
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
  req := v1.DataRequest{
		Numbers: []int32{10, 43, 13, 24, 56, 16},
		K: 42,
	}
	writer := karmem.NewWriter(20 * 1024)
	if _, err := req.WriteAsRoot(writer); err != nil {
		panic(err)
	}
	reqBytes := writer.Bytes()
```

Deserialization could be accomplished in a similar (mirrored) manner.


## GENERAL APPROACH IN DETAIL
* PART I: HOW TO PASS THE DATA IN. Create instance of DataRequest type, serialize it into array of bytes. Call guest's
  `Malloc` function to allocate exact number of bytes. Copy bytes form host's array to just allocated guest's array.
  Take address of the guest's array. Pass to `ProcessRequest` two numbers: address of the array and length of the array.
  After the `ProcessRequest` returns result, call `Free` on the guest's byte array with serialized request data.

* PART II: On the guest side. Take the address and length and deserialize the request into the `DataRequest` object.
  Then process request somehow (according to some business logic of guest application) and construct the `DataResponse`
  object.

* PART III: HOW TO PASS THE DATA OUT. To send the `DataResponse` object back to host we should first serialize it to
  array of bytes. Then, the first intention could be to return to the host the address of this array + it's length. But
  this would be bad idea. Because this byte array is managed by guest's GC, so it could be deallocated at any time. That
  is why, instead, we should on the guest side call `Malloc` to allocate additional byte array and copy all the data
  there. Then it would be safe to return to the host the address of that second array and it's length. But this is not
  the end of the story, we have two issues here:
  - If we return a tuple of two integers from `ProcessRequest` then, after compilation to Wasm the signature looks very
    weird. Obviously this is some kind of a mechanism which TinyGo/Wasm uses to return tuples. But I could not find out
    how to properly use that mechanism. So, instead of returning two int32 numbers, we return single int64 with
    32 higher bits containing the address of the resulting byte array and 32 lower bits - containing the length of that
    array
  - After we deserialize `ProcessRequest` result we should deallocate the memory which is occupied by the resulting
    byte array. Because nor guest's nor host's GC will not collect it. So it is the hosts's responsibility to call
    `Free` on the resulting byte array.

* As could be seen we additionally needed:
  - Own implementation of `Malloc` and `Free` functions on the guest side.
  - Some helper functions for packing/unpacking int64 to pair of int32 and vice versa.

## WALK THROUGH THE CODE STEP BY STEP

### Guest's Malloc and Free implementations

### Guest's API with Karmem

### Pass the RequestData into Wasm

### Deserialize and process the DataRequest

### Pass the ResponseData out from Wasm

## NOTES ABOUT HOW TO RUN THE EXAMPLE
See the complete sources in my GitHub repo. Install TinyGo, Go and make. Run `make` in the repo root, this should build
and run everything. See results in the console. Cool!

## CONCLUSION
It is not that hard, but has some difficulties. Every detail was clearly explained. I hope this article will help
somebody.
