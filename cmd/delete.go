package cmd

import (
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del"},
	Short:   "Delete a custom profile from the keychain",
	Run: func(cmd *cobra.Command, args []string) {
		Verbose(cmd)

		keyChain.Delete(name)
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&name, "name", "", "", "Name of the profile")
	deleteCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(deleteCmd)
}
