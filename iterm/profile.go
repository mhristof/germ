package iterm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/mhristof/germ/log"
	"github.com/mitchellh/go-homedir"
)

type Profiles struct {
	Profiles []Profile `json:"Profiles"`
}

type Profile struct {
	AllowTitleSetting   bool                   `json:"Allow Title Setting"`
	BadgeText           string                 `json:"Badge Text"`
	Command             string                 `json:"Command"`
	CustomCommand       string                 `json:"Custom Command"`
	CustomDirectory     string                 `json:"Custom Directory"`
	CustomWindowTitle   string                 `json:"Custom Window Title"`
	FlashingBell        bool                   `json:"Flashing Bell"`
	GUID                string                 `json:"Guid"`
	KeyboardMap         map[string]KeyboardMap `json:"Keyboard Map"`
	Name                string                 `json:"Name"`
	SilenceBell         bool                   `json:"Silence Bell"`
	SmartSelectionRules []SmartSelectionRule   `json:"Smart Selection Rules"`
	Tags                []string               `json:"Tags"`
	TitleComponents     int64                  `json:"Title Components"`
	Triggers            []Trigger              `json:"Triggers"`
	UnlimitedScrollback bool                   `json:"Unlimited Scrollback"`
	BackgroundColor     Color                  `json:"Background Color"`
}

type Color struct {
	AlphaComponent float64 `json:"Alpha Component"`
	BlueComponent  float64 `json:"Blue Component"`
	ColorSpace     string  `json:"Color Space"`
	GreenComponent float64 `json:"Green Component"`
	RedComponent   float64 `json:"Red Component"`
}

type KeyboardMap struct {
	Action int64  `json:"Action"`
	Text   string `json:"Text"`
}

type Trigger struct {
	Action    string `json:"action"`
	Parameter string `json:"parameter"`
	Partial   bool   `json:"partial"`
	Regex     string `json:"regex"`
}

type SmartSelectionRuleAction struct {
	Action    int64  `json:"action"`
	Parameter string `json:"parameter"`
	Title     string `json:"title"`
}

type SmartSelectionRule struct {
	Actions   []SmartSelectionRuleAction `json:"actions"`
	Notes     string                     `json:"notes"`
	Precision string                     `json:"precision"`
	Regex     string                     `json:"regex"`
}

func (p *Profiles) Add(prof Profile) {
	p.Profiles = append(p.Profiles, prof)
}

func (p *Profile) HasTag(needle string) bool {
	for _, tag := range p.Tags {
		if tag == needle {
			return true
		}
	}
	return false
}

func (p *Profile) FindTag(key string) (string, bool) {
	var found bool

	for _, tag := range p.Tags {
		parts := strings.Split(tag, "=")
		if len(parts) != 2 {
			continue
		}

		if parts[0] == key {
			return parts[1], true
		}
	}

	return "", found
}

func (p *Profiles) FindGUID(guid string) (Profile, bool) {

	for _, prof := range p.Profiles {
		if prof.GUID == guid {
			return prof, true
		}
	}
	return Profile{}, false
}

func NewProfile(name string, config map[string]string) *Profile {
	var prof = Profile{
		Name:                name,
		GUID:                name,
		Tags:                Tags(config),
		CustomDirectory:     "Recycle",
		SmartSelectionRules: SmartSelectionRules("~/.germ.ssr.json"),
		Triggers:            Triggers(),
		BadgeText:           name,
		TitleComponents:     32,
		CustomWindowTitle:   name,
		AllowTitleSetting:   false,
		FlashingBell:        true,
		SilenceBell:         true,
		KeyboardMap:         CreateKeyboardMap(config),
		UnlimitedScrollback: true,
	}

	v, found := config["Command"]
	if found {
		prof.Command = v
		prof.CustomCommand = "Yes"
	}

	v, found = config["BadgeText"]
	if found {
		prof.BadgeText = v
	}

	v, found = config["AllowTitleSetting"]
	if found {
		value, err := strconv.ParseBool(v)
		if err != nil {
			log.WithFields(log.Fields{
				"v":    v,
				"name": name,
			}).Fatal("Value is not convertable to bool")

		}

		prof.AllowTitleSetting = value
	}

	prof.Colors()
	return &prof
}

