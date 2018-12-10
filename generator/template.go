package generator

var codeTemplate = `
package {{.Package}}

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// {{.GRPCServiceName}}TestRunner is a runner to run the {{.GRPCServiceName}} service test.
type {{.GRPCServiceName}}TestRunner struct {
	Client {{.GRPCServiceName}}Client
}

// NewTestClient returns new {{.GRPCServiceName}}Runner.
func NewTestClient(client {{.GRPCServiceName}}Client) *{{.GRPCServiceName}}TestRunner {
	return &{{.GRPCServiceName}}TestRunner{
		Client: client,
	}
}

// RunGRPCTest sends a gPRC request according to the scenario written in the JSON file and tests the response.
// testHandlerMap takes a gRPC method name as a key and value has a function that compares expected response and actual response and defines how to handle the test.
func (runner *{{.GRPCServiceName}}TestRunner) RunGRPCTest(t *testing.T, jsonPath string, testHandlerMap map[string]*func(t *testing.T, expectedResponse, response interface{})) {
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

func (runner *{{.GRPCServiceName}}TestRunner) runTest(ctx context.Context, t *testing.T, testCase map[string]interface{}, testHandlerMap map[string]*func(t *testing.T, expectedResponse, response interface{})) {
	action := testCase[actionJSONKey].(string)
	f := func(t *testing.T) {
		switch action {
		{{- range $i, $v := .GRPCMethods }}
		case "{{$v.Name}}":
			testHandler := testHandlerMap["{{$v.Name}}"]
			runner.test{{$v.Name}}(ctx, t, testCase, testHandler)
		{{- end }}
		}
	}
	t.Run(action, f)
}

{{- $GRPCServiceName := .GRPCServiceName }}
{{- $PackageName := .Package }}
{{ range $i, $v := .GRPCMethods }}
func (runner *{{$GRPCServiceName}}TestRunner) test{{$v.Name}}(ctx context.Context, t *testing.T, testCase map[string]interface{}, testHandler *func(t *testing.T, expectedResponse, response interface{})) {
	reqJSON, reqErr := json.Marshal(testCase[requestJSONKey])
	if reqErr != nil {
		panic(reqErr)
	}
	req := {{$v.RequestType}}{}
	json.Unmarshal(reqJSON, &req)
	res, err := runner.Client.{{$v.Name}}(ctx, &req)
	errExpectation := testCase[errorExpectationJSONKey].(bool)
	if errExpectation {
		errCodeF := testCase[expectedErrorCodeJSONKey].(float64)
		errCodeU := uint32(errCodeF)
		expectedErrCode := codes.Code(errCodeU)
		if expectedErrCode != grpc.Code(err) {
			t.Fatalf("The error code of the response of {{$v.Name}} is not as expected. Expected: %d, Actual: %d\n", expectedErrCode, grpc.Code(err))
		}
	} else {
		resJSON, resErr := json.Marshal(testCase[expectedResponseJSONKey])
		if resErr != nil {
			panic(resErr)
		}
		expectedRes := {{$v.ResponseType}}{}
		json.Unmarshal(resJSON, &expectedRes)
		if testHandler != nil {
			handler := *testHandler
			handler(t, expectedRes, *res)
		} else {
			if !reflect.DeepEqual(expectedRes, *res) {
				t.Fatal("The actual response of the {{$v.Name}} was not equal to the expected response.")
			}
		}
	}
}
{{ end }}
`
