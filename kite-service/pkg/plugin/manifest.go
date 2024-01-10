package plugin

type PluginManifest struct {
	ID          string
	Events      []string
	Permissions []string
}
