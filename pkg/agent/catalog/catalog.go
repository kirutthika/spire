package catalog

import (
	"context"

	"github.com/sirupsen/logrus"
	km_disk "github.com/spiffe/spire/pkg/agent/plugin/keymanager/disk"
	km_memory "github.com/spiffe/spire/pkg/agent/plugin/keymanager/memory"
	na_aws_iid "github.com/spiffe/spire/pkg/agent/plugin/nodeattestor/aws"
	na_azure_msi "github.com/spiffe/spire/pkg/agent/plugin/nodeattestor/azure"
	na_gcp_iit "github.com/spiffe/spire/pkg/agent/plugin/nodeattestor/gcp"
	na_join_token "github.com/spiffe/spire/pkg/agent/plugin/nodeattestor/jointoken"
	na_k8s_psat "github.com/spiffe/spire/pkg/agent/plugin/nodeattestor/k8s/psat"
	na_k8s_sat "github.com/spiffe/spire/pkg/agent/plugin/nodeattestor/k8s/sat"
	na_x509pop "github.com/spiffe/spire/pkg/agent/plugin/nodeattestor/x509pop"
	wa_docker "github.com/spiffe/spire/pkg/agent/plugin/workloadattestor/docker"
	wa_k8s "github.com/spiffe/spire/pkg/agent/plugin/workloadattestor/k8s"
	wa_unix "github.com/spiffe/spire/pkg/agent/plugin/workloadattestor/unix"
	"github.com/spiffe/spire/pkg/common/catalog"
	"github.com/spiffe/spire/proto/agent/keymanager"
	"github.com/spiffe/spire/proto/agent/nodeattestor"
	"github.com/spiffe/spire/proto/agent/workloadattestor"
)

type Catalog interface {
	GetKeyManager() keymanager.KeyManager
	GetNodeAttestor() NodeAttestor
	GetWorkloadAttestors() []WorkloadAttestor
}

type CatalogCloser struct {
	Catalog
	catalog.Closer
}

type GlobalConfig = catalog.GlobalConfig
type HCLPluginConfig = catalog.HCLPluginConfig
type HCLPluginConfigMap = catalog.HCLPluginConfigMap

func KnownPlugins() []catalog.PluginClient {
	return []catalog.PluginClient{
		keymanager.PluginClient,
		nodeattestor.PluginClient,
		workloadattestor.PluginClient,
	}
}

func KnownServices() []catalog.ServiceClient {
	return []catalog.ServiceClient{}
}

func BuiltIns() []catalog.Plugin {
	return []catalog.Plugin{
		km_disk.BuiltIn(),
		km_memory.BuiltIn(),
		na_aws_iid.BuiltIn(),
		na_join_token.BuiltIn(),
		na_gcp_iit.BuiltIn(),
		na_x509pop.BuiltIn(),
		na_azure_msi.BuiltIn(),
		na_k8s_sat.BuiltIn(),
		na_k8s_psat.BuiltIn(),
		wa_k8s.BuiltIn(),
		wa_unix.BuiltIn(),
		wa_docker.BuiltIn(),
	}
}

type NodeAttestor struct {
	catalog.PluginInfo
	nodeattestor.NodeAttestor
}

type WorkloadAttestor struct {
	catalog.PluginInfo
	workloadattestor.WorkloadAttestor
}

type Plugins struct {
	KeyManager        keymanager.KeyManager
	NodeAttestor      NodeAttestor
	WorkloadAttestors []WorkloadAttestor `catalog:"min=1"`
}

var _ Catalog = (*Plugins)(nil)

func (p *Plugins) GetKeyManager() keymanager.KeyManager {
	return p.KeyManager
}

func (p *Plugins) GetNodeAttestor() NodeAttestor {
	return p.NodeAttestor
}

func (p *Plugins) GetWorkloadAttestors() []WorkloadAttestor {
	return p.WorkloadAttestors
}

type Config struct {
	Log          logrus.FieldLogger
	GlobalConfig GlobalConfig
	PluginConfig HCLPluginConfigMap
	HostServices []catalog.HostServiceServer
}

func Load(ctx context.Context, config Config) (*CatalogCloser, error) {
	pluginConfig, err := catalog.PluginConfigFromHCL(config.PluginConfig)
	if err != nil {
		return nil, err
	}

	p := new(Plugins)
	closer, err := catalog.Fill(ctx, catalog.Config{
		Log:           config.Log,
		GlobalConfig:  config.GlobalConfig,
		PluginConfig:  pluginConfig,
		KnownPlugins:  KnownPlugins(),
		KnownServices: KnownServices(),
		BuiltIns:      BuiltIns(),
		HostServices:  config.HostServices,
	}, p)
	if err != nil {
		return nil, err
	}

	return &CatalogCloser{
		Catalog: p,
		Closer:  closer,
	}, nil
}
