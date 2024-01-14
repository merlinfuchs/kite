package postgres

import "github.com/merlinfuchs/kite/kite-service/pkg/model"

func (c *Client) CreatePlugin(plugin *model.Plugin) error {
	return nil
}

func (c *Client) CreatePluginDeployment(deployment *model.PluginDeployment) error {
	return nil
}

func (c *Client) OverridePluginDeployments(deployments []model.PluginDeployment) error {
	return nil
}
