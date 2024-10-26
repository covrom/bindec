# bindec [![Build Status](https://travis-ci.org/covrom/bindec.svg?branch=master)](https://travis-ci.org/covrom/bindec) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Go Report Card](https://goreportcard.com/badge/github.com/covrom/bindec)](https://goreportcard.com/report/github.com/covrom/bindec)

`bindec` generates encoders and decoders to encode and decode binary representations of types. Encoders and decoders are code-generated, thus generating code tailored specifically for your type, making encoding and decoding really fast.

### Install

```
go install github.com/covrom/bindec/cmd/bindec@latest
```

### Usage

```
bindec -type=MyType
```

Since the default receiver is `t` and that might cause a lint error if your type has other methods with a different receiver, you can use the `-recv` argument.

```
bindec -recv=mt -type=MyType
```

You can also generate a encoders and decoders for a type from a package that is not your working directory by simply passing the path as an argument after all the other flags.

```
bindec -recv=somerecv -type=SomeType /path/to/package
```

If you want to generate encoders and decoders for more than one file and output them in the same file, you can do so passing comma-separated types.

```
bindec -type=FirstType,SecondType,ThirdType
```

If you give them receiver names, you must give them in the same order as the types are passed, also comma-separated.

```
bindec -recv=a,b,c -type=A,B,C
```

### Encode and decode

After generating the code you will have in your package a file `yourtype_bindec.go` with four methods added to the type: `EncodeBinary`, `WriteBinary`, `DecodeBinaryFromBytes` and `DecodeBinary`.

```go
// This is our type.
type Person {
    Name string
    Age int
    Gender string
}

// EncodeBinary and DecodeBinary will be on person_bindec.go

// We get some data from somewhere as []byte
data := getSomeDataFromSomewhere()

var p Person
// Then we decode the person from the data.
if err := p.DecodeBinaryFromBytes(data); err != nil {
    // handle err
}

// We can even read it from a reader.
var p2 Person
if err := p.DecodeBinary(reader); err != nil {
    // handle err
}

// And encode it again.
encoded, err := p.EncodeBinary()
if err != nil {
    // handle err
}

// Or write it to a writer.
writer := bytes.NewBuffer(nil)
if err := p.WriteBinary(writer); err != nil {
    // handle err
}
```

### Ignore fields

You may have fields you don't want to encode or decode. You can do so using `bindec:"-"` struct tag.

```go
type User struct {
    Username string
    Email string
    Password string `bindec:"-"`
}
```

This way, only `Username` and `Email` will be encoded and decoded, but not `Password`.

### Benchmarks

Speed:

```
goos: darwin
goarch: amd64
pkg: github.com/covrom/bindec/bench
BenchmarkEncode/bindec-4                  500000              2274 ns/op     576 B/op           6 allocs/op
BenchmarkEncode/gob-4                     100000             12464 ns/op    2432 B/op          41 allocs/op
BenchmarkDecode/bindec-4                  500000              2590 ns/op     528 B/op          22 allocs/op
BenchmarkDecode/gob-4                      50000             37126 ns/op    9572 B/op         249 allocs/op
PASS
ok      github.com/covrom/bindec/bench    6.137s
```

Size:

```
BINDEC: 112 bytes
GOB: 205 bytes
```

Benchmarked against `encoding/gob` on a MacBook Pro (Retina, 13-inch, Early 2015) (2,7 GHz Intel Core i5) on macOS Mojave 10.14.3.

### Why not just `encoding/gob`?

It's slower, takes more space and requires registering every single type that's going to be encoded.
`bindec` offers a fast way to convert a struct into bytes with the smallest possible size. All you need to do is add `//go:generate` tags to you code and run `go generate ./...`.

In a future version, `bindec` will support adding validations to the fields during decoding via struct tags, which will allow things like "accept only 100 bytes on this field so an atacker can't send as many bytes as they want" to be performed automatically.

### Supported types

- Integers: byte, int, uint, uint64, ...
- Floats: float32, float64
- Strings
- Maps with keys and values of supported types
- Pointers
- Structs (without cyclic references) with fields of supported types
- Arrays of supported types
- Slices of supported types
- Booleans

### Limitations

- Interface, function and channel types are not supported. The reason interfaces are not supported is because they can be anything, potentially even from any package, on runtime and bindec decoders and encoders are generated beforehand.
- Cyclic structures are not supported.

### Specification

For more details about the format used to encode the types, see [SPEC.md](/SPEC.md).

### Constraints

TODO.

### LICENSE

MIT License, see [LICENSE](/LICENSE)
