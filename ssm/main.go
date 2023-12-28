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
	log "github.com/sirupsen/logrus"
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
		log.WithFields(log.Fields{
			"error":  err,
			"config": config,
		}).Error("Failed to parse AWS config")

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

		log.WithFields(log.Fields{
			"profile": profile,
			"region":  region,
		}).Trace("searching")

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

			log.WithFields(log.Fields{
				"profile": profile,
				"region":  region,
				"count":   len(profiles),
			}).Debug("Generated profiles")

			for k, v := range profileInstances {
				instances[k] = v
			}
		}()
	}

	wg.Wait()

	if len(failedProfiles) > 0 {
		log.WithFields(log.Fields{
			"profiles": failedProfiles,
		}).Warning("Failed to search profiles")
	}

	return ret
}

func generateForProfile(profile, region string, instanceIDs map[string]string) ([]iterm.Profile, map[string]string) {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		log.WithFields(log.Fields{
			"profile": profile,
			"error":   err,
		}).Debug("Failed to load AWS config")

		return nil, instanceIDs
	}

	ssmcli := awsssm.NewFromConfig(cfg)
	ec2cli := ec2.NewFromConfig(cfg)
	iamcli := iam.NewFromConfig(cfg)
	stscli := sts.NewFromConfig(cfg)

	accountID, err := stscli.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil || accountID == nil {
		log.WithFields(log.Fields{
			"error":   err,
			"profile": profile,
		}).Debug("failed to retrieve account id")

		return nil, instanceIDs
	}

	accountAliases, err := iamcli.ListAccountAliases(context.Background(), &iam.ListAccountAliasesInput{})
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"profile": profile,
		}).Debug("failed to retrieve account aliases")

		return nil, instanceIDs
	}

	accountAlias := *accountID.Account
	if len(accountAliases.AccountAliases) != 0 {
		accountAlias = accountAliases.AccountAliases[0]
	}

	instances, err := ssmcli.DescribeInstanceInformation(context.Background(), &awsssm.DescribeInstanceInformationInput{})
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"profile": profile,
		}).Debug("Failed to describe instances")

		return nil, instanceIDs
	}

	ret := []iterm.Profile{}

	for _, instance := range instances.InstanceInformationList {
		ec2Instance, err := ec2cli.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
			InstanceIds: []string{*instance.InstanceId},
		})
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to describe instances")
		}

		_, found := instanceIDs[*instance.InstanceId]
		if found {
			log.WithFields(log.Fields{
				"id": *instance.InstanceId,
			}).Debug("Instance already found")

			continue
		}

		name := ""

		for _, tag := range ec2Instance.Reservations[0].Instances[0].Tags {
			if *tag.Key == "Name" {
				log.WithFields(log.Fields{
					"id":   *instance.InstanceId,
					"name": *tag.Value,
				}).Debug("Instance")

				name = *tag.Value
			}
		}

		log.WithFields(log.Fields{
			"id":   *instance.InstanceId,
			"name": name,
		}).Debug("Instance")

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

		instanceIDs[*instance.InstanceId] = name
	}

	return ret, instanceIDs
}
