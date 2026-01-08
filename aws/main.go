package aws

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mhristof/germ/iterm"
	"github.com/mhristof/germ/profile"
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

	var profiles []iterm.Profile
	for name, section := range ini.GetAll() {
		if name == "" {
			continue
		}
		tName := strings.TrimPrefix(name, "profile ")
		
		// Create main profile
		mainProfile := createAWSProfile(prefix, tName, section)
		profiles = append(profiles, *mainProfile)
		
		// Create login profile if needed
		if loginProfile := createLoginProfile(tName, section); loginProfile != nil {
			profiles = append(profiles, *loginProfile)
		}
	}

	return profiles
}

func createAWSProfile(prefix, name string, config map[string]string) *iterm.Profile {
	builder := profile.NewAWSProfileBuilder(name).
		WithAWSProfile(name).
		WithPrefix(prefix)
	
	// Add any additional config from the section
	for key, value := range config {
		builder.WithConfig(key, value)
	}
	
	return builder.Build()
}

func createLoginProfile(name string, config map[string]string) *iterm.Profile {
	_, sourceProfile := config["source_profile"]
	_, sso := config["sso_account_id"]

	// Only create login profile if it's not a source profile or SSO profile
	if sourceProfile || sso {
		return nil
	}
	
	loginCmd := buildLoginCommand(name, config)
	if loginCmd == "" {
		// If no specific login command, create a basic login profile
		// This maintains compatibility with the original behavior
		loginCmd = fmt.Sprintf("echo 'No login command configured for %s'", name)
	}
	
	builder := profile.NewAWSProfileBuilder(fmt.Sprintf("login-%s", name)).
		WithAWSLoginCommand(name, loginCmd)
	
	// Add any additional config from the section
	for key, value := range config {
		builder.WithConfig(key, value)
	}
	
	log.Debug().
		Str("profile", name).
		Str("loginProfile", fmt.Sprintf("login-%s", name)).
		Msg("create login profile")
	
	return builder.Build()
}

func buildLoginCommand(name string, config map[string]string) string {
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
