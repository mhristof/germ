package cmd

import (
	"strings"

	"github.com/mhristof/germ/k8s"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var kubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Fetch kubeconfig for all available AWS clusters",
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := homedir.Expand("~/.aws/config")
		if err != nil {
			panic(err)
		}

		config, err := ini.Load(dir)
		if err != nil {
			log.Panic().Err(err).Msg("cannot load aws config")
		}

		for _, section := range config.Sections() {
			if !strings.HasSuffix(section.Name(), "AdministratorAccess") {
				continue
			}
			log.Debug().Str("section.Name()", section.Name()).Msg("retrieving k8s clusters")

			k8s.GenerateK8sFromAWS(strings.ReplaceAll(section.Name(), "profile ", ""))
		}
	},
}

func init() {
	rootCmd.AddCommand(kubeconfigCmd)
}
