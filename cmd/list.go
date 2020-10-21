package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the keychain profiles",
	Run: func(cmd *cobra.Command, args []string) {
		Verbose(cmd)

		fmt.Println(keyChain.List())
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
