package xrm

import (
	"context"
	"fmt"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/wonderivan/logger"
	"github.com/xdmybl/engine-gate/util/config"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
)

// 此文件参考 go-control-plane xDS Delta Server 实现.
// TODO 后续整理这些常量

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

type Server struct {
	xdsServer server.Server
}

//func NewServer(ctx context.Context, cache cachev3.Cache, cb *test.Callbacks) *Server {
//	srv := server.NewServer(ctx, cache, cb)
//	return &Server{srv}
//}

func (s *Server) registerServer(grpcServer *grpc.Server) {
	// register services
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, s.xdsServer)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, s.xdsServer)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, s.xdsServer)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, s.xdsServer)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, s.xdsServer)
	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, s.xdsServer)
	runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, s.xdsServer)
}

func (s *Server) Run(port uint) {
	// gRPC golang library sets a very small upper bound for the number gRPC/h2
	// streams over a single TCP connection. If a proxy multiplexes requests over
	// a single connection to the management server, then it might lead to
	// availability problems. Keepalive timeouts based on connection_keepalive parameter https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples#dynamic
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    grpcKeepaliveTime,
			Timeout: grpcKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcKeepaliveMinTime,
			PermitWithoutStream: true,
		}),
	)
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Error(err)
	}

	s.registerServer(grpcServer)

	log.Printf("management server listening on %d\n", port)
	logger.Info("")
	if err = grpcServer.Serve(lis); err != nil {
		logger.Error(err)
	}
}

func registerServer(grpcServer *grpc.Server, server server.Server) {
	// register services
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, server)
	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, server)
	runtimeservice.RegisterRuntimeDiscoveryServiceServer(grpcServer, server)
}

// RunServer starts an xDS server at the given port.
func RunServer(srv server.Server, address string) {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    grpcKeepaliveTime,
			Timeout: grpcKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcKeepaliveMinTime,
			PermitWithoutStream: true,
		}),
	)
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		logger.Error(err)
	}

	registerServer(grpcServer, srv)
	logger.Info("xDS Server listen on %s", address)
	if err = grpcServer.Serve(lis); err != nil {
		logger.Error(err)
	}
}

// GenerateSnapshot TODO
// 这里创建了 默认的 资源
// 后续应该改为空资源, 因为 envoy 连接后给的配置为空.
func GenerateSnapshot() *cachev3.Snapshot {
	defaultVersion := "1"
	snap, _ := cachev3.NewSnapshot(defaultVersion,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  {makeCluster(ClusterName)},
			resource.RouteType:    {makeRoute(RouteName, ClusterName)},
			resource.ListenerType: {makeHTTPListener(ListenerName, RouteName)},
		},
		//map[resource.Type][]types.Resource{},
	)
	return snap
}

func Init() error {
	cfg := config.GetConfig()
	nodeID := cfg.Envoy.NodeId

	// Create a cache
	cache := cachev3.NewSnapshotCache(false, cachev3.IDHash{}, Logger{})

	// Create the snapshot that we'll serve to Envoy
	snapshot := GenerateSnapshot()
	// 校验 snapshot 的版本号和资源的 SHA 值是否一致.
	if err := snapshot.Consistent(); err != nil {
		logger.Error("snapshot inconsistency: %+v\n%+v", snapshot, err)
		return err
	}
	logger.Debug("will serve snapshot %+v", snapshot)

	// 为某一个静态ID为 $nodeID 的 envoy 提供服务
	// Add the snapshot to the cache
	if err := cache.SetSnapshot(context.Background(), nodeID, snapshot); err != nil {
		logger.Error("snapshot error %q for %+v", err, snapshot)
		return err
	}

	// Run the xDS server
	ctx := context.Background()
	// callbacks 是提供用户自定义的回调功能的接口, test.Callbacks 实现了
	cb := &test.Callbacks{Debug: true}
	srv := server.NewServer(ctx, cache, cb)
	RunServer(srv, fmt.Sprintf("%s:%s", cfg.XDS.Address, cfg.XDS.Port))
	return nil
}
