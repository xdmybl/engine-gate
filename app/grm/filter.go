package grm

import (
	"context"
	"github.com/wonderivan/logger"
	"github.com/xdmybl/engine-gate/util/config"
	"github.com/xdmybl/engine-gate/util/constant"
	v1 "github.com/xdmybl/gate-type/pkg/api/gate.xdmybl.io/v1"
	v12 "github.com/xdmybl/gate-type/pkg/api/gate.xdmybl.io/v1/providers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FilterManager struct {
	FilterClient v1.FilterClient
}

func (f *FilterManager) Init() error {
	cfg := config.GetConfig()
	// TODO 加载 kubeconfig 或 静态 token, master url, 等配置
	// 优先加载 kubeconfig
	kubeConnectConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Error("kube connect init err:  %v", err)
	}
	if cfg.KubeConfig != "" {
		kubeConnectConfig, err = clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
	} else if cfg.Token != "" && cfg.MasterUrl != "" {
		kubeConnectConfig.BearerToken = cfg.Token
		kubeConnectConfig.Host = cfg.MasterUrl
	}
	if err != nil {
		logger.Error("kube config err:  %v", err)
		os.Exit(constant.KubernetesConnectError)
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

// Filter todo
func (f *FilterManager) Filter() []*v1.Filter {

	return []*v1.Filter{}
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
