package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/mhristof/germ/aws"
	"github.com/mhristof/germ/iterm"
	"github.com/mhristof/germ/k8s"
	"github.com/mhristof/germ/log"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	output     string
	write      bool
	kubeConfig string
	AWSConfig  string
)

var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generate the profiles",
	Aliases: []string{"gen"},
	Run: func(cmd *cobra.Command, args []string) {
		Verbose(cmd)

		if write && dryRun {
			log.WithFields(log.Fields{
				"write":  write,
				"dryrun": dryRun,
			}).Panic("--write is incompatible with --dry-run")
		}

		var prof iterm.Profiles

		prof.Profiles = append(prof.Profiles, aws.Profiles(AWSConfig)...)
		prof.Profiles = append(prof.Profiles, k8s.Profiles(kubeConfig, dryRun)...)
		prof.Profiles = append(prof.Profiles, keyChain.Profiles()...)
		prof.Profiles = append(prof.Profiles, *iterm.NewProfile("default", map[string]string{
			"BadgeText": "",
		}))
		prof.UpdateKeyboardMaps()

		profJSON, err := json.MarshalIndent(prof, "", "    ")
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Panic("Cannot indent json results")
		}

		if write {
			err = ioutil.WriteFile(output, profJSON, 0644)
			if err != nil {
				log.WithFields(log.Fields{
					"output": output,
					"err":    err,
				}).Panic("Cannot write to file")
			}
		} else {
			fmt.Println(string(profJSON))
		}
	},
}

func expandUser(path string) string {
	out, err := homedir.Expand(path)
	if err != nil {
		log.WithFields(log.Fields{
			"path": path,
			"err":  err,
		}).Panic("Cannot expand homedir")
	}
	return out
}

func init() {
	generateCmd.Flags().StringVarP(
		&output, "output", "o",
		expandUser("~/Library/Application Support/iTerm2/DynamicProfiles/aws-profiles.json"),
		"File to save the generated profiles",
	)
	generateCmd.Flags().StringVarP(
		&AWSConfig, "aws-config", "a",
		expandUser("~/.aws/config"),
		"AWS config file path",
	)
	generateCmd.Flags().StringVarP(
		&kubeConfig, "kube-config", "k",
		expandUser("~/.kube/config"),
		"Kubernetes configuration file",
	)
	generateCmd.Flags().BoolVarP(&write, "write", "w", false, "Write the output to the destination file")

	rootCmd.AddCommand(generateCmd)
}
