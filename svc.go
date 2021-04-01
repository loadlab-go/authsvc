package main

import (
	"context"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/loadlab-go/pkg/proto/auth"
	authpb "github.com/loadlab-go/pkg/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type jwtSvc struct {
	authpb.UnimplementedJWTServer
	Key []byte
}

func (s *jwtSvc) GenerateJWT(_ context.Context, req *auth.GenerateJWTRequest) (*auth.GenerateJWTResponse, error) {
	now := time.Now()
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
		Id:        uuid.New().String(),
		Issuer:    "authmgr",
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix() - 30,
		ExpiresAt: now.AddDate(0, 0, 1).Unix(),
		Subject:   strconv.FormatInt(req.Id, 10),
	})
	token, err := jwtToken.SignedString(s.Key)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "sign failed: %v", err)
	}
	return &authpb.GenerateJWTResponse{Token: token}, nil
}

func (s *jwtSvc) ValidateJWT(_ context.Context, _ *auth.ValidateJWTRequest) (*auth.ValidateJWTResponse, error) {
	panic("not implemented") // TODO: Implement
}
