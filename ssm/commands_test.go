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
	trigger := func(command string) []iterm.Trigger {
		return []iterm.Trigger{{
			Action:    "SendTextTrigger",
			Parameter: command + "\\r",
			Regex:     sessionStartRegex,
		}}
	}

	profiles := []iterm.Profile{
		{Name: "acc:ap-northeast-1:ssm-marketmaker-apne1-az1-1"},
		{Name: "acc:ap-northeast-1:ssm-marketmaker-apne1-az2-1"},
		{Name: "acc:ap-northeast-1:ssm-MarketMaker-euw1-1"},
		{Name: "acc:ap-northeast-1:ssm-other"},
	}

	ApplyCommands(profiles, []Command{
		{Match: "[", Command: "invalid"},
		{Match: "^marketmaker-apne1-az1-1$", Command: "exact"},
		{Match: "^marketmaker-apne1-.*", Command: "regex"},
		{Match: "(?i)^marketmaker-", Command: "case insensitive"},
	})

	assert.Equal(t, trigger("exact"), profiles[0].Triggers)
	assert.Equal(t, trigger("regex"), profiles[1].Triggers)
	assert.Equal(t, trigger("case insensitive"), profiles[2].Triggers)
	assert.Empty(t, profiles[3].Triggers)
}
