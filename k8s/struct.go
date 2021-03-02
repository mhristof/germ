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
		CertificateAuthority     string `yaml:"certificate-authority,omitempty"`
		CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
		Extensions               []struct {
			Extension struct {
				Lastupdate string `yaml:"last-update"`
				Provider   string `yaml:"provider"`
				Version    string `yaml:"version"`
			} `yaml:"extension"`
			Name string `yaml:"name"`
		} `yaml:"extensions"`
		Server string `yaml:"server"`
	} `yaml:"cluster"`
	Name string `yaml:"name"`
}

type Context struct {
	Context struct {
		Cluster    string `yaml:"cluster"`
		Extensions []struct {
			Extension struct {
				Lastupdate string `yaml:"last-update,omitempty"`
				Provider   string `yaml:"provider,omitempty"`
				Version    string `yaml:"version,omitempty"`
			} `yaml:"extension,omitempty"`
			Name string `yaml:"name"`
		} `yaml:"extensions,omitempty"`
		Namespace string `yaml:"namespace"`
		User      string `yaml:"user"`
	} `yaml:"context,omitempty"`
	Name string `yaml:"name"`
}

type User struct {
	Name string `yaml:"name"`
	User UserT  `yaml:"user"`
}

type UserT struct {
	ClientCertificate     string `yaml:"client-certificate,omitempty"`
	ClientCertificateData string `yaml:"client-certificate-data,omitempty"`
	ClientKey             string `yaml:"client-key,omitempty"`
	ClientKeyData         string `yaml:"client-key-data,omitempty"`
	Exec                  Exec   `yaml:"exec,omitempty"`
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
