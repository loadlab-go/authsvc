package main

import (
	"context"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/uuid"
	authpb "github.com/loadlab-go/pkg/proto/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type jwtSvc struct {
	authpb.UnimplementedJWTServer
	Key []byte
}

func (s *jwtSvc) GenerateJWT(_ context.Context, req *authpb.GenerateJWTRequest) (*authpb.GenerateJWTResponse, error) {
	now := time.Now()
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
		ID:        uuid.New().String(),
		Issuer:    "authmgr",
		IssuedAt:  jwt.NewTime(float64(now.Second())),
		NotBefore: jwt.NewTime(float64(now.Unix()) - 30),
		ExpiresAt: jwt.NewTime(float64(now.AddDate(0, 0, 1).Unix())),
		Subject:   strconv.FormatInt(req.Id, 10),
	})
	token, err := jwtToken.SignedString(s.Key)
	if err != nil {
		logger.Warn("sign failed", zap.Error(err))
		return nil, status.Errorf(codes.Aborted, "sign failed: %v", err)
	}
	return &authpb.GenerateJWTResponse{Token: token}, nil
}

func (s *jwtSvc) ValidateJWT(_ context.Context, req *authpb.ValidateJWTRequest) (*authpb.ValidateJWTResponse, error) {
	jwtToken, err := jwt.ParseWithClaims(req.Token, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return s.Key, nil
	})
	if err != nil {
		logger.Warn("jwt parse failed", zap.Error(err))
		return nil, status.Errorf(codes.Aborted, "jwt parse failed: %v", err)
	}
	claims := jwtToken.Claims.(*jwt.StandardClaims)
	return &authpb.ValidateJWTResponse{
		Aud: claims.Audience,
		Exp: claims.ExpiresAt.Unix(),
		Jti: claims.ID,
		Iat: claims.IssuedAt.Unix(),
		Iss: claims.Issuer,
		Nbf: claims.NotBefore.Unix(),
		Sub: claims.Subject,
	}, nil
}
