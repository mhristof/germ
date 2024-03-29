package cmd

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"strings"
	"syscall"

	"github.com/mhristof/germ/keychain"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/ini.v1"
)

var (
	newName     string
	exportAlias string
	value       string
	file        string
	keyChain    = keychain.KeyChain{
		Service:     "germ",
		AccessGroup: "germ",
	}
	exported bool
)

var newCmd = &cobra.Command{
	Use:     "new",
	Short:   "Create new profile for the given secret. The system will be entered via a prompt to avoid storing it in the cmd history",
	Aliases: []string{"add"},
	Run: func(cmd *cobra.Command, args []string) {
		Verbose(cmd)

		keyChain.Add(newName, findPassword(file))
	},
}

func findPassword(file string) string {
	if file != "" {
		return handleFile(file)
	}

	fmt.Print("Enter secret:")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal().Err(err).Msg("cannot read secret")
	}

	if exported {
		value := fmt.Sprintf("export %s='%s'", strings.ToUpper(newName), string(bytePassword))

		if exportAlias != "" {
			value = fmt.Sprintf("%s %s='%s'", value, strings.ToUpper(exportAlias), string(bytePassword))
		}

		bytePassword = []byte(value)
	}

	return string(bytePassword)
}

func loadRootKey(file string) (string, error) {
	cfg, err := ini.Load(file)
	if err != nil {
		return "", errors.Wrap(err, "cannot read ini file")
	}

	keyID, err := cfg.Section("").GetKey("AWSAccessKeyId")
	if err != nil {
		return "", errors.Wrap(err, "could not find AWSAccessKeyId key")
	}

	secretKey, err := cfg.Section("").GetKey("AWSSecretKey")
	if err != nil {
		return "", errors.Wrap(err, "could not find AWSSecretKey key")
	}

	return exportAWS(keyID.String(), secretKey.String()), nil
}

func loadCredentials(file string) (string, error) {
	records := slurpCsv(file)

	if len(records[0]) < 4 {
		return "", errors.New("invalid number of columns")
	}

	if records[0][2] != "Access key ID" {
		return "", errors.New("invalid header for AWS creds file")
	}

	if records[0][3] != "Secret access key" {
		return "", errors.New("invalid header for AWS creds file")
	}

	return exportAWS(records[1][2], records[1][3]), nil
}

func loadAccessKeys(file string) (string, error) {
	records := slurpCsv(file)

	if records[0][0][0] == 239 {
		// unicode chars are in the csv file and we need to trim them
		// records[0][0]: "\ufeffAccess key ID" string
		records[0][0] = records[0][0][3:]
	}

	if records[0][0] != "Access key ID" {
		return "", errors.New("invalid header for AWS creds file")
	}

	if records[0][1] != "Secret access key" {
		return "", errors.New("invalid header for AWS creds file")
	}

	return exportAWS(records[1][0], records[1][1]), nil
}

func handleFile(file string) string {
	funcs := []func(string) (string, error){
		loadRootKey,
		loadCredentials,
		loadAccessKeys,
	}

	for _, f := range funcs {
		ret, err := f(file)
		if err == nil {
			return ret
		}
	}

	log.Fatal().Str("file", file).Msg("cannot handle this type of file")
	return ""
}

func slurpCsv(file string) [][]string {
	in, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot read file")
	}
	r := csv.NewReader(strings.NewReader(string(in)))

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot read csv file")
	}

	return records
}

func exportAWS(access, secret string) string {
	return fmt.Sprintf("export AWS_ACCESS_KEY_ID=%s AWS_SECRET_ACCESS_KEY=%s", access, secret)
}

func init() {
	newCmd.Flags().StringVarP(&newName, "name", "", "", "Name of the profile")
	newCmd.Flags().StringVarP(&file, "file", "f", "", "Credentials file to parse")
	newCmd.Flags().BoolVarP(&exported, "export", "e", false, "Treat the password as an exported variable. The name of the variable will be the uppercased name provided.")
	newCmd.Flags().StringVarP(&exportAlias, "alias", "a", "", "environment variable alias for the exported variable")
	newCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(newCmd)
}
