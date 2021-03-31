package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var etcdCli *clientv3.Client

func mustInitEtcdCli(etcdUrls string) {
	urls := strings.Split(etcdUrls, ",")
	for i := 0; i < len(urls); i++ {
		urls[i] = strings.TrimSpace(urls[i])
	}
	cli, err := clientv3.NewFromURLs(urls)
	if err != nil {
		logger.Panic("init etcd client failed", zap.Error(err))
	}
	etcdCli = cli
	logger.Info("etcd client initiated", zap.Strings("urls", urls))
}

func registerEndpointWithRetry(advertiseClient string) error {
	b := backoff.NewExponentialBackOff()
	for {
		var count int

		var ch <-chan *clientv3.LeaseKeepAliveResponse
		b.Reset()
		err := backoff.Retry(func() error {
			var err error
			ch, err = register(advertiseClient)
			if err != nil {
				count++
				logger.Warn("register endpoint failed, try again", zap.Error(err), zap.Int("error count", count))
				return err
			}
			logger.Info("register endpoint succeed", zap.Int("error count", count))
			count = 0
			return nil
		}, b)
		if err != nil {
			return fmt.Errorf("backoff retry failed: %v", err)
		}

		for range ch {
			// keep alive
		}
		logger.Warn("keep alive interrupted, try register endpoint again")
	}

}

func register(advertiseClient string) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	const target = "auth-svc"
	em, err := endpoints.NewManager(etcdCli, target)
	if err != nil {
		return nil, fmt.Errorf("init endpoint manager failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	leaseResp, err := etcdCli.Grant(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("grant lease failed: %v", zap.Error(err))
	}
	keepAliveRespCh, err := etcdCli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return nil, fmt.Errorf("keep alive failed: %v", zap.Error(err))
	}

	key := fmt.Sprintf("%s/%s", target, advertiseClient)
	ep := endpoints.Endpoint{Addr: advertiseClient}
	err = em.AddEndpoint(ctx, key, ep, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return nil, fmt.Errorf("register endpoint failed: %v", err)
	}

	logger.Info("register endpoint succeed", zap.Any("endpoint", ep))
	return keepAliveRespCh, nil
}

func grpcDial(svc string) (*grpc.ClientConn, error) {
	res, err := resolver.NewBuilder(etcdCli)
	if err != nil {
		return nil, fmt.Errorf("new resolver failed: %v", err)
	}
	cc, err := grpc.Dial(fmt.Sprintf("etcd:///%v", svc), grpc.WithInsecure(), grpc.WithResolvers(res))
	if err != nil {
		return nil, fmt.Errorf("grpc dial failed: %v", err)
	}
	return cc, nil
}
