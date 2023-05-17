package translator

import (
	"fmt"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	rawbufferv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/raw_buffer/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/wonderivan/logger"
	_type "github.com/xdmybl/engine-gate/type"
	"github.com/xdmybl/engine-gate/util/common"
	v11 "github.com/xdmybl/gate-type/pkg/api/common/v1"
	v1 "github.com/xdmybl/gate-type/pkg/api/gate.xdmybl.io/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// 出错不要在这里解决, 在前面的 webserver 模块去解决, 这里一定要保证不会 error, 即使 error 也应该控制到最小范围

// TranslateUpstreamLsToCluster
// 将 upstream 对象转换成 envoy cluster 对象
// 1. 即使 translate 中出现问题, 尽量不让程序因此退出, 而是帮他生成一个相对正确的资源
func TranslateUpstreamLsToCluster(snapshot _type.Snapshot, slice v1.UpstreamSlice) []*clusterv3.Cluster {
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

		//cluster = &clusterv3.Cluster{
		//	Name:            obj.GetName(),
		//	CircuitBreakers: circuitBreaker,
		//	LoadAssignment:  clusterLoadAssignment,
		//	// TransportSocket: tls,
		//	LbPolicy: clusterv3.Cluster_LbPolicy(obj.Spec.GetLbAlg()),
		//	ClusterDiscoveryType: &clusterv3.Cluster_Type{
		//		Type: clusterv3.Cluster_STRICT_DNS,
		//	},
		//	DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
		//}

		hcs := make([]*corev3.HealthCheck, 0)
		if spe := obj.Spec.GetHcSpecifier(); spe.GetType() == v1.HealthCheckSpecifier_HTTP {
			httpHC := spe.GetHttpHealthCheck()
			eStats := make([]*typev3.Int64Range, 0)
			for _, ui32 := range httpHC.GetExpectedStatuses() {
				start := int64(ui32)
				eStats = append(eStats, &typev3.Int64Range{
					Start: start,
					End:   start + 1,
				})
			}
			// hc 一般只需要一个, 不过 envoy 是支持多个的, 目前系统一个 upstream 只对应一个 hc.
			hc := &corev3.HealthCheck{
				Interval:           &duration.Duration{Seconds: int64(obj.Spec.GetHcInterval())},
				Timeout:            &duration.Duration{Seconds: int64(obj.Spec.GetHcTimeout())},
				HealthyThreshold:   &wrapperspb.UInt32Value{Value: obj.Spec.GetHcHealthyThreshold()},
				UnhealthyThreshold: &wrapperspb.UInt32Value{Value: obj.Spec.GetHcUnhealthyThreshold()},
				HealthChecker: &corev3.HealthCheck_HttpHealthCheck_{
					HttpHealthCheck: &corev3.HealthCheck_HttpHealthCheck{
						Host:             httpHC.GetHost(),
						Path:             httpHC.GetPath(),
						ExpectedStatuses: eStats,
						CodecClientType:  typev3.CodecClientType(typev3.CodecClientType_value[httpHC.GetClientType()]),
					},
				},
			}
			hcs = append(hcs, hc)
		}

		tls := computeClusterTransportSocket(snapshot, obj)

		cluster = &clusterv3.Cluster{
			Name:            obj.GetName(),
			CircuitBreakers: circuitBreaker,
			LoadAssignment:  clusterLoadAssignment,
			TransportSocket: tls,
			HealthChecks:    hcs,
			LbPolicy:        clusterv3.Cluster_LbPolicy(obj.Spec.GetLbAlg()),
			ClusterDiscoveryType: &clusterv3.Cluster_Type{
				// Type: envoy_config_cluster_v3.Cluster_LOGICAL_DNS,
				Type: clusterv3.Cluster_STRICT_DNS,
			},
			DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
			// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#config-cluster-v3-trackclusterstats
			// RequestResponseSizes: true 会在统计中额外发布一些信息
			TrackClusterStats: &clusterv3.TrackClusterStats{RequestResponseSizes: true},
			// 可理解为健康检查的严格模式
			CommonLbConfig: &clusterv3.Cluster_CommonLbConfig{
				HealthyPanicThreshold: &typev3.Percent{
					Value: 0,
				},
			},
		}
		clusterLs = append(clusterLs, cluster)
	}
	return clusterLs
}

