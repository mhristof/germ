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

// Item is a germ entry in the keychain. Directory is stored in the item's
// comment attribute and becomes the starting directory of the profile.
type Item struct {
	Account   string
	Directory string
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

// SetDirectory sets the starting directory of an existing item. An empty
// directory clears it. Only the comment attribute is updated, the stored
// secret is left untouched.
func (k *KeyChain) SetDirectory(name, directory string) {
	err := exec.Command("security", "find-generic-password", "-s", k.Service, "-a", name).Run()
	if err != nil {
		log.Fatal().Err(err).Str("name", name).Msg("keychain item does not exist")
	}

	cmd := exec.Command("security", "add-generic-password",
		"-s", k.Service,
		"-a", name,
		"-j", directory,
		"-U")

	err = cmd.Run()
	if err != nil {
		log.Fatal().Err(err).Str("name", name).Msg("failed to set directory")
	}
}

func (k *KeyChain) List() []string {
	var names []string
	for _, item := range k.Items() {
		names = append(names, item.Account)
	}

	return names
}

func (k *KeyChain) Items() []Item {
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
	icmtPrefix = `"icmt"<blob>="`
	svcePrefix = `"svce"<blob>="`
	entryStart = "class:"
)

// parseKeychainDump extracts the account name and directory (comment) of
// every entry whose service matches the given service from
// `security dump-keychain` output.
//
// The dump groups attributes per entry, with each entry starting at a `class:`
// line. The `acct`, `icmt` and `svce` lines can appear in any order within an
// entry and are not at a fixed distance from one another, so we accumulate
// them per entry and emit the item once we know its service matches.
func parseKeychainDump(service, output string) []Item {
	if output == "" {
		return nil
	}

	svceLine := svcePrefix + service + `"`

	var items []Item
	seen := map[string]bool{}

	var item Item
	var serviceMatches bool

	flush := func() {
		if serviceMatches && item.Account != "" && !seen[item.Account] {
			seen[item.Account] = true
			items = append(items, item)
		}
		item = Item{}
		serviceMatches = false
	}

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, entryStart) {
			flush()
			continue
		}

		if strings.HasPrefix(trimmed, acctPrefix) {
			item.Account = strings.TrimSuffix(strings.TrimPrefix(trimmed, acctPrefix), `"`)
			continue
		}

		if strings.HasPrefix(trimmed, icmtPrefix) {
			item.Directory = strings.TrimSuffix(strings.TrimPrefix(trimmed, icmtPrefix), `"`)
			continue
		}

		if trimmed == svceLine {
			serviceMatches = true
		}
	}

	flush()

	return items
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
	for _, item := range k.Items() {
		ret = append(ret, profile(k.Service, item))
	}

	return ret
}

func profile(service string, item Item) iterm.Profile {
	prof := iterm.NewProfile(fmt.Sprintf("custom/%s", item.Account), map[string]string{})

	prof.KeyboardMap[iterm.KeyboardSortcutAltA] = iterm.KeyboardMap{
		Action: 12,
		Text:   loginCommand(service, item.Account),
	}

	if item.Directory != "" {
		prof.CustomDirectory = "Yes"
		prof.WorkingDirectory = item.Directory
	}

	return *prof
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
