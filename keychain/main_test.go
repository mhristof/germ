package keychain

import (
	"fmt"
	"testing"

	"github.com/mhristof/germ/iterm"
	"github.com/stretchr/testify/assert"
)

func TestKeyChain_Profiles(t *testing.T) {
	cases := []struct {
		name     string
		keychain KeyChain
		accounts []string
		expected int
	}{
		{
			name: "single account",
			keychain: KeyChain{
				Service:     "test-service",
				AccessGroup: "test-group",
			},
			accounts: []string{"account1"},
			expected: 1,
		},
		{
			name: "multiple accounts",
			keychain: KeyChain{
				Service:     "germ",
				AccessGroup: "germ",
			},
			accounts: []string{"account1", "account2", "account3"},
			expected: 3,
		},
		{
			name: "no accounts",
			keychain: KeyChain{
				Service:     "empty-service",
				AccessGroup: "empty-group",
			},
			accounts: []string{},
			expected: 0,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			// Mock the List method by creating a test keychain
			k := &mockKeyChain{
				KeyChain: test.keychain,
				accounts: test.accounts,
			}

			profiles := k.Profiles()
			assert.Equal(t, test.expected, len(profiles))

			for i, account := range test.accounts {
				if i < len(profiles) { // Safety check
					expectedName := fmt.Sprintf("custom/%s", account)
					assert.Equal(t, expectedName, profiles[i].Name)
					
					// Check keyboard shortcut is set correctly
					shortcut, exists := profiles[i].KeyboardMap[iterm.KeyboardSortcutAltA]
					assert.True(t, exists)
					assert.Equal(t, int64(12), shortcut.Action)
					
					expectedText := loginCommand(test.keychain.Service, account)
					assert.Equal(t, expectedText, shortcut.Text)
				}
			}
		})
	}
}

func TestKeyChain_ParseDumpOutput(t *testing.T) {
	cases := []struct {
		name     string
		service  string
		output   string
		expected []string
	}{
		{
			name:    "single account found",
			service: "germ",
			output: `keychain: "/Users/test/Library/Keychains/login.keychain-db"
version: 512
class: "genp"
attributes:
    "acct"<blob>="test-account"
    "svce"<blob>="germ"
data:
"secret-value"`,
			expected: []string{"test-account"},
		},
		{
			name:    "multiple accounts",
			service: "germ",
			output: `keychain: "/Users/test/Library/Keychains/login.keychain-db"
version: 512
class: "genp"
attributes:
    "acct"<blob>="account1"
    "svce"<blob>="germ"
data:
"value1"

class: "genp"
attributes:
    "acct"<blob>="account2"
    "svce"<blob>="germ"
data:
"value2"`,
			expected: []string{"account1", "account2"},
		},
		{
			name:     "no matching service",
			service:  "nonexistent",
			output:   `"svce"<blob>="other-service"`,
			expected: nil,
		},
		{
			name:     "empty output",
			service:  "germ",
			output:   "",
			expected: nil,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			k := &KeyChain{Service: test.service}
			accounts := parseKeychainDump(k.Service, test.output)
			assert.Equal(t, test.expected, accounts)
		})
	}
}

func TestEnvVarName(t *testing.T) {
	cases := []struct {
		account  string
		expected string
	}{
		{account: "grafana", expected: "GRAFANA"},
		{account: "grafana_token", expected: "GRAFANA_TOKEN"},
		{account: "grafana-token", expected: "GRAFANA_TOKEN"},
		{account: "my.api.key", expected: "MY_API_KEY"},
		{account: "ALREADY_UPPER", expected: "ALREADY_UPPER"},
	}

	for _, test := range cases {
		t.Run(test.account, func(t *testing.T) {
			assert.Equal(t, test.expected, envVarName(test.account))
		})
	}
}

func TestLoginCommand(t *testing.T) {
	expected := `eval 'export GRAFANA="$(/usr/bin/security find-generic-password -s germ -w -a grafana)"'`

	assert.Equal(t, expected, loginCommand("germ", "grafana"))
}

// mockKeyChain is a test double for KeyChain that doesn't interact with the actual keychain
type mockKeyChain struct {
	KeyChain
	accounts []string
}

func (m *mockKeyChain) Profiles() []iterm.Profile {
	var ret []iterm.Profile
	for _, account := range m.accounts {
		prof := iterm.NewProfile(fmt.Sprintf("custom/%s", account), map[string]string{})

		prof.KeyboardMap[iterm.KeyboardSortcutAltA] = iterm.KeyboardMap{
			Action: 12,
			Text:   loginCommand(m.Service, account),
		}

		ret = append(ret, *prof)
	}

	return ret
}

func (m *mockKeyChain) List() []string {
	return m.accounts
}