# Passing and returning arrays to and from Go WebAssembly module

## INTRO
WebAssembly is cool. It has a lot of cool features, it is multiplatform, it can work in browser and on servers, etc.

But! WebAssembly standard is a work in progress. At the moment only primitive datatypes (ints and floats) are supported.
In real life usecases we need to pass in and out to WASM module more complex types: strings, structured objects etc. The
general approach may seem to be quite simple - allocate memory, pass pointer to that memory + buffer size to WASM
module, return similar pair from WASM module and don't forget to correctly free the previously allocated memory. But
when it comes to writing the exact code that correctly works and could be safely deployed to production... things may
become a little bit difficult.

## THE PROBLEM THAT WE SOLVE HERE
Create WASM application which accepts some complex request object and returns some complex response object. We use
TinyGo to create this WASM app. Then we embed this WASM app into Go application using Wasmtime. The datatypes of request
and response complex objects could be anything (like any bytes array) but we going to use some library which provides
schema for the data. We'd love to use Protobuf but it doesn't work in TinyGo+WASM, so we found and use Karmem library
for the same purpose. In the end we should have a nice almost-ready-to-go-to-prod code example which illustrates all the
above.

## MOTIVATION
The information in official WASM resources are fuzzy and obscure (TODO: enumerate them). There are some recipies for
Rust (e.g. bindgen) or JavaScript, but not all of them are applicable to Go. Or somewhere we may find an example of how
to pass into the WASM module string (or array) data but there is no good examples about how to return similar string out
from WASM. Or, there are some examples that illustrate the principles but have errors in memory management which makes
such an examples useless and not production ready. Also, most of the examples focus on passing in/out strings, which is
cool. But in real life usecases we usually need something like string+JSON, or Protobuf or Flatbuffers or similar thing.
The problem with those libraries is that not much of them work in TinyGo/WASM, so finding a good practical solution
could really become a huge problem.

## TERMINOLOGY (git repo layout?)
* host - Go application which runs guest application as WASM embedded process using Wasmtime VM runtime.
* guest - Go application which is compiled to WASM binary with TinyGo. It exports one function `ProcessRequest` which
  accepts an instance of request and returns an instance of response.
* api - contains api.km file which is a Karmem schema for request and response types of `ProcessRequest` guest's
  function.

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
  - If we return a tuple of two integers from `ProcessRequest` then, after compilation to WASM the signature looks very
    weird. Obviously this is some kind of a mechanism which TinyGo/WASM uses to return tuples. But I could not find out
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

### Pass the RequestData into WASM

### Deserialize and process the DataRequest

### Pass the ResponseData out from WASM

## NOTES ABOUT HOW TO RUN THE EXAMPLE
See the complete sources in my GitHub repo. Install TinyGo, Go and make. Run `make` in the repo root, this should build
and run everything. See results in the console. Cool!

## CONCLUSION
It is not that hard, but has some difficulties. Every detail was clearly explained. I hope this article will help
somebody.