func Tags(c map[string]string) []string {
	var tags []string

	if account, ok := c["sso_account_id"]; ok == true {
		tags = append(tags, fmt.Sprintf("account=%s", account))
	}

	v, found := c["source_profile"]
	if found {
		tags = append(tags, fmt.Sprintf("source-profile=%s", v))
		tags = append(tags, v)
	}

	if roleArn, ok := c["role_arn"]; ok == true {
		parts := strings.Split(roleArn, ":")
		tags = append(tags, parts[4])
	}

	v, found = c["azure_app_id_uri"]
	if found {
		parts := strings.Split(v, "#")
		tags = append(tags, parts[1])
	}

	v, found = c["azure_default_role_arn"]
	if found {
		parts := strings.Split(v, ":")
		tags = append(tags, parts[5])
	}

	cTags, found := c["Tags"]
	if found {
		tags = append(tags, strings.Split(cTags, ",")...)
	}

	return tags
}

func SmartSelectionRules(custom string) []SmartSelectionRule {
	var ssr = []SmartSelectionRule{
		{
			Notes:     "shellcheck code",
			Precision: "normal",
			Regex:     `(SC\d*)`,
			Actions: []SmartSelectionRuleAction{
				SmartSelectionRuleAction{
					Title:     "open webpage",
					Action:    1,
					Parameter: `https://github.com/koalaman/shellcheck/wiki/\1`,
				},
			},
		},
		{
			Notes:     "terraform aws resource",
			Precision: "normal",
			Regex:     `resource "aws_([a-zA-Z_]*)"`,
			Actions: []SmartSelectionRuleAction{
				SmartSelectionRuleAction{
					Title:     "open webpage",
					Action:    1,
					Parameter: `https://www.terraform.io/docs/providers/aws/r/\1.html`,
				},
			},
		},
		{
			Notes:     "terraform aws data",
			Precision: "normal",
			Regex:     "data \"aws_([a-zA-Z_]*)\"",
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "open webpage",
					Action:    1,
					Parameter: "https://www.terraform.io/docs/providers/aws/d/\\1.html",
				},
			},
		},
		{
			Notes:     "aws acm-pca",
			Precision: "normal",
			Regex:     "arn:aws:acm-pca:([\\w-]*):(\\d*):certificate-authority/([\\w-]*)",
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "open webpage",
					Action:    1,
					Parameter: "https://\\1.console.aws.amazon.com/acm-pca/home?region=\\1#/certificateAuthorities?arn=arn:aws:acm-pca:\\1:\\2:certificate-authority~2F\\3",
				},
			},
		},
		{
			Notes:     "aws iam-policy",
			Precision: "normal",
			Regex:     "arn:aws:iam::(\\d*):policy/([\\w-]*)",
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "open webpage",
					Action:    1,
					Parameter: "https://console.aws.amazon.com/iam/home?#/policies/arn:aws:iam::\\1:policy/\\2$serviceLevelSummary",
				},
			},
		},
		{
			Notes:     "aws iam-role",
			Precision: "normal",
			Regex:     "arn:aws:iam::\\d*:role/([\\w-_]*)",
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "open webpage",
					Action:    1,
					Parameter: "https://console.aws.amazon.com/iam/home?#/roles/\\1",
				},
			},
		},
		{
			Notes:     "aws lambda",
			Precision: "normal",
			Regex:     "arn:aws:lambda:([\\w-]*):\\d*:function:([\\w-_]*)",
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "open webpage",
					Action:    1,
					Parameter: "https://\\1.console.aws.amazon.com/lambda/home?region=\\1#/functions/\\2?tab=configuration",
				},
			},
		},
	}

	return append(ssr, loadUserSSR(custom)...)
}

func loadUserSSR(path string) []SmartSelectionRule {
	userSsr, err := homedir.Expand(path)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Cannot expand path")
	}

	if _, err := os.Stat(userSsr); os.IsNotExist(err) {
		return []SmartSelectionRule{}
	}

	bytes, err := ioutil.ReadFile(userSsr)
	if err != nil {
		log.WithFields(log.Fields{
			"userSsr": userSsr,
			"err":     err,
		}).Fatal("Cannot read file")
	}

	var userSSRs []SmartSelectionRule

	err = json.Unmarshal(bytes, &userSSRs)
	if err != nil {
		log.WithFields(log.Fields{
			"string(bytes)": string(bytes),
			"userSsr":       userSsr,
			"err":           err,
		}).Fatal("Cannot parse json file")
	}

	return userSSRs
}

