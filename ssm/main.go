package ssm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/mhristof/germ/iterm"
	"github.com/rs/zerolog/log"
	"github.com/zieckey/goini"
)

func expandUser(path string) string {
	if strings.HasPrefix(path, "~/") {
		path = strings.Replace(path, "~", os.Getenv("HOME"), 1)
	}

	return path
}

func Generate() []iterm.Profile {
	ini := goini.New()
	config := expandUser("~/.aws/config")
	err := ini.ParseFile(config)
	if err != nil {
		log.Error().Err(err).Str("config", config).Msg("Failed to parse AWS config")
		return nil
	}

	ret := []iterm.Profile{}
	instances := map[string]string{}
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}

	failedProfiles := []string{}

	for name, config := range ini.GetAll() {
		if name == "" {
			continue
		}
		profile := strings.TrimPrefix(name, "profile ")
		region := config["region"]

		log.Trace().Str("profile", profile).Str("region", region).Msg("searching")

		wg.Add(1)
		go func() {
			defer wg.Done()

			profiles, profileInstances := generateForProfile(profile, region, instances)

			if len(profiles) == 0 {
				failedProfiles = append(failedProfiles, profile)
				return
			}

			lock.Lock()
			defer lock.Unlock()

			ret = append(ret, profiles...)

			log.Debug().Str("profile", profile).Str("region", region).Int("count", len(profiles)).Msg("Generated profiles")

			for k, v := range profileInstances {
				instances[k] = v
			}
		}()
	}

	wg.Wait()

	if len(failedProfiles) > 0 {
		log.Warn().Str("profiles", strings.Join(failedProfiles, ",")).Msg("Failed to search profiles")
	}

	return ret
}

// create instanceID mutex
var instanceIDMutex = &sync.Mutex{}

func generateForProfile(profile, region string, instanceIDs map[string]string) ([]iterm.Profile, map[string]string) {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		log.Debug().Err(err).Str("profile", profile).Msg("Failed to load AWS config")

		return nil, instanceIDs
	}

	ssmcli := awsssm.NewFromConfig(cfg)
	ec2cli := ec2.NewFromConfig(cfg)
	iamcli := iam.NewFromConfig(cfg)
	stscli := sts.NewFromConfig(cfg)

	accountID, err := stscli.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil || accountID == nil {
		log.Debug().Err(err).Str("profile", profile).Msg("Failed to retrieve account id")

		return nil, instanceIDs
	}

	accountAliases, err := iamcli.ListAccountAliases(context.Background(), &iam.ListAccountAliasesInput{})
	if err != nil {
		log.Debug().Err(err).Str("profile", profile).Msg("Failed to retrieve account aliases")

		return nil, instanceIDs
	}

	accountAlias := *accountID.Account
	if len(accountAliases.AccountAliases) != 0 {
		accountAlias = accountAliases.AccountAliases[0]
	}

	ret := []iterm.Profile{}
	asgs := map[string]struct{}{}

	// get all instances with a paginator
	paginator := awsssm.NewDescribeInstanceInformationPaginator(ssmcli, &awsssm.DescribeInstanceInformationInput{})

	for paginator.HasMorePages() {
		instances, err := paginator.NextPage(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("Failed to describe instances")
			return nil, instanceIDs
		}

		for _, instance := range instances.InstanceInformationList {
			ec2Instance, err := ec2cli.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
				InstanceIds: []string{*instance.InstanceId},
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to describe instances")
			}

			_, found := instanceIDs[*instance.InstanceId]
			if found {
				log.Debug().Str("id", *instance.InstanceId).Msg("Instance already found")

				continue
			}

			name := ""

			for _, tag := range ec2Instance.Reservations[0].Instances[0].Tags {
				_, found := asgs[*tag.Value]
				if found {
					// If the ASG group has already been processed, skip setting
					// the instance name so the instance is not processed again.
					log.Debug().Str("id", *instance.InstanceId).Str("asg", *tag.Value).Msg("Instance already found")
					continue
				}

				if *tag.Key == "aws:autoscaling:groupName" {
					asgs[*tag.Value] = struct{}{}
				}

				if *tag.Key == "Name" {
					log.Debug().Str("id", *instance.InstanceId).Str("name", *tag.Value).Msg("Instance found")

					name = *tag.Value
				}
			}

			log.Debug().Str("id", *instance.InstanceId).Str("name", name).Msg("Instance found")

			if name == "" {
				log.Debug().Str("id", *instance.InstanceId).Msg("Instance has no name")

				continue
			}

			bashCommand := fmt.Sprintf("bash -c 'AWS_PROFILE=%s ssm %s'", profile, name)
			config := map[string]string{
				"Initial Text":   bashCommand,
				"Custom Command": "No",
				"Tags":           fmt.Sprintf("AWS, %s", accountAlias) + ",account=" + *accountID.Account,
			}

			newProfile := iterm.NewProfile(fmt.Sprintf("%s:%s:ssm-%s", accountAlias, region, name), config)

			newProfile.KeyboardMap[iterm.KeyboardSortcutAltA] = iterm.KeyboardMap{
				Action: iterm.KeyboardSendText,
				Text:   fmt.Sprintf("AWS_PROFILE=%s aws sso login && %s\n", profile, bashCommand),
			}

			ret = append(ret, *newProfile)

			log.Info().Str("profile", profile).Str("region", region).Str("instance", name).Str("instanceID", *instance.InstanceId).Msg("Generated profile")

			instanceIDMutex.Lock()
			instanceIDs[*instance.InstanceId] = name
			instanceIDMutex.Unlock()
		}
	}

	return ret, instanceIDs
}
