package main

import (
	"os"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"

	"github.com/yoshd/protoc-gen-stest/generator"
	"github.com/yoshd/protoc-gen-stest/processor"
)

var generateCodeFunc = func(packageName, serviceName string, methods []*descriptor.MethodDescriptorProto) string {
	grpcMethods := make([]generator.GRPCMethod, len(methods))
	for i, m := range methods {
		reqType := m.GetInputType()[1:]
		resType := m.GetOutputType()[1:]
		grpcMethods[i] = generator.GRPCMethod{
			Name:         m.GetName(),
			RequestType:  reqType,
			ResponseType: resType,
		}
	}
	grpcCodeGenInfo := generator.GRPCCodeGenInfo{
		Package:         packageName,
		GRPCServiceName: serviceName,
		GRPCMethods:     grpcMethods,
	}
	code, err := generator.GenerateGRPCTestCode(grpcCodeGenInfo)
	if err != nil {
		panic(err)
	}
	return code
}

func main() {
	req, err := processor.ParseRequest(os.Stdin)
	if err != nil {
		panic(err)
	}
	res := processor.ProcessRequest(req, generateCodeFunc)
	processor.EmitResponse(res)
}
