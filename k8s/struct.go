package k8s

type KubeConfig struct {
	APIVersion     string    `yaml:"apiVersion"`
	Clusters       []Cluster `yaml:"clusters"`
	Contexts       []Context `yaml:"contexts"`
	Kind           string    `yaml:"kind"`
	Preferences    struct{}  `yaml:"preferences"`
	Users          []User    `yaml:"users"`
	CurrentContext string    `yaml:"current-context,omitempty"`
}

type Cluster struct {
	Cluster struct {
		CertificateAuthorityData string `yaml:"certificate-authority-data"`
		Server                   string `yaml:"server"`
	} `yaml:"cluster"`
	Name string `yaml:"name"`
}

type Context struct {
	Context struct {
		Cluster string `yaml:"cluster"`
		User    string `yaml:"user"`
	} `yaml:"context"`
	Name string `yaml:"name"`
}

type User struct {
	Name string `yaml:"name"`
	User UserT  `yaml:"user"`
}
type UserT struct {
	Exec Exec `yaml:"exec"`
}

type Exec struct {
	APIVersion string   `yaml:"apiVersion"`
	Args       []string `yaml:"args"`
	Command    string   `yaml:"command"`
	Env        []Env    `yaml:"env"`
}

type Env struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}
