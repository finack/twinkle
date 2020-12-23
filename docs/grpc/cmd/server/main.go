package main

import (
	"log"
	"net"

	"github.com/finack/twinkle/internal/grpcserver"
	pb "github.com/finack/twinkle/rpc/metarmap"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen on port %v: %v", port, err)
	}

	s := grpc.NewServer()

	pb.RegisterMetarMapServer(s, &grpserver.Server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
