![test](https://github.com/mm-technologies/protoc-gen-stest/workflows/Go/badge.svg)

# protoc-gen-stest

This is a protoc plugin which generates golang source code for gRPC scenario test (only Unary).
The plugin can test the gRPC methods defined in your .proto file.
The necessary preparation is the source code that calls the test using your .proto file and the JSON file that defines the test scenario, and the simple gRPC service client and testing package.

To use this plugin, you need to use [protoc-gen-go](https://github.com/golang/protobuf/tree/master/protoc-gen-go) to generate Golang source code.

# Installation

```
go get -u github.com/yoshd/protoc-gen-stest
```


# Invoking the Plugin

```
protoc -I. --plugin=path/to/protoc-gen-stest --stest_out=. your.proto
```

# Usage

## the simple example

See [examples](examples/)

* Suppose you have the following yoshd.proto:

```protobuf
syntax = "proto3";

option go_package = "pb";

service Yoshd {
    rpc Yoshi (YoshiRequest) returns (YoshiResponse) {
    }
}

message YoshiRequest {
    string req_msg = 1;
}
message YoshiResponse {
    string res_msg = 1;
}
```

* Generate source codes.

```
protoc -I. --plugin=path/to/protoc-gen-stest --go_out=plugins=grpc:pb --stest_out=pb your.proto
```

* The fields of JSON are as follows.
    * For `action` , write gRPC method name.
    * For `request` , write request parameters.
    * For `expected_response` , write the value of the expected response. If you expect error response, you do not need to write it.
    * For `loop` , specify the number of times to repeat the request. Default `1`
    * For `success_rule` , specify the rule for considering the test as successful. There are two kinds of rules as follows.　Default `all`
        * `all` : All the responses in the `loop` must be responses as expected.
        * `once` : If the response is as expected even once in the `loop` , the test is regarded as successful.
    * For `sleep` , specify the number of seconds to sleep before sending the request. Default `0`
    * For `error_expectation` , write whether or not to expect an error response. Default `false`
    * `For expected_error_code` , write the expected gPRC error code as a numerical value.

The field names of the request and response are the same as those of the JSON tag attached to the structure of the code generated by [protoc-gen-go](https://github.com/golang/protobuf/tree/master/protoc-gen-go).

In this example, the first test will succeed if the expected response is returned at least once while looping `Yoshi` twice. The first test sleeps for 3 seconds each time before calling `Yoshi`.
In the second test, an error response is returned, and if the gRPC error code is 3 (InvalidArgument), the test succeeds.
Please refer to [codes](https://godoc.org/google.golang.org/grpc/codes) for the error code of gPRC.

```json
[
    {
        "action": "Yoshi",
        "request": {
            "req_msg": "Yoshi!"
        },
        "expected_response": {
            "res_msg": "YoshiYoshi!"
        },
        "loop": 2,
        "sleep": 3,
        "success_rule": "once"
    },
    {
        "action": "Yoshi",
        "request": {
            "req_msg": "Yoshi"
        },
        "error_expectation": true,
        "expected_error_code": 3
    }
]
```

* Write gRPC client, code to compare expected response and actual response, test call in Golang.
    * The default behavior is to compare expected response and actual response with `reflect.DeepEqual`

```go
package examples

import (
	"testing"

	"github.com/mm-technologies/protoc-gen-stest/examples/pb"

	"google.golang.org/grpc"
)

func TestScenario(t *testing.T) {
	target := "localhost:13009"
	client, _ := grpc.Dial(target, grpc.WithInsecure())
	defer client.Close()

	yoshd := pb.NewYoshdClient(client)
	testClient := pb.NewTestClient(yoshd)
	testClient.RunGRPCTest(
		t,
		"path/to/yoshd.json",
		nil,
	)
}
```

* If you want to specify how you want to compare the expected response to the actual response, you need the code on how to compare the responses. The function must accept the following arguments and return an error.
    * `func(expectedResponse, response interface{}) error`
        * Since it is `interface`, we need to cast it to the response type of each gPRC method and compare it.
    * In the `compareFuncMap` argument of `RunGRPCTest` , specify the gPRC method name in key and put the above function in value.

```go
package examples

import (
    "errors"
	"testing"

	"github.com/mm-technologies/protoc-gen-stest/examples/pb"

	"google.golang.org/grpc"
)

var compareFuncMap = map[string]*func(expectedResponse, response interface{}) error{}

func TestScenario(t *testing.T) {
	target := "localhost:13009"
	client, _ := grpc.Dial(target, grpc.WithInsecure())
	defer client.Close()
	yoshd := pb.NewYoshdClient(client)
	testClient := pb.NewTestClient(yoshd)
	testClient.RunGRPCTest(
		t,
		"path/to/yoshd.json",
		compareFuncMap,
	)
}

func setUp() {
    // How to compare the expected response with the actual response and return an error
	YoshiCompareFunc := func(expectedResponse, response interface{}) error {
		if expectedResponse == nil || response == nil {
			return nil
        }
        // Requires cast from interface
		er := expectedResponse.(pb.YoshiResponse)
		r := response.(pb.YoshiResponse)
		if er.ResMsg != r.ResMsg {
            return errors.New("The actual response of the Yoshi was not equal to the expected response")
        }
    }
    // Specify the gPRC method name as key and put a function to compare the response to value.
	compareFuncMap["Yoshi"] = &YoshiCompareFunc
}

func TestMain(m *testing.M) {
	setUp()
	m.Run()
}
```

* Run the test

```
go test -v yoshd_test.go
```
