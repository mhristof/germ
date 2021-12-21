package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/mhristof/germ/aws"
	"github.com/mhristof/germ/iterm"
	"github.com/mhristof/germ/k8s"
	"github.com/mhristof/germ/log"
	"github.com/mhristof/germ/vim"

	//"github.com/mhristof/germ/vim"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	output         string
	write          bool
	kubeConfig     string
	diff           bool
	AWSConfig      = expandUser("~/.aws/config")
	AWSCredentials = expandUser("~/.aws/credentials")
	DefaultProfile = "default-profile"
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
			}).Fatal("--write is incompatible with --dry-run")
		}

		if write && diff {
			log.WithFields(log.Fields{
				"write": write,
				"diff":  diff,
			}).Fatal("--write and --diff are incompatible")
		}

		var prof iterm.Profiles

		prof.Profiles = append(prof.Profiles, aws.Profiles("", AWSConfig)...)
		// prof.Profiles = append(prof.Profiles, aws.Profiles("credentials", AWSCredentials)...)
		prof.Profiles = append(prof.Profiles, k8s.Profiles(kubeConfig, dryRun)...)
		prof.Profiles = append(prof.Profiles, keyChain.Profiles()...)
		prof.Profiles = append(prof.Profiles, *iterm.NewProfile(DefaultProfile, map[string]string{
			"AllowTitleSetting": "true",
			"BadgeText":         "",
		}))
		prof.Profiles = append(prof.Profiles, vim.Profile())
		prof.UpdateKeyboardMaps()
		prof.UpdateAWSSmartSelectionRules()

		profJSON, err := json.MarshalIndent(prof, "", "    ")
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatal("Cannot indent json results")
		}

		// unescape "&" character.
		profJSON = []byte(strings.ReplaceAll(string(profJSON), `\u0026`, "&"))

		if write {
			err = ioutil.WriteFile(output, profJSON, 0644)
			if err != nil {
				log.WithFields(log.Fields{
					"output": output,
					"err":    err,
				}).Fatal("Cannot write to file")
			}
		} else if diff {
			curr, err := ioutil.ReadFile(output)
			if err != nil {
				log.WithFields(log.Fields{
					"err":    err,
					"output": output,
				}).Fatal("Cannot read file")
			}

			var current iterm.Profiles
			err = json.Unmarshal(curr, &current)
			if err != nil {
				log.WithFields(log.Fields{
					"err":    err,
					"output": output,
				}).Fatal("Cannot unmarshal output file")
			}

			sort.Slice(current.Profiles, func(i, j int) bool {
				return current.Profiles[i].GUID < current.Profiles[j].GUID
			})

			sort.Slice(prof.Profiles, func(i, j int) bool {
				return prof.Profiles[i].GUID < prof.Profiles[j].GUID
			})

			if diff := cmp.Diff(current, prof); diff != "" {
				fmt.Println("Updating (-current +new):", diff)
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
		}).Fatal("Cannot expand homedir")
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
		AWSConfig,
		"AWS config file path",
	)
	generateCmd.Flags().StringVarP(
		&AWSCredentials, "aws-credentials", "c",
		AWSCredentials,
		"AWS credentials file path",
	)
	generateCmd.Flags().StringVarP(
		&kubeConfig, "kube-config", "k",
		expandUser("~/.kube/config"),
		"Kubernetes configuration file",
	)
	generateCmd.Flags().BoolVarP(&write, "write", "w", false, "Write the output to the destination file")
	generateCmd.Flags().BoolVarP(&diff, "diff", "d", false, "Generate a diff for the new changes")

	rootCmd.AddCommand(generateCmd)
}
