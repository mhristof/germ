package ssm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProfileSelection(t *testing.T) {
	cases := []struct {
		name           string
		profiles       map[string]map[string]string // profile name -> config
		expectedChoice map[string]string            // account-region -> expected profile name
	}{
		{
			name: "admin profile preferred over readonly",
			profiles: map[string]map[string]string{
				"account-prod-AdministratorAccess-us-east-1": {
					"region": "us-east-1",
				},
				"account-prod-ReadOnlyAccess-us-east-1": {
					"region": "us-east-1",
				},
			},
			expectedChoice: map[string]string{
				"account-prod-us-east-1": "account-prod-AdministratorAccess-us-east-1",
			},
		},
		{
			name: "admin profile preferred over other roles",
			profiles: map[string]map[string]string{
				"account-prod-AdministratorAccess-ap-northeast-1": {
					"region": "ap-northeast-1",
				},
				"account-prod-serverless-dev-ap-northeast-1": {
					"region": "ap-northeast-1",
				},
				"account-prod-deltix-user-ap-northeast-1": {
					"region": "ap-northeast-1",
				},
			},
			expectedChoice: map[string]string{
				"account-prod-ap-northeast-1": "account-prod-AdministratorAccess-ap-northeast-1",
			},
		},
		{
			name: "fallback to other role when no admin",
			profiles: map[string]map[string]string{
				"account-prod-serverless-dev-eu-central-1": {
					"region": "eu-central-1",
				},
				"account-prod-ReadOnlyAccess-eu-central-1": {
					"region": "eu-central-1",
				},
			},
			expectedChoice: map[string]string{
				"account-prod-eu-central-1": "account-prod-serverless-dev-eu-central-1",
			},
		},
		{
			name: "multi-part regions handled correctly",
			profiles: map[string]map[string]string{
				"account-prod-AdministratorAccess-eu-central-1": {
					"region": "eu-central-1",
				},
				"account-prod-serverless-dev-eu-central-1": {
					"region": "eu-central-1",
				},
			},
			expectedChoice: map[string]string{
				"account-prod-eu-central-1": "account-prod-AdministratorAccess-eu-central-1",
			},
		},
		{
			name: "different environments handled separately",
			profiles: map[string]map[string]string{
				"account-prod-AdministratorAccess-us-west-2": {
					"region": "us-west-2",
				},
				"account-test-AdministratorAccess-us-west-2": {
					"region": "us-west-2",
				},
				"account-prod-serverless-dev-us-west-2": {
					"region": "us-west-2",
				},
			},
			expectedChoice: map[string]string{
				"account-prod-us-west-2": "account-prod-AdministratorAccess-us-west-2",
				"account-test-us-west-2": "account-test-AdministratorAccess-us-west-2",
			},
		},
		{
			name: "readonly as last resort",
			profiles: map[string]map[string]string{
				"account-prod-ReadOnlyAccess-ap-east-1": {
					"region": "ap-east-1",
				},
			},
			expectedChoice: map[string]string{
				"account-prod-ap-east-1": "account-prod-ReadOnlyAccess-ap-east-1",
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result := selectProfiles(test.profiles)

			// Verify expectations
			for expectedAccountRegion, expectedProfile := range test.expectedChoice {
				found := false
				for selectedProfile := range result {
					if selectedProfile == expectedProfile {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected profile %s for account-region %s not found in selection", expectedProfile, expectedAccountRegion)
			}

			// Verify no unexpected profiles are selected
			assert.Equal(t, len(test.expectedChoice), len(result), "Number of selected profiles doesn't match expected")
		})
	}
}

// selectProfiles extracts the profile selection logic for testing
func selectProfiles(allProfiles map[string]map[string]string) map[string]map[string]string {
	adminProfiles := make(map[string]string)
	otherProfiles := make(map[string]string)
	profileConfigs := make(map[string]map[string]string)

	for profile, config := range allProfiles {
		profileConfigs[profile] = config

		var accountRegion string
		if strings.Contains(profile, "AdministratorAccess") {
			accountRegion = strings.Replace(profile, "-AdministratorAccess", "", 1)
			adminProfiles[accountRegion] = profile
		} else if strings.Contains(profile, "ReadOnlyAccess") {
			accountRegion = strings.Replace(profile, "-ReadOnlyAccess", "", 1)
			if _, hasAdmin := adminProfiles[accountRegion]; !hasAdmin {
				otherProfiles[accountRegion] = profile
			}
		} else {
			parts := strings.Split(profile, "-")
			if len(parts) < 4 {
				continue
			}

			regionStartIdx := -1
			for i := 2; i < len(parts); i++ {
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

			accountParts := parts[:2]
			regionParts := parts[regionStartIdx:]
			accountRegion = strings.Join(accountParts, "-") + "-" + strings.Join(regionParts, "-")

			if _, hasAdmin := adminProfiles[accountRegion]; !hasAdmin {
				otherProfiles[accountRegion] = profile
			}
		}
	}

	// Build final selection
	profilesToProcess := make(map[string]map[string]string)
	for _, profileName := range adminProfiles {
		profilesToProcess[profileName] = profileConfigs[profileName]
	}
	for accountRegion, profileName := range otherProfiles {
		if _, hasAdmin := adminProfiles[accountRegion]; !hasAdmin {
			profilesToProcess[profileName] = profileConfigs[profileName]
		}
	}

	return profilesToProcess
}

func TestRegionDetection(t *testing.T) {
	cases := []struct {
		name           string
		profileName    string
		expectedRegion string
		shouldMatch    bool
	}{
		{
			name:           "single part region",
			profileName:    "account-prod-serverless-dev-us-1",
			expectedRegion: "us-1",
			shouldMatch:    true,
		},
		{
			name:           "multi part region",
			profileName:    "account-prod-role-eu-central-1",
			expectedRegion: "eu-central-1",
			shouldMatch:    true,
		},
		{
			name:           "complex region",
			profileName:    "account-test-complex-role-ap-northeast-1",
			expectedRegion: "ap-northeast-1",
			shouldMatch:    true,
		},
		{
			name:           "no region found",
			profileName:    "account-prod-role-without-region",
			expectedRegion: "",
			shouldMatch:    false,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			parts := strings.Split(test.profileName, "-")
			regionStartIdx := -1

			for i := 2; i < len(parts); i++ {
				part := parts[i]
				if strings.Contains(part, "east") || strings.Contains(part, "west") ||
					strings.Contains(part, "central") || strings.Contains(part, "north") ||
					strings.Contains(part, "south") || part == "eu" || part == "us" ||
					part == "ap" || part == "ca" || part == "sa" || part == "af" || part == "me" {
					regionStartIdx = i
					break
				}
			}

			if test.shouldMatch {
				assert.NotEqual(t, -1, regionStartIdx, "Should have found region start index")
				if regionStartIdx != -1 {
					regionParts := parts[regionStartIdx:]
					actualRegion := strings.Join(regionParts, "-")
					assert.Equal(t, test.expectedRegion, actualRegion, "Region extraction mismatch")
				}
			} else {
				assert.Equal(t, -1, regionStartIdx, "Should not have found region start index")
			}
		})
	}
}