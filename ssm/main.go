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
	profilebuilder "github.com/mhristof/germ/profile"
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

	for profileName, config := range profilesToProcess {
		region := config["region"]

		log.Trace().Str("profile", profileName).Str("region", region).Msg("searching")

		wg.Add(1)
		go func(pName, reg string) {
			defer wg.Done()

			profiles, profileInstances := generateForProfile(pName, reg, instances)

			if len(profiles) == 0 {
				failedProfiles = append(failedProfiles, pName)
				return
			}

			lock.Lock()
			defer lock.Unlock()

			ret = append(ret, profiles...)

			log.Debug().Str("profile", pName).Str("region", reg).Int("count", len(profiles)).Msg("Generated profiles")

			for k, v := range profileInstances {
				instances[k] = v
			}
		}(profileName, region)
	}

	wg.Wait()

	if len(failedProfiles) > 0 {
		log.Warn().Str("profiles", strings.Join(failedProfiles, ",")).Msg("Failed to search profiles")
	}

	return ret
}

// create instanceID mutex
var instanceIDMutex = &sync.Mutex{}

// AWSClients holds all the AWS service clients needed for SSM profile generation
type AWSClients struct {
	SSM *awsssm.Client
	EC2 *ec2.Client
	IAM *iam.Client
	STS *sts.Client
}

// AccountInfo holds AWS account information
type AccountInfo struct {
	ID    string
	Alias string
}

// InstanceInfo holds EC2 instance information relevant for SSM profiles
type InstanceInfo struct {
	ID      string
	Name    string
	ASGName string
	Tags    map[string]string
}

func generateForProfile(profile, region string, instanceIDs map[string]string) ([]iterm.Profile, map[string]string) {
	clients, err := createAWSClients(profile)
	if err != nil {
		log.Debug().Err(err).Str("profile", profile).Msg("Failed to create AWS clients")
		return nil, instanceIDs
	}

	accountInfo, err := getAccountInfo(clients)
	if err != nil {
		log.Debug().Err(err).Str("profile", profile).Msg("Failed to retrieve account info")
		return nil, instanceIDs
	}

	instances, err := discoverSSMInstances(clients, instanceIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to discover SSM instances")
		return nil, instanceIDs
	}

	profiles, updatedInstanceIDs := createSSMProfiles(instances, profile, region, accountInfo, instanceIDs)
	
	return profiles, updatedInstanceIDs
}

// createAWSClients initializes all required AWS service clients
func createAWSClients(profile string) (*AWSClients, error) {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return nil, err
	}

	return &AWSClients{
		SSM: awsssm.NewFromConfig(cfg),
		EC2: ec2.NewFromConfig(cfg),
		IAM: iam.NewFromConfig(cfg),
		STS: sts.NewFromConfig(cfg),
	}, nil
}

// getAccountInfo retrieves AWS account ID and alias
func getAccountInfo(clients *AWSClients) (*AccountInfo, error) {
	accountID, err := clients.STS.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil || accountID == nil {
		return nil, err
	}

	accountAliases, err := clients.IAM.ListAccountAliases(context.Background(), &iam.ListAccountAliasesInput{})
	if err != nil {
		return nil, err
	}

	alias := *accountID.Account
	if len(accountAliases.AccountAliases) != 0 {
		alias = accountAliases.AccountAliases[0]
	}

	return &AccountInfo{
		ID:    *accountID.Account,
		Alias: alias,
	}, nil
}

// discoverSSMInstances finds all SSM-managed instances and their details
func discoverSSMInstances(clients *AWSClients, existingInstanceIDs map[string]string) ([]InstanceInfo, error) {
	var instances []InstanceInfo
	asgs := make(map[string]string) // Track ASG instances to avoid duplicates

	paginator := awsssm.NewDescribeInstanceInformationPaginator(clients.SSM, &awsssm.DescribeInstanceInformationInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		for _, instance := range page.InstanceInformationList {
			if shouldSkipInstance(*instance.InstanceId, existingInstanceIDs) {
				continue
			}

			instanceInfo, err := getInstanceDetails(clients.EC2, *instance.InstanceId)
			if err != nil {
				log.Error().Err(err).Str("instanceId", *instance.InstanceId).Msg("Failed to get instance details")
				continue
			}

			if shouldSkipASGInstance(instanceInfo, asgs) {
				continue
			}

			if instanceInfo.Name == "" {
				log.Debug().Str("id", instanceInfo.ID).Msg("Instance has no name")
				continue
			}

			// Track ASG instances
			if instanceInfo.ASGName != "" {
				asgs[instanceInfo.ASGName] = instanceInfo.ID
			}

			instances = append(instances, *instanceInfo)
		}
	}

	return instances, nil
}

