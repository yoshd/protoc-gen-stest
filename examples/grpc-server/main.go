package main

import (
	"flag"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yoshd/protoc-gen-stest/examples/pb"
)

var (
	addr = flag.String("addr", "localhost:13009", "addr host:port")
)

type server struct{}

func (s *server) Hello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{ResMsg: "Hello!"}, nil
}

func (s *server) Bye(ctx context.Context, in *pb.ByeRequest) (*pb.ByeResponse, error) {
	if in.ReqMsg == "error" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument")
	}
	return &pb.ByeResponse{ResMsg: "Bye!"}, nil
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)

	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()

	pb.RegisterSampleServer(s, &server{})
	s.Serve(lis)
}
