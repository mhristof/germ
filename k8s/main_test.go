package k8s

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/mhristof/germ/iterm"
	"github.com/stretchr/testify/assert"
)

func getUser(t *testing.T) string {
	user, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}

	return user.Username
}

func TestProfile(t *testing.T) {
	cases := []struct {
		name    string
		in      *KubeConfig
		out     *iterm.Profile
		command string
		tags    []string
	}{
		{
			name: "k8s profile with AWS",
			in: &KubeConfig{
				Clusters: []Cluster{
					Cluster{
						Name: "test",
					},
				},
				Users: []User{
					User{
						Name: "test",
						User: UserT{
							Exec: Exec{
								Env: []Env{
									Env{
										Name:  "AWS_PROFILE",
										Value: "profile",
									},
								},
							},
						},
					},
				},
			},
			out:     &iterm.Profile{},
			command: "/usr/bin/env KUBECONFIG=path AWS_PROFILE=profile /usr/bin/login -fp " + getUser(t),
			tags:    []string{"k8s", "aws-profile=profile"},
		},
		{
			name: "k8s profile without AWS, ie minikube",
			in: &KubeConfig{
				Clusters: []Cluster{
					Cluster{
						Name: "test",
					},
				},
				Users: []User{
					User{
						Name: "test",
					},
				},
			},
			out:     &iterm.Profile{},
			command: "/usr/bin/env KUBECONFIG=path /usr/bin/login -fp " + getUser(t),
			tags:    []string{"k8s"},
		},
	}

	for _, test := range cases {
		prof := test.in.Profile("path")
		assert.Equal(t, test.command, prof.Command, test.name)
		assert.Equal(t, test.tags, prof.Tags[1:], test.name)
	}
}

func TestLoadAndSplit(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  map[string]string
	}{
		{
			name: "simple config with a couple of clusters",
			in: heredoc.Doc(`
				apiVersion: v1
				clusters:
				- cluster:
					certificate-authority-data: data
					server: https://kubernetes.docker.internal:6443
				  name: docker-desktop
				- cluster:
					certificate-authority: /path/ca.crt
					extensions:
					- extension:
						last-update: Mon, 01 Mar 2021 16:01:03 GMT
						provider: minikube.sigs.k8s.io
						version: v1.17.1
					  name: cluster_info
					server: https://127.0.0.1:32768
				  name: minikube
				contexts:
				- context:
					cluster: docker-desktop
					user: docker-desktop
				  name: docker-desktop
				- context:
					cluster: minikube
					extensions:
					- extension:
						last-update: Mon, 01 Mar 2021 16:01:03 GMT
						provider: minikube.sigs.k8s.io
						version: v1.17.1
					  name: context_info
					namespace: default
					user: minikube
				  name: minikube
				current-context: minikube
				kind: Config
				preferences: {}
				users:
				- name: docker-desktop
				  user:
					client-certificate-data: data
					client-key-data: data
				- name: minikube
				  user:
					client-certificate: /path/client.crt
					client-key: /path/client.key
			`),
			out: map[string]string{
				"docker-desktop.yml": heredoc.Doc(`
					apiVersion: v1
					clusters:
					- cluster:
					    extensions: []
					    server: ""
					  name: docker-desktop
					contexts:
					- name: docker-desktop
					kind: Config
					preferences: {}
					users:
					- name: docker-desktop
					  user: {}
					current-context: docker-desktop
				`),
				"minikube.yml": heredoc.Doc(`
					apiVersion: v1
					clusters:
					- cluster:
					    extensions: []
					    server: ""
					  name: minikube
					contexts:
					- name: minikube
					kind: Config
					preferences: {}
					users:
					- name: minikube
					  user: {}
					current-context: minikube
				`),
			},
		},
	}

	for _, test := range cases {
		out, err := ioutil.TempDir("", "example")
		if err != nil {
			t.Fatal(err)
		}

		defer os.RemoveAll(out)

		config := filepath.Join(out, "config")

		err = ioutil.WriteFile(config, []byte(noTabs(test.in)), 0644)
		if err != nil {
			t.Fatal(err)
		}

		kConfig := Load(config)
		kConfig.SplitFiles(out)

		for file, content := range test.out {
			data, err := ioutil.ReadFile(filepath.Join(out, file))
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, noTabs(content), string(data), fmt.Sprintf("%s/%s", test.name, file))
		}
	}
}

func noTabs(in string) string {
	return strings.Replace(in, "\t", "  ", -1)
}
