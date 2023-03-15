package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var ecrCmd = &cobra.Command{
	Use:   "ecr",
	Short: "Generate ECR credentials helper config",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := exec.LookPath("docker-credential-ecr-login")
		if err != nil {
			log.Fatal().Err(err).Msg("docker-credential-ecr-login not found in PATH. `brew install docker-credential-helper-ecr`")
		}

		configPath, err := homedir.Expand("~/.aws/config")
		if err != nil {
			panic(err)
		}

		config, err := ini.Load(configPath)
		if err != nil {
			panic(err)
		}

		repos := map[string]string{}

		for _, section := range config.Sections() {
			if !strings.HasPrefix(section.Name(), "profile ") {
				continue
			}

			account := section.Key("sso_account_id")
			region := section.Key("sso_region")

			repos[fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", account, region)] = "ecr-login"
		}

		dockerConfig := struct {
			CredHelpers map[string]string `json:"credHelpers"`
		}{
			CredHelpers: repos,
		}

		prettyjson, err := json.MarshalIndent(dockerConfig, "", "  ")
		if err != nil {
			panic(err)
		}

		dockerConfigPath, err := homedir.Expand("~/.docker/config.json")
		if err != nil {
			panic(err)
		}

		_, err = os.Stat(dockerConfigPath)
		if err == nil {
			slurp, err := os.ReadFile(dockerConfigPath)
			if err != nil {
				panic(err)
			}

			// write file
			err = os.WriteFile(dockerConfigPath+".bak", slurp, 0o644)
			if err != nil {
				panic(err)
			}

			log.Info().Str("path", dockerConfigPath).Msg("Docker config file already exists, created backup")
		}

		err = os.WriteFile(dockerConfigPath, prettyjson, 0o644)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(ecrCmd)
}
