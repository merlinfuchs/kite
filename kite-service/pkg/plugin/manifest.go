package plugin

type Manifest struct {
	ID          string
	Name        string
	Description string
	Events      []string
	Permissions []string
	Commands    []ManifestCommand
}

type ManifestCommand struct {
	Type                     string
	Name                     string
	Description              string
	DefaultMemberPermissions []string
	DMPermission             bool
	NSFW                     bool
	Options                  []ManifestCommandOptions
}

type ManifestCommandOptions struct {
	Type        string
	Name        string
	Description string
	Required    bool
	MinValue    int
	MaxValue    int
	MinLength   int
	MaxLength   int
	Choices     []ManifestCommandOptionChoice
	Options     []ManifestCommandOptions
}

type ManifestCommandOptionChoice struct {
	Name  string
	Value string
}
