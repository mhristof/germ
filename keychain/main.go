package keychain

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/mhristof/germ/iterm"
	"github.com/rs/zerolog/log"
)

type KeyChain struct {
	Service     string
	AccessGroup string
}

func (k *KeyChain) Add(name, value string) {
	// Use the security command-line tool instead of deprecated APIs
	cmd := exec.Command("security", "add-generic-password", 
		"-s", k.Service, 
		"-a", name, 
		"-w", value,
		"-U") // -U updates if exists
	
	err := cmd.Run()
	if err != nil {
		log.Fatal().Err(err).Str("name", name).Msg("failed to add keychain item")
	}
}

func (k *KeyChain) List() []string {
	// Use security command to list accounts
	cmd := exec.Command("security", "dump-keychain")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal().Err(err).Str("k.Service", k.Service).Msg("cannot retrieve the accounts")
	}
	
	var accounts []string
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		// Look for lines containing our service and extract account names
		if strings.Contains(line, fmt.Sprintf(`"svce"<blob>="%s"`, k.Service)) {
			// Find the account name in the same keychain entry
			for i, searchLine := range lines {
				if searchLine == line {
					// Look for the account line near this service line
					for j := i - 5; j < i + 5 && j < len(lines) && j >= 0; j++ {
						if strings.Contains(lines[j], `"acct"<blob>=`) {
							// Extract account name from: "acct"<blob>="account_name"
							start := strings.Index(lines[j], `"acct"<blob>="`) + 14
							end := strings.LastIndex(lines[j], `"`)
							if start < end && start > 13 {
								account := lines[j][start:end]
								accounts = append(accounts, account)
							}
							break
						}
					}
					break
				}
			}
		}
	}
	
	return accounts
}

func (k *KeyChain) Delete(name string) {
	log.Debug().Str("name", name).Msg("deleting keychain object")

	cmd := exec.Command("security", "delete-generic-password", "-s", k.Service, "-a", name)
	err := cmd.Run()
	if err != nil {
		log.Fatal().Err(err).Str("name", name).Msg("cannot delete item")
	}
}

func (k *KeyChain) Profiles() []iterm.Profile {
	var ret []iterm.Profile
	for _, account := range k.List() {
		prof := iterm.NewProfile(fmt.Sprintf("custom/%s", account), map[string]string{})

		prof.KeyboardMap[iterm.KeyboardSortcutAltA] = iterm.KeyboardMap{
			Action: 12,
			Text:   fmt.Sprintf("eval $(/usr/bin/security find-generic-password  -s %s -w -a %s)", k.Service, account),
		}

		ret = append(ret, *prof)

	}

	return ret
}
