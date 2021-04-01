package main

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	authpb "github.com/loadlab-go/pkg/proto/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	flagListenAddr      = flag.String("listen", ":8080", "listen address")
	flagEtcdEndpoints   = flag.String("etcd-endpoints", "localhost:2379", "etcd endpoints")
	flagAdvertiseClient = flag.String("advertise-client", "localhost:8080", "advertise client url")
	flagJWTKey          = flag.String("jwt-key", "this is secret", "jwt key")
)

func main() {
	flag.Parse()

	mustInitEtcdCli(*flagEtcdEndpoints)

	go func() {
		err := registerEndpointWithRetry(*flagAdvertiseClient)
		if err != nil {
			logger.Panic("register endpoint faield", zap.Error(err))
		}
	}()

	srv := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
		)))

	js := &jwtSvc{Key: []byte(*flagJWTKey)}
	authpb.RegisterJWTServer(srv, js)

	l, err := net.Listen("tcp", *flagListenAddr)
	if err != nil {
		logger.Fatal("listen failed", zap.Error(err))
	}
	logger.Info("auth service server startup", zap.String("listen", l.Addr().String()))

	go signalSet(srv.GracefulStop)

	err = srv.Serve(l)
	if err != nil {
		logger.Panic("auth service server serve failed", zap.Error(err))
	}
	logger.Info("server stopped")
}

func signalSet(cb func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	s := <-sigCh
	logger.Warn("exit signal", zap.String("signal", s.String()))

	cb()
}
