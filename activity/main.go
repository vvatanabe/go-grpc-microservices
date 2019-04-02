package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	pbActivity "github.com/vvatanabe/go-grpc-microservices/proto/activity"
	"github.com/vvatanabe/go-grpc-microservices/shared/interceptor"
	"google.golang.org/grpc"
)

const port = ":50051"

func main() {
	srv := grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		interceptor.XTraceID(),
		interceptor.Logging(),
		interceptor.XUserID(),
	)))
	pbActivity.RegisterActivityServiceServer(srv, &ActivityService{
		store: NewStoreOnMemory(),
	})
	go func() {
		listener, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to create listener: %s",
				err)
		}
		log.Println("start server on port", port)
		if err := srv.Serve(listener); err != nil {
			log.Println("failed to exit serve: ", err)
		}
	}()
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM)
	<-sigint
	log.Println("received a signal of graceful shutdown")
	stopped := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(stopped)
	}()
	ctx, cancel := context.WithTimeout(
		context.Background(), 1*time.Minute)
	select {
	case <-ctx.Done():
		srv.Stop()
	case <-stopped:
		cancel()
	}
	log.Println("completed graceful shutdown")
}
