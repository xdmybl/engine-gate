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

type FilterManager struct {
	FilterClient v1.FilterClient
}

func (f *FilterManager) Init() error {
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
	factory := v12.FilterClientFromConfigFactoryProvider()
	caClient, err := factory(kubeConnectConfig)
	f.FilterClient = caClient
	return err
}

func (f *FilterManager) Create(ctx context.Context, obj *v1.Filter) error {
	err := f.FilterClient.CreateFilter(ctx, obj)
	return err
}

func (f *FilterManager) Get(ctx context.Context, name string) (*v1.Filter, error) {
	cfg := config.GetConfig()

	o := client.ObjectKey{
		Namespace: cfg.Namespace,
		Name:      name,
	}
	return f.FilterClient.GetFilter(ctx, o)
}

func (f *FilterManager) Filter(ctx context.Context, fo FilterOptions) ([]v1.Filter, error) {
	_ = fo
	filterList, err := f.FilterClient.ListFilter(ctx)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return filterList.Items, nil

}

func (f *FilterManager) Delete(ctx context.Context, name string) error {
	cfg := config.GetConfig()
	o := client.ObjectKey{
		Namespace: cfg.Namespace,
		Name:      name,
	}
	return f.FilterClient.DeleteFilter(ctx, o)
}

func UpsertFilterFunc(existing, desired *v1.Filter) error {
	return nil
}

// Update obj 表示更新后的对象, 之前的对象可以通过 key 自己找出来的
func (f *FilterManager) Update(ctx context.Context, obj *v1.Filter) error {
	err := f.FilterClient.UpsertFilter(ctx, obj, UpsertFilterFunc)
	if err != nil {
		logger.Error("filter update err:  %v", err)
	}
	return err
}
