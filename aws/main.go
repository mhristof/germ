package aws

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/mhristof/germ/iterm"
	"github.com/rs/zerolog/log"
	"github.com/zieckey/goini"
)

func Profiles(prefix, config string) []iterm.Profile {
	ini := goini.New()
	err := ini.ParseFile(config)
	if err != nil {
		log.Warn().Err(err).Str("config", config).Msg("parse INI failed")
		return nil
	}

	var prof iterm.Profiles
	for name, section := range ini.GetAll() {
		if name == "" {
			continue
		}
		tName := strings.TrimPrefix(name, "profile ")
		add(&prof, prefix, fmt.Sprintf("%s", tName), section)
	}

	return prof.Profiles
}

func add(p *iterm.Profiles, prefix, name string, config map[string]string) {
	user, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot find current user")
	}

	config["Command"] = fmt.Sprintf("/usr/bin/env AWS_PROFILE=%s /usr/bin/login -fp %s", name, user.Username)
	pName := name

	if prefix != "" {
		pName = fmt.Sprintf("%s-%s", prefix, name)
	}

	profile := iterm.NewProfile(pName, config)
	p.Add(*profile)

	_, sourceProfile := config["source_profile"]
	_, sso := config["sso_account_id"]

	if !sourceProfile && !sso {
		config["Command"] = loginCmd(name, config)
		loginProfile := iterm.NewProfile(fmt.Sprintf("login-%s", name), config)

		log.Debug().
			Str("profile.GUID", profile.GUID).
			Str("loginProfile.GUID", loginProfile.GUID).
			Msg("create login profile")

		p.Add(*loginProfile)
	}
}

func loginCmd(name string, config map[string]string) string {
	var tool, toolCmd string
	_, azure := config["azure_tenant_id"]
	_, ssoAccountId := config["sso_account_id"]

	if azure {
		tool = "aws-azure-login"
		toolCmd = fmt.Sprintf("%s --no-prompt", tool)
	} else if ssoAccountId {
		tool = "aws"
		toolCmd = "aws sso login"
	} else {
		return ""
	}

	bin, err := exec.LookPath(tool)
	if err != nil {
		log.Fatal().Err(err).
			Str("tool", tool).
			Msg("cannot find executable")
	}

	return fmt.Sprintf(
		"bash -c 'AWS_PROFILE=%s PATH=%s NODE_EXTRA_CA_CERTS=%s %s || sleep 60'",
		name, filepath.Dir(bin), os.Getenv("NODE_EXTRA_CA_CERTS"), toolCmd,
	)
}

// Regions retrieve all AWS regions. This list is generated from
// https://docs.aws.amazon.com/general/latest/gr/rande.html
func Regions() []string {
	return []string{
		"us-east-2",
		"us-east-1",
		"us-west-1",
		"us-west-2",
		"af-south-1",
		"ap-east-1",
		"ap-south-1",
		"ap-northeast-3",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
		"ca-central-1",
		"cn-north-1",
		"cn-northwest-1",
		"eu-central-1",
		"eu-west-1",
		"eu-west-2",
		"eu-south-1",
		"eu-west-3",
		"eu-north-1",
		"me-south-1",
		"sa-east-1",
	}
}