// shouldSkipInstance checks if an instance should be skipped (already processed)
func shouldSkipInstance(instanceID string, existingInstanceIDs map[string]string) bool {
	_, found := existingInstanceIDs[instanceID]
	if found {
		log.Debug().Str("id", instanceID).Msg("Instance already found")
		return true
	}
	return false
}

// getInstanceDetails retrieves detailed information about an EC2 instance
func getInstanceDetails(ec2Client *ec2.Client, instanceID string) (*InstanceInfo, error) {
	result, err := ec2Client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return nil, err
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("no instance found with ID %s", instanceID)
	}

	instance := result.Reservations[0].Instances[0]
	info := &InstanceInfo{
		ID:   instanceID,
		Tags: make(map[string]string),
	}

	// Process instance tags
	for _, tag := range instance.Tags {
		if tag.Key == nil || tag.Value == nil {
			continue
		}

		key := *tag.Key
		value := *tag.Value
		info.Tags[key] = value

		log.Trace().Str("key", key).Str("value", value).Msg("Tag found")

		switch key {
		case "aws:autoscaling:groupName":
			info.ASGName = value
			log.Debug().Str("id", instanceID).Str("asg", value).Msg("ASG found")
		case "Name":
			info.Name = value
		}
	}

	log.Debug().
		Str("id", instanceID).
		Str("name", info.Name).
		Int("tags", len(instance.Tags)).
		Msg("Instance found")

	return info, nil
}

// shouldSkipASGInstance checks if an ASG instance should be skipped (not the first one)
func shouldSkipASGInstance(instanceInfo *InstanceInfo, asgs map[string]string) bool {
	if instanceInfo.ASGName == "" {
		return false
	}

	existingInstanceID, exists := asgs[instanceInfo.ASGName]
	if exists && existingInstanceID != instanceInfo.ID {
		log.Debug().
			Str("id", instanceInfo.ID).
			Str("asg", instanceInfo.ASGName).
			Msg("ASG already handled")
		return true
	}

	return false
}

// createSSMProfiles generates iTerm profiles for the discovered instances
func createSSMProfiles(instances []InstanceInfo, profile, region string, accountInfo *AccountInfo, instanceIDs map[string]string) ([]iterm.Profile, map[string]string) {
	var profiles []iterm.Profile
	updatedInstanceIDs := make(map[string]string)

	// Copy existing instance IDs
	for k, v := range instanceIDs {
		updatedInstanceIDs[k] = v
	}

	for _, instance := range instances {
		ssmProfile := createSSMProfile(instance, profile, region, accountInfo)
		profiles = append(profiles, *ssmProfile)

		log.Info().
			Str("profile", profile).
			Str("region", region).
			Str("instance", instance.Name).
			Str("instanceID", instance.ID).
			Str("asg", instance.ASGName).
			Msg("Generated profile")

		// Thread-safe update of instance IDs
		instanceIDMutex.Lock()
		updatedInstanceIDs[instance.ID] = instance.Name
		instanceIDMutex.Unlock()
	}

	return profiles, updatedInstanceIDs
}

// createSSMProfile creates a single SSM profile for an instance
func createSSMProfile(instance InstanceInfo, profile, region string, accountInfo *AccountInfo) *iterm.Profile {
	var regionTags []string
	if tags, ok := iterm.AWSRegionTags[region]; ok {
		regionTags = tags
	}

	return profilebuilder.NewSSMProfileBuilder(accountInfo.Alias, region, instance.Name).
		WithSSMCommand(profile, instance.Name).
		WithAWSAccountInfo(accountInfo.Alias, accountInfo.ID, region, regionTags).
		Build()
}
