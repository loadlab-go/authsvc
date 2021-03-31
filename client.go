package main

import (
	"github.com/loadlab-go-go/authsvc/idl/proto/userpb"
	"go.uber.org/zap"
)

var (
	userClient userpb.UserClient
)

func mustDiscoverServices() error {
	usercc, err := grpcDial("user-svc")
	if err != nil {
		logger.Panic("grpc dial user-svc failed", zap.Error(err))
	}
	userClient = userpb.NewUserClient(usercc)

	return nil
}
