package examples

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yoshd/protoc-gen-stest/examples/pb"

	"google.golang.org/grpc"
)

var testHandlerMap = map[string]*func(t *testing.T, expectedResponse, response interface{}){}

func TestScenario(t *testing.T) {
	target := "localhost:13009"
	client, err := grpc.Dial(target, grpc.WithInsecure())
	assert.NoError(t, err)
	defer client.Close()
	sampleClient := pb.NewSampleClient(client)
	testClient := pb.NewTestClient(sampleClient)
	testClient.RunGRPCTest(
		t,
		"scenario/sample.json",
		testHandlerMap,
	)
}

func setUp() {
	helloTestHandler := func(t *testing.T, expectedResponse, response interface{}) {
		if expectedResponse == nil || response == nil {
			return
		}
		er := expectedResponse.(pb.HelloResponse)
		r := response.(pb.HelloResponse)
		assert.Equal(t, er.ResMsg, r.ResMsg)
	}
	testHandlerMap["Hello"] = &helloTestHandler
	byeTestHandler := func(t *testing.T, expectedResponse, response interface{}) {
		if expectedResponse == nil || response == nil {
			return
		}
		er := expectedResponse.(pb.ByeResponse)
		r := response.(pb.ByeResponse)
		assert.Equal(t, er.ResMsg, r.ResMsg)
	}
	testHandlerMap["Bye"] = &byeTestHandler
}

func TestMain(m *testing.M) {
	setUp()
	m.Run()
}
