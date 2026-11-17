package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:     "set [name]",
	Aliases: []string{"update-token"},
	Short:   "Update the secret for an existing profile. The value is read via a hidden prompt so it never lands in the shell history",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		Verbose(cmd)

		if len(args) == 1 {
			newName = args[0]
		}

		if newName == "" {
			log.Fatal().Msg("a profile name is required (positional arg or --name)")
		}

		keyChain.Add(newName, findPassword(file))
	},
}

func init() {
	setCmd.Flags().StringVarP(&newName, "name", "", "", "Name of the profile")

	rootCmd.AddCommand(setCmd)
}
