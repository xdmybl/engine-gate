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

type UpstreamManager struct {
	UpstreamClient v1.UpstreamClient
}

func (u *UpstreamManager) Init() error {
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
	factory := v12.UpstreamClientFromConfigFactoryProvider()
	caClient, err := factory(kubeConnectConfig)
	u.UpstreamClient = caClient
	return err
}

func (u *UpstreamManager) Create(ctx context.Context, obj *v1.Upstream) error {
	err := u.UpstreamClient.CreateUpstream(ctx, obj)
	return err
}

func (u *UpstreamManager) Get(ctx context.Context, name string) (*v1.Upstream, error) {
	cfg := config.GetConfig()

	o := client.ObjectKey{
		Namespace: cfg.Namespace,
		Name:      name,
	}
	return u.UpstreamClient.GetUpstream(ctx, o)
}

func (u *UpstreamManager) Filter(ctx context.Context, fo FilterOptions) ([]v1.Upstream, error) {
	_ = fo
	upstreamList, err := u.UpstreamClient.ListUpstream(ctx)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return upstreamList.Items, nil
}

func (u *UpstreamManager) Delete(ctx context.Context, name string) error {
	cfg := config.GetConfig()
	o := client.ObjectKey{
		Namespace: cfg.Namespace,
		Name:      name,
	}
	return u.UpstreamClient.DeleteUpstream(ctx, o)
}

// UpsertUpstreamFunc 这个函数应该是考虑复杂场景的 obj 更新, 就是并非一定是完全替换, 就像 git merge 一样, 可能是有规则的 merge, 就要实现这个函数
func UpsertUpstreamFunc(existing, desired *v1.Upstream) error {
	return nil
}

// Update obj 表示更新后的对象, 之前的对象可以通过 key 自己找出来的
func (u *UpstreamManager) Update(ctx context.Context, obj *v1.Upstream) error {
	err := u.UpstreamClient.UpsertUpstream(ctx, obj, UpsertUpstreamFunc)
	if err != nil {
		logger.Error("upstream update err:  %v", err)
	}
	return err
}
