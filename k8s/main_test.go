package k8s

import (
	"testing"

	"github.com/mhristof/germ/iterm"
	"github.com/stretchr/testify/assert"
)

func TestProfile(t *testing.T) {
	var cases = []struct {
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
			command: "/usr/bin/env KUBECONFIG=path AWS_PROFILE=profile /usr/bin/login -fp mhristof",
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
			command: "/usr/bin/env KUBECONFIG=path /usr/bin/login -fp mhristof",
			tags:    []string{"k8s"},
		},
	}

	for _, test := range cases {
		prof := test.in.Profile("path")
		assert.Equal(t, test.command, prof.Command, test.name)
		assert.Equal(t, test.tags, prof.Tags, test.name)
		//assert.Equal(t, test.out, test.in.Profile("path"), test.name)
	}
}
