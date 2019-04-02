package support

import (
	"context"

	pbUser "github.com/vvatanabe/go-grpc-microservices/proto/user"
)

type contextKeyTraceID struct{}
type contextKeyUser struct{}

func GetTraceIDFromContext(ctx context.Context) string {
	id := ctx.Value(contextKeyTraceID{})
	traceID, ok := id.(string)
	if !ok {
		return ""
	}
	return traceID
}

func AddTraceIDToContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, contextKeyTraceID{}, traceID)
}

func GetUserFromContext(ctx context.Context) *pbUser.User {
	u := ctx.Value(contextKeyUser{})
	pUser, ok := u.(*pbUser.User)
	if !ok {
		return nil
	}
	return pUser
}

func AddUserToContext(ctx context.Context,
	user *pbUser.User) context.Context {
	return context.WithValue(ctx, contextKeyUser{}, user)
}
