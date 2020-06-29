package processor

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"unicode"

	"github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

// ParseRequest parses the input and returns the CodeGeneratorRequest of protoc.
func ParseRequest(r io.Reader) (*plugin.CodeGeneratorRequest, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var req plugin.CodeGeneratorRequest
	if err = proto.Unmarshal(buf, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// ProcessRequest processes the request and returns a response to generate the code.
func ProcessRequest(req *plugin.CodeGeneratorRequest, genCodeFunc func(packageName, serviceName string, methods []*descriptor.MethodDescriptorProto) string) *plugin.CodeGeneratorResponse {
	files := make(map[string]*descriptor.FileDescriptorProto)
	for _, f := range req.ProtoFile {
		files[f.GetName()] = f
	}
	var res plugin.CodeGeneratorResponse
	for _, fname := range req.FileToGenerate {
		f := files[fname]
		for _, service := range f.GetService() {
			// packageName := f.GetOptions().GetGoPackage()
			packageName := f.GetPackage()
			serviceName := service.GetName()
			methods := service.GetMethod()
			genCode := genCodeFunc(packageName, serviceName, methods)
			serviceNameSnakeCase := toSnakeCase(service.GetName())
			outputFname := serviceNameSnakeCase + "_scenariotest.go"
			res.File = append(res.File, &plugin.CodeGeneratorResponse_File{
				Name:    proto.String(outputFname),
				Content: proto.String(genCode),
			})
		}
	}
	return &res
}

// EmitResponse returns the response of protoc
func EmitResponse(res *plugin.CodeGeneratorResponse) error {
	buf, err := proto.Marshal(res)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(buf)
	return err
}

func toSnakeCase(str string) (snakeCaseStr string) {
	for i, c := range str {
		if unicode.IsUpper(c) {
			lower := strings.ToLower(string(c))
			if i == 0 {
				snakeCaseStr = lower
				continue
			}
			snakeCaseStr = snakeCaseStr + "_" + lower
			continue
		}
		snakeCaseStr = snakeCaseStr + string(c)
	}
	return snakeCaseStr
}
