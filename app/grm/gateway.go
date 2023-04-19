package grm

import (
	"context"
	"github.com/wonderivan/logger"
	"github.com/xdmybl/engine-gate/util/config"
	v1 "github.com/xdmybl/gate-type/pkg/api/gate.xdmybl.io/v1"
	v12 "github.com/xdmybl/gate-type/pkg/api/gate.xdmybl.io/v1/providers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GatewayManager struct {
	GatewayClient v1.GatewayClient
}

func (c *GatewayManager) Init() error {
	cfg := config.GetConfig()
	kubeConnectConfig := &rest.Config{}
	var err error
	if cfg.KubeConfig != "" {
		kubeConnectConfig, err = clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
	} else if cfg.Token != "" && cfg.MasterUrl != "" {
		kubeConnectConfig.BearerToken = cfg.Token
		kubeConnectConfig.Host = cfg.MasterUrl
	}
	if err != nil {
		logger.Error("kube config err:  %v", err)
		return err
	}
	factory := v12.GatewayClientFromConfigFactoryProvider()
	caClient, err := factory(kubeConnectConfig)
	c.GatewayClient = caClient
	return err
}

func (c *GatewayManager) Create(ctx context.Context, obj *v1.Gateway) error {
	err := c.GatewayClient.CreateGateway(ctx, obj)
	return err
}

func (c *GatewayManager) Get(ctx context.Context, name string) (*v1.Gateway, error) {
	cfg := config.GetConfig()

	o := client.ObjectKey{
		Namespace: cfg.Namespace,
		Name:      name,
	}
	return c.GatewayClient.GetGateway(ctx, o)
}

// Filter todo
func (c *GatewayManager) Filter() []*v1.CaCertificate {

	return []*v1.CaCertificate{}
}

func (c *GatewayManager) Delete(ctx context.Context, name string) error {
	cfg := config.GetConfig()
	o := client.ObjectKey{
		Namespace: cfg.Namespace,
		Name:      name,
	}
	return c.GatewayClient.DeleteGateway(ctx, o)
}

func UpsertGatewayFunc(existing, desired *v1.Gateway) error {
	return nil
}

// Update obj 表示更新后的对象, 之前的对象可以通过 key 自己找出来的
func (c *GatewayManager) Update(ctx context.Context, obj *v1.Gateway) error {
	err := c.GatewayClient.UpsertGateway(ctx, obj, UpsertGatewayFunc)
	if err != nil {
		logger.Error("gateway update err:  %v", err)
	}
	return err
}
