package ergopack

type Ergopack struct {
	Apps map[string]ErgopackApp `yaml:"apps"`
}

type ErgopackApp struct {
	Path          string            `yaml:"path"`
	Image         string            `yaml:"image"`
	PublicPort    string            `yaml:"publicPort"`
	InternalPorts []string          `yaml:"internalPorts"`
	Env           map[string]string `yaml:"env"`
}
