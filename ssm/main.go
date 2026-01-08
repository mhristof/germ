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

	// First pass: collect all profiles and prefer admin ones
	allProfiles := ini.GetAll()
	adminProfiles := make(map[string]string) // account-region -> profile name
	otherProfiles := make(map[string]string)
	profileConfigs := make(map[string]map[string]string) // profile name -> config
	
	// Extract account-region key from profile names and group by permission level
	for name, config := range allProfiles {
		if name == "" || name == "default" {
			continue
		}
		profile := strings.TrimPrefix(name, "profile ")
		profileConfigs[profile] = config
		
		// Extract account-region by removing the role part
		var accountRegion string
		if strings.Contains(profile, "AdministratorAccess") {
			// account-prod-AdministratorAccess-ap-northeast-1 -> account-prod-ap-northeast-1
			accountRegion = strings.Replace(profile, "-AdministratorAccess", "", 1)
			adminProfiles[accountRegion] = profile
		} else if strings.Contains(profile, "ReadOnlyAccess") {
			// account-prod-ReadOnlyAccess-ap-northeast-1 -> account-prod-ap-northeast-1  
			accountRegion = strings.Replace(profile, "-ReadOnlyAccess", "", 1)
			if _, hasAdmin := adminProfiles[accountRegion]; !hasAdmin {
				otherProfiles[accountRegion] = profile
			}
		} else {
			// For other roles, find the role part and remove it
			// Pattern: account-{env}-{role}-{region}
			parts := strings.Split(profile, "-")
			if len(parts) < 4 {
				continue
			}
			
			// Find where the region starts (regions contain geographic indicators)
			regionStartIdx := -1
			for i := 2; i < len(parts); i++ { // Start from index 2 (after account-prod/test)
				part := parts[i]
				if strings.Contains(part, "east") || strings.Contains(part, "west") || 
				   strings.Contains(part, "central") || strings.Contains(part, "north") ||
				   strings.Contains(part, "south") || part == "eu" || part == "us" || 
				   part == "ap" || part == "ca" || part == "sa" || part == "af" || part == "me" {
					regionStartIdx = i
					break
				}
			}
			
			if regionStartIdx == -1 {
				continue
			}
			
			// Reconstruct account-region: account + region parts
			accountParts := parts[:2] // account-prod or account-test
			regionParts := parts[regionStartIdx:] // region parts
			accountRegion = strings.Join(accountParts, "-") + "-" + strings.Join(regionParts, "-")
			
			if _, hasAdmin := adminProfiles[accountRegion]; !hasAdmin {
				otherProfiles[accountRegion] = profile
			}
		}
	}
	
	// Use admin profiles where available, fall back to others
	profilesToProcess := make(map[string]map[string]string)
	for _, profileName := range adminProfiles {
		profilesToProcess[profileName] = profileConfigs[profileName]
	}
	for accountRegion, profileName := range otherProfiles {
		// Only add if we don't have an admin profile for this account-region
		if _, hasAdmin := adminProfiles[accountRegion]; !hasAdmin {
			profilesToProcess[profileName] = profileConfigs[profileName]
		}
	}

	for profile, config := range profilesToProcess {
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
	asgs := map[string]string{}

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
			asgName := ""

			for _, tag := range ec2Instance.Reservations[0].Instances[0].Tags {
				log.Trace().Str("key", *tag.Key).Str("value", *tag.Value).Msg("Tag found")

				if *tag.Key == "aws:autoscaling:groupName" {
					firstInstance := asgs[*tag.Value] == ""
					if firstInstance {
						asgs[*tag.Value] = *instance.InstanceId
					}

					asgName = *tag.Value
					log.Debug().
						Bool("first", firstInstance).
						Str("id", *instance.InstanceId).Str("asg", asgName).Msg("ASG found")
				}

				if *tag.Key == "Name" {
					name = *tag.Value
				}
			}

			log.Debug().
				Str("id", *instance.InstanceId).
				Str("name", name).
				Int("tags", len(ec2Instance.Reservations[0].Instances[0].Tags)).
				Msg("Instance found")

			if name == "" {
				log.Debug().Str("id", *instance.InstanceId).Msg("Instance has no name")

				continue
			}

			if asgName != "" && asgs[asgName] != *instance.InstanceId {
				log.Debug().Str("id", *instance.InstanceId).Str("asg", asgName).Msg("ASG already handled")

				continue
			}

			tags := fmt.Sprintf("AWS, %s", accountAlias) + ",account=" + *accountID.Account
			if regionTags, ok := iterm.AWSRegionTags[region]; ok {
				if len(regionTags) > 2 {
					tags += ",region_id=" + regionTags[2]
				}
			}
			bashCommand := fmt.Sprintf("bash -c 'AWS_PROFILE=%s ssm %s'", profile, name)
			config := map[string]string{
				"Initial Text":   bashCommand,
				"Custom Command": "No",
				"Tags":           tags,
			}

			newProfile := iterm.NewProfile(fmt.Sprintf("%s:%s:ssm-%s", accountAlias, region, name), config)

			newProfile.KeyboardMap[iterm.KeyboardSortcutAltA] = iterm.KeyboardMap{
				Action: iterm.KeyboardSendText,
				Text:   fmt.Sprintf("AWS_PROFILE=%s aws sso login && %s\n", profile, bashCommand),
			}

			ret = append(ret, *newProfile)

			log.Info().
				Str("profile", profile).
				Str("region", region).
				Str("instance", name).
				Str("instanceID", *instance.InstanceId).
				Str("asg", asgName).
				Msg("Generated profile")

			instanceIDMutex.Lock()
			instanceIDs[*instance.InstanceId] = name
			instanceIDMutex.Unlock()
		}
	}

	return ret, instanceIDs
}
