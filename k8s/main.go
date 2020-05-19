package k8s

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"path/filepath"

	"github.com/mhristof/germ/iterm"
	"github.com/mhristof/germ/log"
	"gopkg.in/yaml.v2"
)

func (k *KubeConfig) GetCluster(name string) (*KubeConfig, bool) {
	var config = KubeConfig{
		APIVersion:     k.APIVersion,
		Kind:           k.Kind,
		Preferences:    k.Preferences,
		CurrentContext: name,
	}

	found := false

	for _, cluster := range k.Clusters {
		if cluster.Name == name {
			config.Clusters = append(config.Clusters, cluster)
			found = true
			break
		}
	}

	for _, context := range k.Contexts {
		if context.Name == name {
			config.Contexts = append(config.Contexts, context)
			break
		}
	}

	for _, user := range k.Users {
		if user.Name == name {
			config.Users = append(config.Users, user)
			break
		}
	}

	return &config, found
}

func Profiles(config string, dry bool) []iterm.Profile {
	clusters := Load(config)

	return clusters.Profiles(filepath.Dir(config), dry)
}

func (k *KubeConfig) Profiles(dest string, dry bool) []iterm.Profile {
	var ret []iterm.Profile

	for _, cluster := range k.Clusters {
		this, found := k.GetCluster(cluster.Name)
		if !found {
			log.WithFields(log.Fields{
				"cluster.Name": cluster.Name,
			}).Panic("Cluster not found")
		}

		var path = fmt.Sprintf("dry/run/path/%s", this.name())
		if !dry {
			path = this.Print(dest)
		}
		profile := this.Profile(path)
		ret = append(ret, *profile)
	}

	return ret
}

func (k *KubeConfig) name() string {
	if len(k.Clusters) != 1 {
		log.WithFields(log.Fields{
			"len(k.Clusters)": len(k.Clusters),
		}).Panic("Cannot handle multiple cluster definitions")
	}

	return k.Clusters[0].Name
}

func (k *KubeConfig) Profile(path string) *iterm.Profile {
	if len(k.Clusters) != 1 {
		log.WithFields(log.Fields{
			"len(k.Clusters)": len(k.Clusters),
		}).Panic("Cannot handle multiple cluster definitions")
	}

	name := k.Clusters[0].Name
	awsProfile := k.AWSProfile()
	if awsProfile == "" {
		log.WithFields(log.Fields{
			"awsProfile": awsProfile,
			"cluster":    name,
		}).Panic("Not found in cluster")
	}

	user, err := user.Current()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Panic("Cannot find current user")
	}

	cmd := fmt.Sprintf(
		"/usr/bin/env KUBECONFIG=%s AWS_PROFILE=%s /usr/bin/login -fp %s",
		path,
		awsProfile,
		user.Username,
	)
	prof := iterm.NewProfile(
		fmt.Sprintf("k8s-%s", name),
		map[string]string{
			"Command": cmd,
			"Tags":    fmt.Sprintf("k8s,aws-profile=%s", awsProfile),
		},
	)

	return prof
}

func (k *KubeConfig) AWSProfile() string {
	if len(k.Clusters) != 1 {
		log.WithFields(log.Fields{
			"len(k.Clusters)": len(k.Clusters),
		}).Panic("Cannot handle multiple clusters")
	}

	for _, item := range k.Users[0].User.Exec.Env {
		if item.Name == "AWS_PROFILE" {
			return item.Value
		}
	}
	return ""
}

func Load(config string) *KubeConfig {
	yamlBytes, err := ioutil.ReadFile(config)
	if err != nil {
		log.WithFields(log.Fields{
			"config": config,
			"err":    err,
		}).Panic("Cannot read file")
	}

	var kConfig KubeConfig

	err = yaml.Unmarshal(yamlBytes, &kConfig)
	if err != nil {
		log.WithFields(log.Fields{
			"config": config,
			"err":    err,
		}).Panic("Cannot unmarshal yaml bytes from config")
	}

	return &kConfig
}

func (k *KubeConfig) Print(dest string) string {
	if len(k.Clusters) != 1 {
		log.WithFields(log.Fields{
			"len(k.Clusters)": len(k.Clusters),
		}).Panic("Cannot handle multiple cluster definitions")
	}

	bytes, err := yaml.Marshal(k)
	destFile := fmt.Sprintf("%s/%s.yml", dest, k.Clusters[0].Name)
	err = ioutil.WriteFile(destFile, bytes, 0644)
	if err != nil {
		log.WithFields(log.Fields{
			"destFile": destFile,
			"err":      err,
		}).Panic("Cannot write to file")
	}

	return destFile
}

func (k *KubeConfig) SplitFiles(dest string) {
	for _, cluster := range k.Clusters {
		this, found := k.GetCluster(cluster.Name)
		if !found {
			log.WithFields(log.Fields{
				"cluster.Name": cluster.Name,
			}).Panic("Cluster not found")
		}

		bytes, err := yaml.Marshal(this)
		destFile := fmt.Sprintf("%s/%s.yml", dest, cluster.Name)
		err = ioutil.WriteFile(destFile, bytes, 0644)
		if err != nil {
			log.WithFields(log.Fields{
				"destFile": destFile,
				"err":      err,
			}).Panic("Cannot write to file")
		}
	}

}
