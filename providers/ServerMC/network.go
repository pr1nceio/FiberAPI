package ServerMC

import (
	"github.com/fruitspace/schemas/db/go/db"
	kubejconfig "go.minekube.com/gate/pkg/edition/java/config"
	kubeliteconfig "go.minekube.com/gate/pkg/edition/java/lite/config"
	kubeconfig "go.minekube.com/gate/pkg/gate/config"
	"go.minekube.com/gate/pkg/util/configutil"
	"gopkg.in/yaml.v3"
)

type Network struct {
	p            *ServerMCProvider
	network      *db.MinecraftNetwork
	routerConfig db.MinecraftRouterConfig
}

func (n *Network) Fetch(uuid string) error {
	err := n.p.db.Model(&db.MinecraftNetwork{}).Preload("MinecraftServers").
		Where("uuid=?", uuid).First(&n.network).Error
	if err == nil {
		n.routerConfig, err = n.network.GetRouterConfig()
	}
	return err
}

// ------
func (n *Network) exportRouterConfig() (string, error) {
	conf := kubejconfig.Config{}
	conf.Bind = "0.0.0.0:25565"
	lite := conf.Lite
	lite.Enabled = true

	motd := configutil.TextComponent{}
	err := motd.UnmarshalJSON([]byte(n.routerConfig.OfflineMessage.Motd))
	if err != nil {
		return "", err
	}

	route := kubeliteconfig.Route{
		Host:    []string{n.routerConfig.Host},
		Backend: n.routerConfig.Backends,
		Fallback: &kubeliteconfig.Status{
			MOTD: &motd,
		},
		ProxyProtocol:   n.routerConfig.ProxyProtocol,
		TCPShieldRealIP: n.routerConfig.TCPShieldProtocol,
	}
	lite.Routes = append(lite.Routes, route)

	conf.Lite = lite
	conf.ProxyProtocol = true
	//FIXME: Can be disruptive. Subject to test
	conf.Quota.Connections = kubejconfig.QuotaSettings{
		Enabled:    true,
		OPS:        5,
		Burst:      10,
		MaxEntries: 1000,
	}

	globalConf := kubeconfig.Config{}
	globalConf.Config = conf

	bytes, err := yaml.Marshal(&globalConf)

	return string(bytes), err
}
