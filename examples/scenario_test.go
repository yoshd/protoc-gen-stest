package examples

import (
	"errors"
	"testing"

	"github.com/yoshd/protoc-gen-stest/examples/pb"

	"google.golang.org/grpc"
)

var responseCompareFuncMap = map[string]*func(expectedResponse, response interface{}) error{}

func TestScenario(t *testing.T) {
	target := "localhost:13009"
	client, _ := grpc.Dial(target, grpc.WithInsecure())
	defer client.Close()
	sampleClient := pb.NewSampleClient(client)
	testClient := pb.NewTestClient(sampleClient)
	testClient.RunGRPCTest(
		t,
		"scenario/sample.json",
		responseCompareFuncMap,
	)
}

func setUp() {
	helloResponseCompareFunc := func(expectedResponse, response interface{}) error {
		if expectedResponse == nil || response == nil {
			return nil
		}
		er := expectedResponse.(pb.HelloResponse)
		r := response.(pb.HelloResponse)
		if er.ResMsg != r.ResMsg {
			return errors.New("the actual response of the Hello was not equal to the expected response")
		}
		return nil
	}
	responseCompareFuncMap["Hello"] = &helloResponseCompareFunc
	byeResponseCompareFunc := func(expectedResponse, response interface{}) error {
		if expectedResponse == nil || response == nil {
			return nil
		}
		er := expectedResponse.(pb.ByeResponse)
		r := response.(pb.ByeResponse)
		if er.ResMsg != r.ResMsg {
			return errors.New("the actual response of the Bye was not equal to the expected response")
		}
		return nil
	}
	responseCompareFuncMap["Bye"] = &byeResponseCompareFunc
}

func TestMain(m *testing.M) {
	setUp()
	m.Run()
}