func computeClusterTransportSocket(snapshot _type.Snapshot, upstream *v1.Upstream) *corev3.TransportSocket {
	var (
		transportSocket *corev3.TransportSocket
		tlsClient       *v11.TlsClient
	)
	if tlsClient = upstream.Spec.GetSslConfigurations(); tlsClient == nil {
		logger.Error("%s: %v", upstream.GetName(), "tlsClient configuration nil")
	}

	switch tlsClient.GetTlsMode() {
	case v11.TlsMode_TLS_NONE:
		if _any, err := common.MessageToAny(&rawbufferv3.RawBuffer{}); err == nil {
			transportSocket = &corev3.TransportSocket{
				Name: wellknown.TransportSocketTls,
				ConfigType: &corev3.TransportSocket_TypedConfig{
					TypedConfig: _any,
				},
			}
		}
	case v11.TlsMode_TLS_V1_SIMPLE:
		// 简单认证模式下，upstream验证对端数字证书，故只需要一份对端数字证书的根证书
		dataSource := common.DataSourceGenerator(true)

		caCert := snapshot.GetCaByName(tlsClient.GetCaCertRef().GetName())

		if caCert == nil {
			// 为空时，默认使用envoy提供的根证书
			tlsCtx := &tlsv3.UpstreamTlsContext{
				CommonTlsContext: &tlsv3.CommonTlsContext{
					TlsParams: &tlsv3.TlsParameters{
						CipherSuites: []string{
							// Envoy 默认列表
							"ECDHE-ECDSA-AES128-GCM-SHA256",
							"ECDHE-ECDSA-CHACHA20-POLY1305",
							"ECDHE-RSA-AES128-GCM-SHA256",
							"ECDHE-RSA-CHACHA20-POLY1305",
							"ECDHE-ECDSA-AES256-GCM-SHA384",
							"ECDHE-RSA-AES256-GCM-SHA384",
							// 适配深信服设备
							"TLS_RSA_WITH_AES_256_CBC_SHA",
						},
					},
				},
				AllowRenegotiation: tlsClient.GetAllowRenegotiation(),
				Sni:                tlsClient.GetSni(),
			}
			typedConfig, err := common.MessageToAny(tlsCtx)
			if typedConfig != nil {
				transportSocket = &corev3.TransportSocket{
					Name: wellknown.TransportSocketTls,
					ConfigType: &corev3.TransportSocket_TypedConfig{
						TypedConfig: typedConfig,
					},
				}
			} else {
				logger.Error("%s: %v", upstream.GetName(), err)
			}
		} else {
			tlsCtx := &tlsv3.UpstreamTlsContext{
				CommonTlsContext: &tlsv3.CommonTlsContext{
					TlsParams: &tlsv3.TlsParameters{
						TlsMinimumProtocolVersion: tlsv3.TlsParameters_TlsProtocol(caCert.Spec.GetTlsParameters().GetMinimumProtocolVersion()),
						TlsMaximumProtocolVersion: tlsv3.TlsParameters_TlsProtocol(caCert.Spec.GetTlsParameters().GetMaximumProtocolVersion()),
						CipherSuites:              caCert.Spec.GetTlsParameters().GetCipherSuites(),
						EcdhCurves:                caCert.Spec.GetTlsParameters().GetEcdhCurves(),
					},
					AlpnProtocols: caCert.Spec.GetAlpnProtocols(),
				},
				AllowRenegotiation: tlsClient.GetAllowRenegotiation(),
				Sni:                tlsClient.GetSni(),
			}

			tlsCtx.CommonTlsContext.ValidationContextType = &tlsv3.CommonTlsContext_ValidationContext{
				ValidationContext: &tlsv3.CertificateValidationContext{
					TrustedCa:               dataSource(caCert.Spec.GetCa()),
					Crl:                     dataSource(caCert.Spec.GetCrl()),
					AllowExpiredCertificate: caCert.Spec.GetAllowExpiredCertificate(),
				},
			}

			typedConfig, err := common.MessageToAny(tlsCtx)
			if typedConfig != nil {
				transportSocket = &corev3.TransportSocket{
					Name: wellknown.TransportSocketTls,
					ConfigType: &corev3.TransportSocket_TypedConfig{
						TypedConfig: typedConfig,
					},
				}
			} else {
				logger.Error("%s: %v", upstream.GetName(), err)
			}
		}
	case v11.TlsMode_TLS_V1_MUTUAL:
		dataSource := common.DataSourceGenerator(true)

		caCert := snapshot.GetCaByName(tlsClient.GetCaCertRef().GetName())
		sslCert := snapshot.GetCertByName(tlsClient.GetSslCertRef().GetName())

		if caCert == nil || sslCert == nil {
			logger.Error(fmt.Errorf("ca or cert no found"))
			return nil
		}

		// Mutual模式需要验证数字证书

		tlsCtx := &tlsv3.UpstreamTlsContext{
			CommonTlsContext: &tlsv3.CommonTlsContext{
				TlsParams: &tlsv3.TlsParameters{
					TlsMinimumProtocolVersion: tlsv3.TlsParameters_TlsProtocol(sslCert.Spec.GetTlsParameters().GetMinimumProtocolVersion()),
					TlsMaximumProtocolVersion: tlsv3.TlsParameters_TlsProtocol(sslCert.Spec.GetTlsParameters().GetMaximumProtocolVersion()),
					CipherSuites:              sslCert.Spec.GetTlsParameters().GetCipherSuites(),
					EcdhCurves:                sslCert.Spec.GetTlsParameters().GetEcdhCurves(),
				},
				AlpnProtocols: sslCert.Spec.GetAlpnProtocols(),
			},
			AllowRenegotiation: tlsClient.GetAllowRenegotiation(),
			Sni:                tlsClient.GetSni(),
		}

		tlsCtx.CommonTlsContext.TlsCertificates = []*tlsv3.TlsCertificate{
			{
				CertificateChain: dataSource(sslCert.Spec.GetCertificateChain()),
				PrivateKey:       dataSource(sslCert.Spec.GetPrivateKey()),
			},
		}

		tlsCtx.CommonTlsContext.ValidationContextType = &tlsv3.CommonTlsContext_ValidationContext{
			ValidationContext: &tlsv3.CertificateValidationContext{
				TrustedCa:               dataSource(caCert.Spec.GetCa()),
				Crl:                     dataSource(caCert.Spec.GetCrl()),
				AllowExpiredCertificate: caCert.Spec.GetAllowExpiredCertificate(),
			},
		}

		typedConfig, err := common.MessageToAny(tlsCtx)
		if typedConfig != nil {
			transportSocket = &corev3.TransportSocket{
				Name: wellknown.TransportSocketTls,
				ConfigType: &corev3.TransportSocket_TypedConfig{
					TypedConfig: typedConfig,
				},
			}
		} else {
			logger.Error("%s: %v", upstream.GetName(), err)
		}
	}

	return transportSocket
}
