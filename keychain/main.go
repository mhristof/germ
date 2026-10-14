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

	return parseKeychainDump(k.Service, string(output))
}

const (
	acctPrefix = `"acct"<blob>="`
	svcePrefix = `"svce"<blob>="`
	entryStart = "class:"
)

// parseKeychainDump extracts the account names of every entry whose service
// matches the given service from `security dump-keychain` output.
//
// The dump groups attributes per entry, with each entry starting at a `class:`
// line. The `acct` and `svce` lines can appear in any order within an entry and
// are not at a fixed distance from one another, so we accumulate both per entry
// and emit the account once we know its service matches.
func parseKeychainDump(service, output string) []string {
	if output == "" {
		return nil
	}

	svceLine := svcePrefix + service + `"`

	var accounts []string
	seen := map[string]bool{}

	var account string
	var serviceMatches bool

	flush := func() {
		if serviceMatches && account != "" && !seen[account] {
			seen[account] = true
			accounts = append(accounts, account)
		}
		account = ""
		serviceMatches = false
	}

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, entryStart) {
			flush()
			continue
		}

		if strings.HasPrefix(trimmed, acctPrefix) {
			account = strings.TrimSuffix(strings.TrimPrefix(trimmed, acctPrefix), `"`)
			continue
		}

		if trimmed == svceLine {
			serviceMatches = true
		}
	}

	flush()

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
			Text:   loginCommand(k.Service, account),
		}

		ret = append(ret, *prof)

	}

	return ret
}

// loginCommand builds the shortcut that, when evaluated, exports the raw token
// stored in the keychain as an environment variable named after the account
// (uppercased). The token itself is read at eval time, so only the raw secret
// lives in the keychain, never the surrounding `export` wrapping.
func loginCommand(service, account string) string {
	variable := envVarName(account)

	return fmt.Sprintf(
		`eval 'export %s="$(/usr/bin/security find-generic-password -s %s -w -a %s)"'`,
		variable, service, account,
	)
}

// envVarName converts an account name into a valid shell environment variable
// name: uppercased, with any character that is not a letter, digit or
// underscore replaced by an underscore.
func envVarName(account string) string {
	var b strings.Builder

	for _, r := range strings.ToUpper(account) {
		switch {
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}

	return b.String()
}
