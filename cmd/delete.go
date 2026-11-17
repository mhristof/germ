package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	deleteName string
)

var deleteCmd = &cobra.Command{
	Use:     "delete [name]",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a custom profile from the keychain",
	Args:    cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completeProfiles(cmd, args, toComplete)
	},
	Run: func(cmd *cobra.Command, args []string) {
		Verbose(cmd)

		if len(args) == 1 {
			deleteName = args[0]
		}

		if deleteName == "" {
			log.Fatal().Msg("a profile name is required (positional arg or --name)")
		}

		keyChain.Delete(deleteName)
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteName, "name", "", "", "Name of the profile")

	rootCmd.AddCommand(deleteCmd)
}
