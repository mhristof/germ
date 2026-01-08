package aws

import (
	"fmt"
	"testing"

	"github.com/mhristof/germ/iterm"
	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	cases := []struct {
		name     string
		config   []map[string]string
		expected []iterm.Profile
	}{
		{
			name: "single profile",
			config: []map[string]string{
				{
					"foo": "bar",
				},
			},
			expected: []iterm.Profile{
				{
					GUID: "0",
				},
				{
					GUID: "login-0",
				},
			},
		},
		{
			name: "profile with source_profile",
			config: []map[string]string{
				{
					"source_profile": "bar",
				},
			},
			expected: []iterm.Profile{
				{
					GUID: "0",
				},
			},
		},
	}

	for _, test := range cases {
		profiles := []iterm.Profile{}
		for i, cfg := range test.config {
			name := fmt.Sprintf("%d", i)
			
			// Create main profile
			mainProfile := createAWSProfile("", name, cfg)
			profiles = append(profiles, *mainProfile)
			
			// Create login profile if needed
			if loginProfile := createLoginProfile(name, cfg); loginProfile != nil {
				profiles = append(profiles, *loginProfile)
			}
		}

		assert.Equal(t, len(test.expected), len(profiles))

		for _, expectedProfile := range test.expected {
			found := false
			for _, profile := range profiles {
				if profile.GUID == expectedProfile.GUID {
					found = true
					break
				}
			}
			assert.True(t, found)
		}

	}
}
