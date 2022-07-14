package iterm

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog/log"
)

func SmartSelectionRules(custom string) []SmartSelectionRule {
	ssr := []SmartSelectionRule{
		{
			Notes:     "gitlab terraform source",
			Precision: "normal",
			Regex:     `git@gitlab.com:(.*)\.git//(.*)\?`,
			Actions: []SmartSelectionRuleAction{
				SmartSelectionRuleAction{
					Title:     "open webpage",
					Action:    1,
					Parameter: `https://gitlab.com/\1/-/tree/master/\2`,
				},
			},
		},
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
		{
			Notes:     "git restore --staged",
			Precision: "normal",
			Regex:     "^\\s*modified:\\s*(.*)",
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "git restore",
					Action:    4,
					Parameter: "git restore --staged \\1;",
				},
			},
		},
		{
			Notes:     "git add",
			Precision: "normal",
			Regex:     "^\\s*both modified:\\s*(.*)",
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "git add",
					Action:    4,
					Parameter: "git add \\1;",
				},
			},
		},
		{
			Notes:     "aws ec2 descripbe-images",
			Precision: "normal",
			Regex:     "(ami-[0-9a-f]*)",
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "aws ec2 describe-images",
					Action:    4,
					Parameter: " aws ec2 describe-images --image-ids \\1\n",
				},
			},
		},
		{
			Notes:     "aws ec2 descripbe-instances",
			Precision: "normal",
			Regex:     "(i-[0-9a-f]*)",
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "aws ec2 describe-instances",
					Action:    4,
					Parameter: " aws ec2 describe-instances --instance-ids \\1\n",
				},
			},
		},
		{
			Notes:     "aws ec2 descripbe-vpcs",
			Precision: "normal",
			Regex:     "(vpc-[0-9a-f]*)",
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "aws ec2 describe-vpcs",
					Action:    4,
					Parameter: " aws ec2 describe-vpcs --vpc-ids \\1\n",
				},
			},
		},
		{
			Notes:     "aws ec2 describe-security-groups",
			Precision: "normal",
			Regex:     "(sg-[0-9a-f]*)",
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "aws ec2 describe-security-groups",
					Action:    4,
					Parameter: "jq --slurp '.[0] + .[1]' <(aws ec2 describe-security-group-rules --filter Name=group-id,Values=\\1) <(aws ec2 describe-security-groups --group-ids \\1)\n",
				},
			},
		},
		{
			Notes:     "switch terrafrom version",
			Precision: "high",
			Regex:     `required_version = "~> (\d*\.\d*\.\d*)"`,
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "change terraform version",
					Action:    4,
					Parameter: "tfswitch --latest-stable \\1\n",
				},
			},
		},
		{
			Notes:     "forget ssh host key",
			Precision: "high",
			Regex:     `Offending .* key in (.*):(\d*)`,
			Actions: []SmartSelectionRuleAction{
				{
					Title:     "remove host signature",
					Action:    4,
					Parameter: `sed -i "" '\2d' \1\n`,
				},
			},
		},
	}

	return append(ssr, loadUserSSR(custom)...)
}

func loadUserSSR(path string) []SmartSelectionRule {
	userSsr, err := homedir.Expand(path)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot expand path")
	}

	if _, err := os.Stat(userSsr); os.IsNotExist(err) {
		return []SmartSelectionRule{}
	}

	bytes, err := ioutil.ReadFile(userSsr)
	if err != nil {
		log.Fatal().Err(err).Str("userSsr", userSsr).Msg("cannot read file")
	}

	var userSSRs []SmartSelectionRule

	err = json.Unmarshal(bytes, &userSSRs)
	if err != nil {
		log.Fatal().Err(err).
			Str("userSsr", userSsr).
			Str("string(bytes)", string(bytes)).
			Msg("cannot parse json file")
	}

	return userSSRs
}
