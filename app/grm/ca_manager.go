package grm

import (
	"context"
	"github.com/wonderivan/logger"
	"github.com/xdmybl/engine-gate/util/config"
	"github.com/xdmybl/gate-type/pkg/api/gate.xdmybl.io/v1"
	v12 "github.com/xdmybl/gate-type/pkg/api/gate.xdmybl.io/v1/providers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CaManager struct {
	CaClient v1.CaCertificateClient
}

func (c *CaManager) Init() error {
	cfg := config.GetConfig()
	// TODO 加载 kubeconfig 或 静态 token, master url, 等配置
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
	factory := v12.CaCertificateClientFromConfigFactoryProvider()
	caClient, err := factory(kubeConnectConfig)
	c.CaClient = caClient
	return err
}

func (c *CaManager) Create(ctx context.Context, obj *v1.CaCertificate) error {
	err := c.CaClient.CreateCaCertificate(ctx, obj)
	return err
}

func (c *CaManager) Get(ctx context.Context, name string) (*v1.CaCertificate, error) {
	cfg := config.GetConfig()

	o := client.ObjectKey{
		Namespace: cfg.Namespace,
		Name:      name,
	}
	return c.CaClient.GetCaCertificate(ctx, o)
}

// Filter todo
func (c *CaManager) Filter() []*v1.CaCertificate {

	return []*v1.CaCertificate{}
}

func (c *CaManager) Delete(ctx context.Context, name string) error {
	cfg := config.GetConfig()
	o := client.ObjectKey{
		Namespace: cfg.Namespace,
		Name:      name,
	}
	return c.CaClient.DeleteCaCertificate(ctx, o)
}

// UpsertCaFunc todo 这里要验证一下, 这个接口是什么意思  CaCertificateTransitionFunction
// 这个函数应该是考虑复杂场景的 obj 更新, 就是并非一定是完全替换, 就像 git merge 一样, 可能是有规则的 merge, 就要实现这个函数
func UpsertCaFunc(existing, desired *v1.CaCertificate) error {
	return nil
}

// Update obj 表示更新后的对象, 之前的对象可以通过 key 自己找出来的
func (c *CaManager) Update(ctx context.Context, obj *v1.CaCertificate) error {
	err := c.CaClient.UpsertCaCertificate(ctx, obj, UpsertCaFunc)
	if err != nil {
		logger.Error("ca update err:  %v", err)
	}
	return err
}
