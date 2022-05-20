package cmd

import (
	"os"

	"github.com/mhristof/germ/sso"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var ssoCmd = &cobra.Command{
	Use:   "sso",
	Short: "Gather information about the sso accounts available",
	Run: func(cmd *cobra.Command, args []string) {
		awsProfile := os.Getenv("AWS_PROFILE")

		if awsProfile == "" {
			log.Fatal().Msg("cannt retrieve AWS_PROFILE")
		}

		dir, err := homedir.Expand("~/.aws/config")
		if err != nil {
			log.Fatal().Err(err).Msg("cannot expand aws config path")
		}

		config, err := ini.Load(dir)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot load aws config")
		}

		newConfig := sso.UpdateConfig(config, awsProfile, sso.ListAccounts())

		out, err := cmd.Flags().GetString("out")
		if err != nil {
			panic(err)
		}
		newConfig.SaveTo(out)
	},
}

func init() {
	dir, err := homedir.Expand("~/.aws/config")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot expand aws config path")
	}

	ssoCmd.PersistentFlags().StringP("config", "f", dir, "AWS config profile")
	ssoCmd.PersistentFlags().StringP("out", "o", dir, "output AWS config profile")
	rootCmd.AddCommand(ssoCmd)
}