func CreateKeyboardMap(config map[string]string) map[string]KeyboardMap {
	var maps = map[string]KeyboardMap{
		"0x5f-0x120000": KeyboardMap{
			Action: 25,
			Text:   "Split Horizontally with Current Profile\nSplit Horizontally with Current Profile",
		},
		"0x7c-0x120000": KeyboardMap{
			Action: 25,
			Text:   "Split Vertically with Current Profile\nSplit Vertically with Current Profile",
		},
	}

	v, found := config["source_profile"]
	if found {
		maps["0x61-0x80000"] = KeyboardMap{
			Action: 28,
			Text:   fmt.Sprintf("login-%s", v),
		}
	}

	return maps
}

func (p *Profiles) UpdateAWSSmartSelectionRules() {
	accounts := map[string]string{}

	for _, profile := range p.Profiles {
		for _, tag := range profile.Tags {
			if !strings.HasPrefix(tag, "account=") {
				continue
			}

			if strings.HasPrefix(profile.Name, "login-") {
				continue
			}

			accounts[profile.Name] = strings.Split(tag, "=")[1]
		}
	}

	var ssr []SmartSelectionRule
	for name, id := range accounts {
		ssr = append(ssr, SmartSelectionRule{
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "Notify the AWS account name",
					Action:    2,
					Parameter: fmt.Sprintf("osascript -e 'display notification \"%s\" with title \"%s\"'", name, id),
				},
			},
			Notes:     fmt.Sprintf("AWS account ID for %s", name),
			Precision: "very_high",
			Regex:     id,
		})
	}

	for i, _ := range p.Profiles {
		p.Profiles[i].SmartSelectionRules = append(p.Profiles[i].SmartSelectionRules, ssr...)
	}

}

func (p *Profiles) UpdateKeyboardMaps() {
	for _, profile := range p.Profiles {
		if !profile.HasTag("k8s") {
			continue
		}

		awsProfile, found := profile.FindTag("aws-profile")

		if !found {
			continue
		}

		fmt.Println(fmt.Sprintf("awsProfile: %+v", awsProfile))

		sourceProfile, found := p.FindGUID(awsProfile)

		if !found {
			log.WithFields(log.Fields{
				"awsProfile": awsProfile,
				"k8s":        profile.GUID,
			}).Error("AWS Profile not found")
		}

		key := "0x61-0x80000"

		profile.KeyboardMap[key] = sourceProfile.KeyboardMap[key]

	}
}

func (p *Profiles) SourceProfiles() []string {
	var ret []string

	for _, profile := range p.Profiles {
		isSource := true
		for _, tag := range profile.Tags {
			if strings.HasPrefix(tag, "source-profile=") {
				isSource = false
			}
		}
		if isSource {
			ret = append(ret, profile.GUID)
		}
	}
	return ret
}

func (p *Profiles) ProfileTree() map[string][]string {
	var ret = map[string][]string{}

	for _, profile := range p.Profiles {
		for _, tag := range profile.Tags {
			if strings.HasPrefix(tag, "source-profile=") {
				parts := strings.Split(tag, "=")
				if _, ok := ret[parts[1]]; ok != true {
					ret[parts[1]] = make([]string, 1)
					ret[parts[1]][0] = profile.GUID
				} else {
					ret[parts[1]] = append(ret[parts[1]], profile.GUID)
				}
			}
		}
	}

	return ret
}

func isProd(name string) bool {
	if strings.Contains(name, "nonprd") {
		return false
	}

	if strings.Contains(name, "nonprod") {
		return false
	}

	if strings.Contains(name, "prod") {
		return true
	}

	if strings.Contains(name, "prd") {
		return true
	}
	return false
}

func (p *Profile) Colors() {
	if isProd(p.Name) {
		p.BackgroundColor.ColorSpace = "sRGB"
		p.BackgroundColor.RedComponent = 0.217376708984375
		p.BackgroundColor.GreenComponent = 0
		p.BackgroundColor.BlueComponent = 0
		p.BackgroundColor.AlphaComponent = 1
		return
	}
	if p.HasTag("k8s") {
		p.BackgroundColor.RedComponent = 0
		p.BackgroundColor.ColorSpace = "sRGB"
		p.BackgroundColor.BlueComponent = 0.38311767578125
		p.BackgroundColor.AlphaComponent = 1
		p.BackgroundColor.GreenComponent = 0
		return
	}
	p.BackgroundColor.ColorSpace = "sRGB"
	p.BackgroundColor.RedComponent = 0
	p.BackgroundColor.GreenComponent = 0
	p.BackgroundColor.BlueComponent = 0
	p.BackgroundColor.AlphaComponent = 1
}
