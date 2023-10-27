package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/rboyer/blankspace/blankpb"
)

func serveGRPC(name, addr string) error {
	if !isValidAddr(addr) {
		return fmt.Errorf("-grpc-addr is invalid %q", addr)
	}
	log.Printf("gRPC listening on %s", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	srv := grpc.NewServer()
	reflection.Register(srv)
	blankpb.RegisterServerServer(srv, &grpcServer{
		name: name,
	})

	return srv.Serve(listener)
}

type grpcServer struct {
	name string
}

var _ blankpb.ServerServer = (*grpcServer)(nil)

func (s *grpcServer) Describe(ctx context.Context, req *blankpb.DescribeRequest) (*blankpb.DescribeResponse, error) {
	resp := &blankpb.DescribeResponse{
		Name: s.name,
	}
	return resp, nil
}
