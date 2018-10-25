package main

import (
	"context"
	"log"
	"net"
	"os"

	pb "github.com/zoncoen-sample/k8s-grpc-client-side-lb/pb/proto"
	"google.golang.org/grpc"
	channelzsvc "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"
)

func init() {
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, os.Stderr, os.Stderr))
}

func main() {
	port, err := net.Listen("tcp", os.Getenv("GRPC_PORT"))
	if err != nil {
		log.Fatalln(err)
	}
	server := grpc.NewServer()
	svc := &service{}
	pb.RegisterInformationServer(server, svc)
	reflection.Register(server)
	channelzsvc.RegisterChannelzServiceToServer(server)
	if err := server.Serve(port); err != nil {
		log.Fatalln(err)
	}
}

type service struct{}

func (s *service) GetHostname(context.Context, *pb.GetHostnameRequest) (*pb.GetHostnameResponse, error) {
	name, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return &pb.GetHostnameResponse{
		Hostname: name,
	}, nil
}
