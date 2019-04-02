package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	pbActivity "github.com/vvatanabe/go-grpc-microservices/proto/activity"
	pbProject "github.com/vvatanabe/go-grpc-microservices/proto/project"
	"github.com/vvatanabe/go-grpc-microservices/shared/interceptor"
	"google.golang.org/grpc"
)

const port = ":50051"

func main() {
	activityConn, err := grpc.Dial(os.Getenv("ACTIVITY_SERVICE_ADDR"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial activity service: %s", err)
	}
	activityClient := pbActivity.NewActivityServiceClient(activityConn)
	srv := grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		interceptor.XTraceID(),
		interceptor.Logging(),
		interceptor.XUserID(),
	)))
	pbProject.RegisterProjectServiceServer(srv, &ProjectService{
		store:          NewStoreOnMemory(),
		activityClient: activityClient,
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
