package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/mhristof/germ/iterm"
	"github.com/mhristof/germ/log"
	"gopkg.in/yaml.v2"
)

func (k *KubeConfig) GetCluster(name string) (*KubeConfig, bool) {
	config := KubeConfig{
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

func GenerateK8sFromAWS(profile string) {
	if profile == "" {
		return
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"profile": profile,
		}).Warning("cannot create aws config")

		return
	}

	client := eks.NewFromConfig(cfg)

	clusters, err := client.ListClusters(context.TODO(), &eks.ListClustersInput{})
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"profile": profile,
		}).Warning("cannot list clusters")

		return
	}

	for _, cluster := range clusters.Clusters {
		command := fmt.Sprintf("aws eks update-kubeconfig --name %s", cluster)
		cmd := exec.Command("bash", "-c", command)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		outStr, errStr := stdout.String(), stderr.String()
		if err != nil {
			log.WithFields(log.Fields{
				"err":     err,
				"command": command,
				"cluster": cluster,
				"outStr":  outStr,
				"errStr":  errStr,
			}).Warning("cannot retrieve kubeconfig for eks cluster")

			continue
		}
	}
}

func (k *KubeConfig) Profiles(dest string, dry bool) []iterm.Profile {
	var ret []iterm.Profile

	for _, cluster := range k.Clusters {
		this, found := k.GetCluster(cluster.Name)
		if !found {
			log.WithFields(log.Fields{
				"cluster.Name": cluster.Name,
			}).Fatal("Cluster not found")
		}

		path := fmt.Sprintf("dry/run/path/%s", this.name())
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
		}).Fatal("Cannot handle multiple cluster definitions")
	}

	return k.Clusters[0].Name
}

func (k *KubeConfig) Profile(path string) *iterm.Profile {
	if len(k.Clusters) != 1 {
		log.WithFields(log.Fields{
			"len(k.Clusters)": len(k.Clusters),
		}).Fatal("Cannot handle multiple cluster definitions")
	}

	tags := map[string]string{
		"Tags": "k8s",
	}
	cmd := fmt.Sprintf("/usr/bin/env KUBECONFIG=%s", path)

	name := filepath.Base(k.Clusters[0].Name)
	awsProfile := k.AWSProfile()
	if awsProfile != "" {
		cmd = fmt.Sprintf("%s AWS_PROFILE=%s", cmd, awsProfile)
		tags["Tags"] += ",aws-profile=" + awsProfile
	}

	user, err := user.Current()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Cannot find current user")
	}

	cmd = fmt.Sprintf("%s /usr/bin/login -fp %s", cmd, user.Username)
	tags["Command"] = cmd
	prof := iterm.NewProfile(fmt.Sprintf("k8s-%s", name), tags)

	return prof
}

func (k *KubeConfig) AWSProfile() string {
	if len(k.Clusters) != 1 {
		log.WithFields(log.Fields{
			"len(k.Clusters)": len(k.Clusters),
		}).Fatal("Cannot handle multiple clusters")
	}

	for _, item := range k.Users[0].User.Exec.Env {
		if item.Name == "AWS_PROFILE" {
			return item.Value
		}
	}
	return ""
}

func Load(config string) *KubeConfig {
	var kConfig KubeConfig

	yamlBytes, err := ioutil.ReadFile(config)
	if err != nil {
		log.WithFields(log.Fields{
			"config": config,
			"err":    err,
		}).Warn("Cannot read file")
		return &kConfig
	}

	err = yaml.Unmarshal(yamlBytes, &kConfig)
	if err != nil {
		log.WithFields(log.Fields{
			"config": config,
			"err":    err,
		}).Fatal("Cannot unmarshal yaml bytes from config")
	}

	return &kConfig
}

func (k *KubeConfig) Print(dest string) string {
	if len(k.Clusters) != 1 {
		log.WithFields(log.Fields{
			"len(k.Clusters)": len(k.Clusters),
		}).Fatal("Cannot handle multiple cluster definitions")
	}

	bytes, err := yaml.Marshal(k)
	destFile := fmt.Sprintf("%s/%s.yml", dest,
		strings.ReplaceAll(
			strings.ReplaceAll(k.Clusters[0].Name, "/", "-"),
			":", "-"),
	)
	err = ioutil.WriteFile(destFile, bytes, 0644)
	if err != nil {
		log.WithFields(log.Fields{
			"destFile": destFile,
			"err":      err,
		}).Fatal("Cannot write to file")
	}

	return destFile
}

func (k *KubeConfig) SplitFiles(dest string) {
	for _, cluster := range k.Clusters {
		this, found := k.GetCluster(cluster.Name)
		if !found {
			log.WithFields(log.Fields{
				"cluster.Name": cluster.Name,
			}).Fatal("Cluster not found")
		}

		bytes, err := yaml.Marshal(this)
		destFile := fmt.Sprintf("%s/%s.yml", dest, cluster.Name)
		err = ioutil.WriteFile(destFile, bytes, 0644)
		if err != nil {
			log.WithFields(log.Fields{
				"destFile": destFile,
				"err":      err,
			}).Fatal("Cannot write to file")
		}
	}
}
