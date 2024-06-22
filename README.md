# go-json-mask

go-json-mask is a simple, customizable Go library for masking JSON sensitive information.

- [go-json-mask](#go-json-mask)
  - [Installation](#installation)
  - [Json value types](#json-value-types)
  - [How to use](#how-to-use)
  - [Benchmarks](#benchmarks)

## Installation

```sh
go get github.com/bolom009/go-json-mask
```

## Json value types

go-json-mask support masking for all JSON value types, also users could create their own necessary masking functions.

| type    | masks        | description                                                                                                                      |
|:--------|:-------------|:---------------------------------------------------------------------------------------------------------------------------------|
| string  | hash, filled | hash - masks the string with sha1 <br/> filled - masks the string with the same number of masking characters or by passed length |
| int     | random int   | masks the integer value by default range (1000) or by passed                                                                     |
| float   | random float | masks the float value by default range (1000.3) or by passed, consists from two parts XXX.XXX                                    |
| array   | all types    | support (string, int, float, object, array)                                                                                      |
| boolean | -            | ignored                                                                                                                          |
| null    | -            | ignored                                                                                                                          |

## How to use

```go
package main

import (
	"fmt"
	"log"

	"github.com/bolom009/go-json-mask"
)

func main() {
	mask := jsonmask.NewJSONMask("key1", "/metadata/labels/key2", "/metadata/labels/key3[1]")
	mask.RegisterMaskStringFunc(jsonmask.MaskHashString())

	v := `{
      "name": "HelloWorld",
      "age": 999,
      "metadata": {
        "labels": {
          "key1": "value1",
          "key2": "value2",
          "key3": ["one", "two"]
        },
        "annotations": {
          "key1": "value1"
        }
      }
    }`

	res, err := mask.Mask(v)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println(res)
}
```

**Output**:
```
{"name":"HelloWorld","age":999,"metadata":{"labels":{"key1":"8107759ababcbfa34bcb02bc4309caf6354982ab","key2":"43f7aa390f1a0265fc2de7010133951c0718a67e", "key3":["one", "ad782ecdac770fc6eb9a62e44f90873fb97fb26b"]},"annotations":{"key1":"8107759ababcbfa34bcb02bc4309caf6354982ab"}}}
```


## Benchmarks
```
BenchmarkNewJSONMaskHashString-16    343420	      3341 ns/op	    1929 B/op	      47 allocs/op
BenchmarkNewJSONMaskFilledString-16  335380	      3125 ns/op	    1771 B/op	      46 allocs/op
BenchmarkNewJSONMaskInt-16    	     355957	      3167 ns/op	    1717 B/op	      44 allocs/op
BenchmarkNewJSONMaskFloat64-16       353574	      3215 ns/op	    1785 B/op	      46 allocs/op
```
