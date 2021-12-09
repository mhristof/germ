package iterm

import (
	"fmt"
	"strings"

	"github.com/mhristof/germ/log"
	"github.com/mitchellh/go-homedir"
)

func notFound(name string) string {
	return fmt.Sprintf("^(bash|/bin/sh): %s: (command )?not found", name)
}

func Triggers() []Trigger {
	idRsa, err := homedir.Expand("~/.ssh/id_rsa")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Panic("Cannot expand ~/.ssh/id_rsa")
	}

	idEd, err := homedir.Expand("~/.ssh/id_ed25519")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Panic("Cannot expand ~/.ssh/id_ed25519")
	}

	return []Trigger{
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
	}
}

func yum(name string) string {
	replacements := map[string]string{
		"openssh-client": "openssh-clients",
	}

	if newName, ok := replacements[name]; ok == true {
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
