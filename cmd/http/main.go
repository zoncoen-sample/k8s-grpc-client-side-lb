package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	pb "github.com/zoncoen-sample/k8s-grpc-client-side-lb/pb/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	channelzsvc "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"
	_ "google.golang.org/grpc/resolver/dns"
)

func init() {
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, os.Stderr, os.Stderr))
}

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	clusterIPClient := newInformantionClient(os.Getenv("CLUSTER_IP_SERVICE"))
	headlessClient := newInformantionClient(os.Getenv("HEADLESS_SERVICE"))
	clusterIPClientWithLB := newInformantionClient(fmt.Sprintf("dns:///%s", os.Getenv("CLUSTER_IP_SERVICE")))
	headlessClientWithLB := newInformantionClient(fmt.Sprintf("dns:///%s", os.Getenv("HEADLESS_SERVICE")))
	mux := http.NewServeMux()
	mux.Handle("/cluster-ip", getHostnameHandler(clusterIPClient))
	mux.Handle("/headless", getHostnameHandler(headlessClient))
	mux.Handle("/cluster-ip/lb", getHostnameHandler(clusterIPClientWithLB))
	mux.Handle("/headless/lb", getHostnameHandler(headlessClientWithLB))
	httpServer := http.Server{
		Addr:    httpPort,
		Handler: mux,
	}

	// create server for checking gRPC channels
	grpcPort, err := net.Listen("tcp", os.Getenv("GRPC_PORT"))
	if err != nil {
		log.Fatalln(err)
	}
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	channelzsvc.RegisterChannelzServiceToServer(grpcServer)

	shutdownCh := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		grpcServer.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Println(err)
		}
		close(shutdownCh)
	}()

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println(err)
		}
	}()
	go func() {
		if err := grpcServer.Serve(grpcPort); err != nil {
			log.Println(err)
		}
	}()

	<-shutdownCh
}

func newInformantionClient(target string) pb.InformationClient {
	conn, err := grpc.Dial(
		target,
		grpc.WithInsecure(),
		grpc.WithBalancerName(roundrobin.Name),
	)
	if err != nil {
		log.Fatalln(err)
	}
	return pb.NewInformationClient(conn)
}

func getHostnameHandler(client pb.InformationClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.GetHostname(r.Context(), &pb.GetHostnameRequest{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte(resp.GetHostname()))
	})
}
