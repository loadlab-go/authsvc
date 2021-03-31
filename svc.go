package main

import (
	"context"

	"github.com/loadlab-go-go/authsvc/idl/proto/authpb"
	"go.uber.org/zap"
)

type authSvc struct {
	authpb.UnimplementedAuthServer
}

func (s *authSvc) Authenticate(_ context.Context, req *authpb.AuthenticateRequest) (*authpb.AuthenticateResponse, error) {
	logger.Info("auth request", zap.Any("req", req))
	return &authpb.AuthenticateResponse{Jwt: "ajsidjoasjdoiasjdioasdl"}, nil
}
func (s *authSvc) Validate(_ context.Context, req *authpb.ValidateRequest) (*authpb.ValidateResponse, error) {
	logger.Info("validate request", zap.Any("req", req))
	return &authpb.ValidateResponse{Ok: true}, nil
}