package store

import "github.com/merlinfuchs/kite/kite-service/pkg/model"

type PluginStore interface {
	CreatePlugin(plugin *model.Plugin) error
	CreatePluginDeployment(deployment *model.PluginDeployment) error
	OverridePluginDeployments(deployments []model.PluginDeployment) error
}
