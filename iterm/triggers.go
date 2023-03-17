package iterm

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func notFound(name string) string {
	return fmt.Sprintf("^(bash|/bin/sh): %s: (command )?not found", name)
}

func profileTriggers(profile string) []Trigger {
	file, err := homedir.Expand(fmt.Sprintf("~/.germ.trigger.%s.json", profile))
	if err != nil {
		return []Trigger{}
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return []Trigger{}
	}

	var ret []Trigger

	err = json.Unmarshal(data, &ret)
	if err != nil {
		panic(err)
	}

	return ret
}

func Triggers(profile string) []Trigger {
	idRsa, err := homedir.Expand("~/.ssh/id_rsa")
	if err != nil {
		log.Panic().Err(err).Msg("cannot expand ~/")
	}

	idEd, err := homedir.Expand("~/.ssh/id_ed25519")
	if err != nil {
		log.Panic().Err(err).Msg("cannot expand ~/")
	}

	var ret []Trigger

	if viper.GetBool("aws_open_device_sso") {
		ret = append(ret, Trigger{
			Regex:     "^([A-Z]{4}-[A-Z]{4})",
			Action:    "MuteCoprocessTrigger",
			Parameter: `open "https://device.sso.ap-southeast-1.amazonaws.com/?user_code=\1"`,
		})
	}

	return append(ret, []Trigger{
		{
			Regex:     "^# timed out waiting for input: auto-logout",
			Action:    "SendTextTrigger",
			Parameter: "top\\r",
		},
		{
			Partial:   true,
			Parameter: "id_rsa",
			Regex:     fmt.Sprintf(`^Enter passphrase for (key ')?%s`, idRsa),
			Action:    "PasswordTrigger",
		},
		{
			Partial:   true,
			Parameter: "id_ed25519",
			Regex:     fmt.Sprintf(`^Enter passphrase for (key ')?%s`, idEd),
			Action:    "PasswordTrigger",
		},
		{
			Action:    "SendTextTrigger",
			Parameter: apt("openssh-client"),
			Regex:     notFound("ssh-add"),
		},
		{
			Action:    "SendTextTrigger",
			Parameter: apt("git"),
			Regex:     notFound("git"),
		},
		{
			Action:    "SendTextTrigger",
			Parameter: apt("iputils-ping"),
			Regex:     notFound("ping"),
		},
		{
			Action:    "SendTextTrigger",
			Parameter: "terraform init",
			Regex:     `^This module is not yet installed. Run "terraform init" to install all modules`,
		},
		{
			Action:    "SendTextTrigger",
			Parameter: "chmod +x !:0 && !!",
			Regex:     `^zsh: permission denied: .*`,
		},
		{
			Action:    "SendTextTrigger",
			Parameter: "git push --set-upstream origin $(git rev-parse --abbrev-ref HEAD)",
			Regex:     "^To push the current branch and set the remote as upstream",
		},
	}...)
}

func yum(name string) string {
	replacements := map[string]string{
		"openssh-client": "openssh-clients",
	}

	if newName, ok := replacements[name]; ok {
		name = newName
	}

	return fmt.Sprintf("(yum install --assumeyes %s)", name)
}

func apk(name string) string {
	return fmt.Sprintf("apk add --no-cache %s", name)
}

func apt(name string) string {
	commands := []string{
		fmt.Sprintf("(apt-get update && apt-get --yes --no-install-recommends install %s)", name),
		yum(name),
		apk(name),
	}

	return strings.Join(commands, " || ")
}
