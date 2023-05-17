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

type CertManager struct {
	CertClient v1.CertificateClient
}

func (c *CertManager) Init() error {
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
	factory := v12.CertificateClientFromConfigFactoryProvider()
	certClient, err := factory(kubeConnectConfig)
	c.CertClient = certClient
	return err
}

func (c *CertManager) Create(ctx context.Context, obj *v1.Certificate) error {
	err := c.CertClient.CreateCertificate(ctx, obj)
	return err
}

func (c *CertManager) Get(ctx context.Context, name string) (*v1.Certificate, error) {
	cfg := config.GetConfig()

	o := client.ObjectKey{
		Namespace: cfg.Namespace,
		Name:      name,
	}
	return c.CertClient.GetCertificate(ctx, o)
}

func (c *CertManager) Filter(ctx context.Context, fo FilterOptions) ([]v1.Certificate, error) {
	_ = fo
	certificateLs, err := c.CertClient.ListCertificate(ctx)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return certificateLs.Items, nil
}

func (c *CertManager) Delete(ctx context.Context, name string) error {
	cfg := config.GetConfig()
	o := client.ObjectKey{
		Namespace: cfg.Namespace,
		Name:      name,
	}
	return c.CertClient.DeleteCertificate(ctx, o)
}

func UpsertCertFunc(existing, desired *v1.Certificate) error {
	return nil
}

// Update obj 表示更新后的对象, 之前的对象可以通过 key 自己找出来的
func (c *CertManager) Update(ctx context.Context, obj *v1.Certificate) error {
	err := c.CertClient.UpsertCertificate(ctx, obj, UpsertCertFunc)
	if err != nil {
		logger.Error("cert update err:  %v", err)
	}
	return err
}
