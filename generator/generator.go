package generator

import (
	"bytes"
	"errors"
	"text/template"
)

// GRPCCodeGenInfo defines the information to be rendered in the template of the GRPC test code.
type GRPCCodeGenInfo struct {
	Package         string
	GRPCServiceName string
	GRPCMethods     []GRPCMethod
}

// GRPCMethod defines the method name and the type string of the request and the type string of the response
type GRPCMethod struct {
	Name         string
	RequestType  string
	ResponseType string
}

// Validate validates that the field does not contain zero values.
func (grpcCodeGenInfo *GRPCCodeGenInfo) Validate() error {
	if grpcCodeGenInfo.Package == "" {
		return errors.New("GRPCCodeGenInfo.Package is not allowed empty")
	}
	if grpcCodeGenInfo.GRPCServiceName == "" {
		return errors.New("GRPCCodeGenInfo.GRPCServiceName is not allowed empty")
	}
	if len(grpcCodeGenInfo.GRPCMethods) == 0 {
		return errors.New("GRPCCodeGenInfo.GRPCMethods is not allowed empty")
	}
	for _, method := range grpcCodeGenInfo.GRPCMethods {
		if method.Name == "" {
			return errors.New("GRPCCodeGenInfo.GRPCMethods is not allowed empty element")
		}
		if method.RequestType == "" {
			return errors.New("GRPCCodeGenInfo.GRPCMethods is not allowed empty element")
		}
		if method.ResponseType == "" {
			return errors.New("GRPCCodeGenInfo.GRPCMethods is not allowed empty element")
		}
	}
	return nil
}

// GenerateGRPCTestCode generates gRPC scenario test code.
func GenerateGRPCTestCode(grpcCodeGenInfo GRPCCodeGenInfo) (string, error) {
	if err := grpcCodeGenInfo.Validate(); err != nil {
		return "", err
	}
	templ, _ := template.New(grpcCodeGenInfo.GRPCServiceName).Parse(codeTemplate)
	buf := bytes.Buffer{}
	if err := templ.Execute(&buf, grpcCodeGenInfo); err != nil {
		return "", err
	}
	return buf.String(), nil
}
