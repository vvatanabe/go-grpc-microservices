package main

import (
	"context"
	"net"

	"os"

	"log"

	"time"

	"os/signal"
	"syscall"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	pbActivity "github.com/vvatanabe/go-grpc-microservices/proto/activity"
	pbProject "github.com/vvatanabe/go-grpc-microservices/proto/project"
	pbTask "github.com/vvatanabe/go-grpc-microservices/proto/task"
	"github.com/vvatanabe/go-grpc-microservices/shared/interceptor"
	"google.golang.org/grpc"
)

const port = ":50051"

func main() {
	// クライアントスタブの生成
	activityConn, err := grpc.Dial(
		os.Getenv("ACTIVITY_SERVICE_ADDR"),
		grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial activity: %s",
			err)
	}
	projectConn, err := grpc.Dial(
		os.Getenv("PROJECT_SERVICE_ADDR"),
		grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial task: %s",
			err)
	}
	// インタセプタの追加
	chain := grpc_middleware.ChainUnaryServer(
		interceptor.XTraceID(),
		interceptor.Logging(),
		interceptor.XUserID(),
	)
	srvOpt := grpc.UnaryInterceptor(chain)
	srv := grpc.NewServer(srvOpt)
	// サービスの登録
	pbTask.RegisterTaskServiceServer(srv, &TaskService{
		store: NewStoreOnMemory(),
		activityClient: pbActivity.
			NewActivityServiceClient(activityConn),
		projectClient: pbProject.
			NewProjectServiceClient(projectConn),
	})
	// gRPC接続の待ち受け
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
	// グレースフルストップ
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
