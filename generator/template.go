package generator

var codeTemplate = `
package {{.Package}}

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

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
// compareFuncMap takes a gRPC method name as a key and value has a function that compares expected response and actual response and return an error.
func (runner *{{.GRPCServiceName}}TestRunner) RunGRPCTest(t *testing.T, jsonPath string, compareFuncMap map[string]*func(expectedResponse, response interface{}) error) {
	scenarioData, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		panic(err)
	}
	var scenario []map[string]interface{}
	json.Unmarshal(scenarioData, &scenario)
	for _, testCase := range scenario {
		ctx := context.Background()
		runner.runTest(ctx, t, testCase, compareFuncMap)
	}
}

const (
	actionJSONKey            = "action"
	requestJSONKey           = "request"
	expectedResponseJSONKey  = "expected_response"
	errorExpectationJSONKey  = "error_expectation"
	expectedErrorCodeJSONKey = "expected_error_code"
	loopJSONKey              = "loop"
	sleepJSONKey             = "sleep"
	successRuleJSONKey       = "success_rule"
	successRuleAll           = "all"
	successRuleOnce          = "once"
)

func (runner *{{.GRPCServiceName}}TestRunner) runTest(ctx context.Context, t *testing.T, testCase map[string]interface{}, compareFuncMap map[string]*func(expectedResponse, response interface{}) error) {
	var action string
	if v, ok := testCase[actionJSONKey]; ok {
		action = v.(string)
	} else {
		panic("Scenario JSON is invalid. Because action is required.")
	}
	f := func(t *testing.T) {
		switch action {
		{{- range $i, $v := .GRPCMethods }}
		case "{{$v.Name}}":
			compareFunc := compareFuncMap["{{$v.Name}}"]
			runner.test{{$v.Name}}(ctx, t, testCase, compareFunc)
		{{- end }}
		}
	}
	t.Run(action, f)
}

{{- $GRPCServiceName := .GRPCServiceName }}
{{- $PackageName := .Package }}
{{ range $i, $v := .GRPCMethods }}
func (runner *{{$GRPCServiceName}}TestRunner) test{{$v.Name}}(ctx context.Context, t *testing.T, testCase map[string]interface{}, compareFunc *func(expectedResponse, response interface{}) error) {
	reqJSON, reqErr := json.Marshal(testCase[requestJSONKey])
	if reqErr != nil {
		panic(reqErr)
	}
	req := {{$v.RequestType}}{}
	json.Unmarshal(reqJSON, &req)

	loop := 1
	if v, ok := testCase[loopJSONKey]; ok {
		loop = int(v.(float64))
	}
FOR_LABEL:
	for i := 1; i <= loop; i++ {
		sleep := 0
		if v, ok := testCase[sleepJSONKey]; ok {
			sleep = int(v.(float64))
		}
		time.Sleep(time.Duration(sleep) * time.Second)

		res, err := runner.Client.{{$v.Name}}(ctx, &req)

		errExpectation := false
		if v, ok := testCase[errorExpectationJSONKey]; ok {
			errExpectation = v.(bool)
		}
		if errExpectation {
			errCodeF := testCase[expectedErrorCodeJSONKey].(float64)
			errCodeU := uint32(errCodeF)
			expectedErrCode := codes.Code(errCodeU)
			if expectedErrCode != grpc.Code(err) {
				t.Fatalf("The error code of the response of {{$v.Name}} is not as expected. Expected: %d, Actual: %d\n", expectedErrCode, grpc.Code(err))
			}
			break FOR_LABEL
		} else {
			resJSON, resErr := json.Marshal(testCase[expectedResponseJSONKey])
			if resErr != nil {
				panic(resErr)
			}
			expectedRes := {{$v.ResponseType}}{}
			json.Unmarshal(resJSON, &expectedRes)
			successRule := successRuleAll
			if v, ok := testCase[successRuleJSONKey]; ok {
				successRule = v.(string)
			}
			var err error
			if compareFunc != nil {
				compare := *compareFunc
				err = compare(expectedRes, *res)
			} else {
				if !reflect.DeepEqual(expectedRes, *res) {
					err = errors.New("The actual response of the {{$v.Name}} was not equal to the expected response")
				}
			}

			switch successRule {
			case successRuleAll:
				if err != nil {
					t.Fatal(err.Error())
				}
			case successRuleOnce:
				if i == loop && err != nil {
					t.Fatal(err.Error())
				}
				if err == nil {
					break FOR_LABEL
				}
			}
		}
	}
}
{{ end }}
`
