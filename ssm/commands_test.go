package ssm

import (
	"testing"

	"github.com/mhristof/germ/iterm"
	"github.com/stretchr/testify/assert"
)

func TestInstanceName(t *testing.T) {
	cases := []struct {
		name     string
		profile  string
		expected string
		ok       bool
	}{
		{"ssm profile", "acc:ap-northeast-1:ssm-marketmaker-apne1-az1-1", "marketmaker-apne1-az1-1", true},
		{"not ssm", "acc:ap-northeast-1:other", "", false},
		{"no colons", "default-profile", "", false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			name, ok := instanceName(test.profile)
			assert.Equal(t, test.ok, ok)
			assert.Equal(t, test.expected, name)
		})
	}
}

func TestApplyCommands(t *testing.T) {
	profiles := []iterm.Profile{
		{Name: "acc:ap-northeast-1:ssm-MarketMaker-1"},
		{Name: "acc:ap-northeast-1:ssm-other"},
	}

	ApplyCommands(profiles, map[string]string{
		"marketmaker-1": "cd ~/mm",
	})

	assert.Equal(t, []iterm.Trigger{{
		Action:    "SendTextTrigger",
		Parameter: "cd ~/mm\\r",
		Regex:     sessionStartRegex,
	}}, profiles[0].Triggers)
	assert.Empty(t, profiles[1].Triggers)
}
