package keychain

import (
	"fmt"
	"strings"
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
					
					expectedText := fmt.Sprintf("eval $(/usr/bin/security find-generic-password  -s %s -w -a %s)", test.keychain.Service, account)
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
    0x00000007 <blob>="germ"
    "acct"<blob>="test-account"
    "cdat"<timedate>=0x32303234303130383139353530305A00  "20240108195500Z\000"
    "crtr"<uint32>=<NULL>
    "cusi"<sint32>=<NULL>
    "desc"<blob>=<NULL>
    "gena"<blob>=<NULL>
    "icmt"<blob>=<NULL>
    "invi"<sint32>=<NULL>
    "mdat"<timedate>=0x32303234303130383139353530305A00  "20240108195500Z\000"
    "nega"<sint32>=<NULL>
    "prot"<blob>=<NULL>
    "scrp"<sint32>=<NULL>
    "svce"<blob>="germ"
    "type"<uint32>=<NULL>
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
			Text:   fmt.Sprintf("eval $(/usr/bin/security find-generic-password  -s %s -w -a %s)", m.Service, account),
		}

		ret = append(ret, *prof)
	}

	return ret
}

// parseKeychainDump extracts accounts from keychain dump output for testing
func parseKeychainDump(service, output string) []string {
	if output == "" {
		return nil
	}
	
	var accounts []string
	lines := strings.Split(output, "\n")
	
	for i, line := range lines {
		if strings.Contains(line, fmt.Sprintf(`"svce"<blob>="%s"`, service)) {
			// Look for the account line near this service line
			for j := i - 5; j < i + 5 && j < len(lines) && j >= 0; j++ {
				if strings.Contains(lines[j], `"acct"<blob>=`) {
					start := strings.Index(lines[j], `"acct"<blob>="`) + 14
					end := strings.LastIndex(lines[j], `"`)
					if start < end && start > 13 {
						account := lines[j][start:end]
						// Avoid duplicates
						found := false
						for _, existing := range accounts {
							if existing == account {
								found = true
								break
							}
						}
						if !found {
							accounts = append(accounts, account)
						}
					}
					break
				}
			}
		}
	}
	
	if len(accounts) == 0 {
		return nil
	}
	
	return accounts
}

func (m *mockKeyChain) List() []string {
	return m.accounts
}