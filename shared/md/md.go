package md

import (
	"context"
	"errors"
	"strconv"

	"google.golang.org/grpc/metadata"
)

const metadataKeyUserID string = "x-user-id"

func AddUserIDToContext(ctx context.Context, userID uint64) context.Context {
	return metadata.AppendToOutgoingContext(ctx, metadataKeyUserID, strconv.FormatUint(userID, 10))
}

var ErrNotFoundUserID = errors.New("not found user id")

func GetUserIDFromContext(ctx context.Context) uint64 {
	userID, err := SafeGetUserIDFromContext(ctx)
	if err != nil {
		panic(err)
	}
	return userID
}

func SafeGetUserIDFromContext(ctx context.Context) (userID uint64, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return userID, ErrNotFoundUserID
	}
	values := md.Get(metadataKeyUserID)
	if len(values) < 1 {
		return userID, ErrNotFoundUserID
	}
	userID, err = strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return userID, err
	}
	return userID, nil
}

const metadataKeyTraceID string = "x-trace-id"

func GetTraceIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(metadataKeyTraceID)
	if len(values) < 1 {
		return ""
	}
	return values[0]
}

func AddTraceIDToContext(ctx context.Context, traceID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, metadataKeyTraceID, traceID)
}
