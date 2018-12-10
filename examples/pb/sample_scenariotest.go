
package pb

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// SampleTestRunner is a runner to run the Sample service test.
type SampleTestRunner struct {
	Client SampleClient
}

// NewTestClient returns new SampleRunner.
func NewTestClient(client SampleClient) *SampleTestRunner {
	return &SampleTestRunner{
		Client: client,
	}
}

// RunGRPCTest sends a gPRC request according to the scenario written in the JSON file and tests the response.
// testHandlerMap takes a gRPC method name as a key and value has a function that compares expected response and actual response and defines how to handle the test.
func (runner *SampleTestRunner) RunGRPCTest(t *testing.T, jsonPath string, testHandlerMap map[string]*func(t *testing.T, expectedResponse, response interface{})) {
	scenarioData, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		panic(err)
	}
	var scenario []map[string]interface{}
	json.Unmarshal(scenarioData, &scenario)
	for _, testCase := range scenario {
		ctx := context.Background()
		runner.runTest(ctx, t, testCase, testHandlerMap)
	}
}

const (
	actionJSONKey            = "action"
	requestJSONKey           = "request"
	expectedResponseJSONKey  = "expected_response"
	errorExpectationJSONKey  = "error_expectation"
	expectedErrorCodeJSONKey = "expected_error_code"
	skipMessage              = "Skip comparing expected response and actual response"
)

func (runner *SampleTestRunner) runTest(ctx context.Context, t *testing.T, testCase map[string]interface{}, testHandlerMap map[string]*func(t *testing.T, expectedResponse, response interface{})) {
	action := testCase[actionJSONKey].(string)
	f := func(t *testing.T) {
		switch action {
		case "Hello":
			testHandler := testHandlerMap["Hello"]
			runner.testHello(ctx, t, testCase, testHandler)
		case "Bye":
			testHandler := testHandlerMap["Bye"]
			runner.testBye(ctx, t, testCase, testHandler)
		}
	}
	t.Run(action, f)
}

func (runner *SampleTestRunner) testHello(ctx context.Context, t *testing.T, testCase map[string]interface{}, testHandler *func(t *testing.T, expectedResponse, response interface{})) {
	reqJSON, reqErr := json.Marshal(testCase[requestJSONKey])
	if reqErr != nil {
		panic(reqErr)
	}
	req := HelloRequest{}
	json.Unmarshal(reqJSON, &req)
	res, err := runner.Client.Hello(ctx, &req)
	errExpectation := testCase[errorExpectationJSONKey].(bool)
	if errExpectation {
		errCodeF := testCase[expectedErrorCodeJSONKey].(float64)
		errCodeU := uint32(errCodeF)
		expectedErrCode := codes.Code(errCodeU)
		if expectedErrCode != grpc.Code(err) {
			t.Fatalf("The error code of the response of Hello is not as expected. Expected: %d, Actual: %d\n", expectedErrCode, grpc.Code(err))
		}
	} else {
		resJSON, resErr := json.Marshal(testCase[expectedResponseJSONKey])
		if resErr != nil {
			panic(resErr)
		}
		expectedRes := HelloResponse{}
		json.Unmarshal(resJSON, &expectedRes)
		if testHandler != nil {
			handler := *testHandler
			handler(t, expectedRes, *res)
		} else {
			if !reflect.DeepEqual(expectedRes, *res) {
				t.Fatal("The actual response of the Hello was not equal to the expected response.")
			}
		}
	}
}

func (runner *SampleTestRunner) testBye(ctx context.Context, t *testing.T, testCase map[string]interface{}, testHandler *func(t *testing.T, expectedResponse, response interface{})) {
	reqJSON, reqErr := json.Marshal(testCase[requestJSONKey])
	if reqErr != nil {
		panic(reqErr)
	}
	req := ByeRequest{}
	json.Unmarshal(reqJSON, &req)
	res, err := runner.Client.Bye(ctx, &req)
	errExpectation := testCase[errorExpectationJSONKey].(bool)
	if errExpectation {
		errCodeF := testCase[expectedErrorCodeJSONKey].(float64)
		errCodeU := uint32(errCodeF)
		expectedErrCode := codes.Code(errCodeU)
		if expectedErrCode != grpc.Code(err) {
			t.Fatalf("The error code of the response of Bye is not as expected. Expected: %d, Actual: %d\n", expectedErrCode, grpc.Code(err))
		}
	} else {
		resJSON, resErr := json.Marshal(testCase[expectedResponseJSONKey])
		if resErr != nil {
			panic(resErr)
		}
		expectedRes := ByeResponse{}
		json.Unmarshal(resJSON, &expectedRes)
		if testHandler != nil {
			handler := *testHandler
			handler(t, expectedRes, *res)
		} else {
			if !reflect.DeepEqual(expectedRes, *res) {
				t.Fatal("The actual response of the Bye was not equal to the expected response.")
			}
		}
	}
}

