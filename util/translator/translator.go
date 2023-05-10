package translator

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/wonderivan/logger"
	"github.com/xdmybl/engine-gate/util/common"

	v1 "github.com/xdmybl/gate-type/pkg/api/gate.xdmybl.io/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// 出错不要在这里解决, 在前面的 webserver 模块去解决, 这里一定要保证不会 error, 即使 error 也应该控制到最小范围

// TranslateUpstreamLsToCluster
// 将 upstream 对象转换成 envoy cluster 对象
// 1. 即使 translate 中出现问题, 尽量不让程序因此退出, 而是帮他生成一个相对正确的资源
func TranslateUpstreamLsToCluster(slice v1.UpstreamSlice) []*clusterv3.Cluster {
	logger.Trace("TranslateUpstreamLsToCluster slice: %v", slice)
	var clusterLs []*clusterv3.Cluster
	for _, obj := range slice {
		var cluster *clusterv3.Cluster
		connPoll := obj.Spec.GetConnPoll()
		if connPoll == nil {
			logger.Error("%s: %v", obj.GetName(), "connection pool nil")
		}

		thresholds := &clusterv3.CircuitBreakers_Thresholds{
			MaxConnections:     &wrapperspb.UInt32Value{Value: common.Int64ToUint32(connPoll.GetMaxConnections())},
			MaxRequests:        &wrapperspb.UInt32Value{Value: common.Int64ToUint32(connPoll.GetMaxRequests())},
			MaxPendingRequests: &wrapperspb.UInt32Value{Value: common.Int64ToUint32(connPoll.GetMaxPendingRequests())},
		}

		circuitBreaker := &clusterv3.CircuitBreakers{
			Thresholds: []*clusterv3.CircuitBreakers_Thresholds{
				thresholds,
			},
		}

		clusterLoadAssignment := &endpointv3.ClusterLoadAssignment{
			Endpoints: []*endpointv3.LocalityLbEndpoints{
				{
					LbEndpoints: []*endpointv3.LbEndpoint{},
				},
			},
			ClusterName: obj.GetName(),
		}

		for _, ep := range obj.Spec.Endpoints {
			lbEndpoint := &endpointv3.LbEndpoint{
				HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
					Endpoint: &endpointv3.Endpoint{
						Address: &corev3.Address{
							Address: &corev3.Address_SocketAddress{
								SocketAddress: &corev3.SocketAddress{
									Address: ep.GetAddress(),
									PortSpecifier: &corev3.SocketAddress_PortValue{
										PortValue: ep.GetPort(),
									},
								},
							},
						},
					},
				},
			}
			if ep.GetLoadBalancingWeight() != 0 {
				lbEndpoint.LoadBalancingWeight = &wrappers.UInt32Value{Value: ep.GetLoadBalancingWeight()}
			}

			clusterLoadAssignment.Endpoints[0].LbEndpoints = append(clusterLoadAssignment.Endpoints[0].LbEndpoints, lbEndpoint)
		}

		cluster = &clusterv3.Cluster{
			Name:            obj.GetName(),
			CircuitBreakers: circuitBreaker,
			LoadAssignment:  clusterLoadAssignment,
			// TransportSocket: tls,
			LbPolicy: clusterv3.Cluster_LbPolicy(obj.Spec.GetLbAlg()),
			ClusterDiscoveryType: &clusterv3.Cluster_Type{
				Type: clusterv3.Cluster_STRICT_DNS,
			},
			DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
		}
		clusterLs = append(clusterLs, cluster)
	}
	return clusterLs
}
