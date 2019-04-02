package main

import (
	"log"
	"os"

	"net/http"

	"github.com/gorilla/mux"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/vvatanabe/go-grpc-microservices/front/handler"
	"github.com/vvatanabe/go-grpc-microservices/front/interceptor"
	"github.com/vvatanabe/go-grpc-microservices/front/middleware"
	"github.com/vvatanabe/go-grpc-microservices/front/session"
	pbActivity "github.com/vvatanabe/go-grpc-microservices/proto/activity"
	pbProject "github.com/vvatanabe/go-grpc-microservices/proto/project"
	pbTask "github.com/vvatanabe/go-grpc-microservices/proto/task"
	pbUser "github.com/vvatanabe/go-grpc-microservices/proto/user"
	"google.golang.org/grpc"
)

const port = ":8080"

func main() {
	// 各マイクロサービスのクライアントスタブの生成
	activityClient := pbActivity.
		NewActivityServiceClient(getGRPCConn(
			os.Getenv("ACTIVITY_SERVICE_ADDR"),
			interceptor.XTraceID,
			interceptor.XUserID))
	projectClient := pbProject.
		NewProjectServiceClient(getGRPCConn(
			os.Getenv("PROJECT_SERVICE_ADDR"),
			interceptor.XTraceID,
			interceptor.XUserID))
	taskClient := pbTask.
		NewTaskServiceClient(getGRPCConn(
			os.Getenv("TASK_SERVICE_ADDR"),
			interceptor.XTraceID,
			interceptor.XUserID))
	userClient := pbUser.
		NewUserServiceClient(getGRPCConn(
			os.Getenv("USER_SERVICE_ADDR"),
			interceptor.XTraceID))
	sessionStore := session.NewStoreOnMemory()
	frontSrv := &handler.FrontServer{
		ActivityClient: activityClient,
		ProjectClient:  projectClient,
		TaskClient:     taskClient,
		UserClient:     userClient,
		SessionStore:   sessionStore,
	}
	r := mux.NewRouter()
	// ミドルウェアを使ってハンドラ共通の処理の追加
	r.Use(middleware.Tracing)
	r.Use(middleware.Logging)
	auth := middleware.
		NewAuthentication(userClient, sessionStore)
	// エンドポイントとメソッドのマッピング
	// （認証が必要なエンドポイントには
	// 認証チェック用のミドルウェアを追加）
	r.Path("/").Methods(http.MethodGet).
		HandlerFunc(auth(frontSrv.ViewHome))
	r.Path("/logout").Methods(http.MethodPost).
		HandlerFunc(auth(frontSrv.Logout))
	r.Path("/project").Methods(http.MethodPost).
		HandlerFunc(auth(frontSrv.CreateProject))
	r.Path("/project/{id}").Methods(http.MethodGet).
		HandlerFunc(auth(frontSrv.ViewProject))
	r.Path("/project/{id}").Methods(http.MethodPost).
		HandlerFunc(auth(frontSrv.UpdateProject))
	r.Path("/task").Methods(http.MethodPost).
		HandlerFunc(auth(frontSrv.CreateTask))
	r.Path("/task/{id}").Methods(http.MethodPost).
		HandlerFunc(auth(frontSrv.UpdateTask))
	r.Path("/signup").Methods(http.MethodGet).
		HandlerFunc(frontSrv.ViewSignup)
	r.Path("/signup").Methods(http.MethodPost).
		HandlerFunc(frontSrv.Signup)
	r.Path("/login").Methods(http.MethodGet).
		HandlerFunc(frontSrv.ViewLogin)
	r.Path("/login").Methods(http.MethodPost).
		HandlerFunc(frontSrv.Login)
	static := http.StripPrefix("/static",
		http.FileServer(http.Dir("static")))
	r.PathPrefix("/static/").Handler(static)
	log.Println("start server on port", port)
	err := http.ListenAndServe(port, r)
	if err != nil {
		log.Println("failed to exit serve: ", err)
	}
}

func getGRPCConn(target string,
	interceptors ...grpc.UnaryClientInterceptor,
) *grpc.ClientConn {
	// インタセプタを使ってRPC共通の処理の追加
	chain := grpc_middleware.
		ChainUnaryClient(interceptors...)
	conn, err := grpc.Dial(target,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(chain))
	if err != nil {
		log.Fatalf("failed to dial: %s", err)
	}
	return conn
}
