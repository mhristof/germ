package sso

import (
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/assert"
	"gopkg.in/ini.v1"
)

func TestUpdateConfig(t *testing.T) {
	cases := []struct {
		name           string
		accounts       []Account
		config         []byte
		profile        string
		expectedConfig []byte
	}{
		{
			name:    "all new configs",
			profile: "existingProfile",
			config: []byte(heredoc.Doc(`
				[profile existingProfile]
				foo = bar
			`)),
			accounts: []Account{
				{
					AccountName: "test",
					AccountID:   "1234",
					Role:        "role",
				},
			},
			expectedConfig: []byte(heredoc.Doc(`
				[profile existingProfile]
				foo = bar

				; autogenerated by germ with <3
				[profile test-role]
				; inherited from existingProfile
				foo            = bar
				sso_account_id = 1234
				sso_role_name  = role
			`)),
		},
		{
			name:    "profile to add exists",
			profile: "existingProfile",
			config: []byte(heredoc.Doc(`
				[profile existingProfile]
				sso_role_name  = existingRole

				[profile test-role]
				this = that
			`)),
			accounts: []Account{
				{
					AccountName: "test",
					AccountID:   "1234",
					Role:        "role",
				},
			},
			expectedConfig: []byte(heredoc.Doc(`
				[profile existingProfile]
				sso_role_name  = existingRole

				[profile test-role]
				this           = that
				; autogenerated by germ with <3
				sso_account_id = 1234
				; autogenerated by germ with <3
				sso_role_name  = role
			`)),
		},
	}

	for _, test := range cases {
		config, err := ini.Load(test.config)
		assert.Nil(t, err, test.name)

		expectedConfig, err := ini.Load(test.expectedConfig)

		assert.Nil(t, err, test.name)
		var expected strings.Builder
		_, err = expectedConfig.WriteTo(&expected)
		assert.Nil(t, err, test.name)

		var generated strings.Builder
		_, err = UpdateConfig(config, test.profile, test.accounts).WriteTo(&generated)
		assert.Nil(t, err, test.name)

		assert.Equal(t, expected.String(), generated.String(), test.name)
	}
}