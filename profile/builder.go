package profile

import (
	"fmt"
	"os/user"
	"strings"

	"github.com/mhristof/germ/iterm"
	"github.com/rs/zerolog/log"
)

// Builder provides a fluent interface for creating iTerm profiles
type Builder struct {
	name        string
	config      map[string]string
	tags        []string
	keyboardMap map[string]iterm.KeyboardMap
	triggers    []iterm.Trigger
	boundHosts  []string
}

// NewBuilder creates a new profile builder with the given name
func NewBuilder(name string) *Builder {
	return &Builder{
		name:        name,
		config:      make(map[string]string),
		tags:        make([]string, 0),
		keyboardMap: make(map[string]iterm.KeyboardMap),
		triggers:    make([]iterm.Trigger, 0),
		boundHosts:  make([]string, 0),
	}
}

// WithCommand sets the command for the profile
func (b *Builder) WithCommand(command string) *Builder {
	b.config["Command"] = command
	b.config["Custom Command"] = "Yes"
	return b
}

// WithInitialText sets the initial text for the profile
func (b *Builder) WithInitialText(text string) *Builder {
	b.config["Initial Text"] = text
	return b
}

// WithTags adds tags to the profile
func (b *Builder) WithTags(tags ...string) *Builder {
	b.tags = append(b.tags, tags...)
	return b
}

// WithTagsString sets tags from a comma-separated string
func (b *Builder) WithTagsString(tagsStr string) *Builder {
	b.config["Tags"] = tagsStr
	return b
}

// WithConfig sets arbitrary configuration values
func (b *Builder) WithConfig(key, value string) *Builder {
	b.config[key] = value
	return b
}

// WithKeyboardShortcut adds a keyboard shortcut
func (b *Builder) WithKeyboardShortcut(key string, action int64, text string) *Builder {
	b.keyboardMap[key] = iterm.KeyboardMap{
		Action: action,
		Text:   text,
	}
	return b
}

// WithAltAShortcut adds the common Alt+A keyboard shortcut
func (b *Builder) WithAltAShortcut(text string) *Builder {
	return b.WithKeyboardShortcut(iterm.KeyboardSortcutAltA, iterm.KeyboardSendText, text)
}

// WithTrigger adds a trigger to the profile
func (b *Builder) WithTrigger(trigger iterm.Trigger) *Builder {
	b.triggers = append(b.triggers, trigger)
	return b
}

// WithBoundHosts sets bound hosts for the profile
func (b *Builder) WithBoundHosts(hosts ...string) *Builder {
	b.boundHosts = append(b.boundHosts, hosts...)
	return b
}

// WithPrefix adds a prefix to the profile name
func (b *Builder) WithPrefix(prefix string) *Builder {
	if prefix != "" {
		b.name = fmt.Sprintf("%s-%s", prefix, b.name)
	}
	return b
}

// Build creates the final iTerm profile
func (b *Builder) Build() *iterm.Profile {
	profile := iterm.NewProfile(b.name, b.config)
	
	// Add additional tags if any were specified
	if len(b.tags) > 0 {
		profile.Tags = append(profile.Tags, b.tags...)
	}
	
	// Add keyboard shortcuts
	if len(b.keyboardMap) > 0 {
		if profile.KeyboardMap == nil {
			profile.KeyboardMap = make(map[string]iterm.KeyboardMap)
		}
		for key, mapping := range b.keyboardMap {
			profile.KeyboardMap[key] = mapping
		}
	}
	
	// Add triggers
	if len(b.triggers) > 0 {
		profile.Triggers = append(profile.Triggers, b.triggers...)
	}
	
	// Add bound hosts
	if len(b.boundHosts) > 0 {
		profile.BoundHosts = b.boundHosts
	}
	
	return profile
}

// AWSProfileBuilder provides AWS-specific profile building functionality
type AWSProfileBuilder struct {
	*Builder
}

// NewAWSProfileBuilder creates a builder for AWS profiles
func NewAWSProfileBuilder(name string) *AWSProfileBuilder {
	return &AWSProfileBuilder{
		Builder: NewBuilder(name),
	}
}

// WithAWSProfile sets up the profile with AWS_PROFILE environment variable
func (b *AWSProfileBuilder) WithAWSProfile(awsProfile string) *AWSProfileBuilder {
	user, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot find current user")
	}
	
	command := fmt.Sprintf("/usr/bin/env AWS_PROFILE=%s /usr/bin/login -fp %s", awsProfile, user.Username)
	b.WithCommand(command)
	return b
}

// WithAWSLoginCommand sets up a login command for AWS SSO
func (b *AWSProfileBuilder) WithAWSLoginCommand(awsProfile, loginCmd string) *AWSProfileBuilder {
	b.WithCommand(loginCmd)
	b.WithAltAShortcut(fmt.Sprintf("AWS_PROFILE=%s aws sso login\n", awsProfile))
	return b
}

// SSHProfileBuilder provides SSH-specific profile building functionality
type SSHProfileBuilder struct {
	*Builder
}

// NewSSHProfileBuilder creates a builder for SSH profiles
func NewSSHProfileBuilder(host string) *SSHProfileBuilder {
	return &SSHProfileBuilder{
		Builder: NewBuilder(host),
	}
}

// WithSSHCommand sets up the SSH command
func (b *SSHProfileBuilder) WithSSHCommand(host string) *SSHProfileBuilder {
	b.WithCommand(fmt.Sprintf("ssh %s", host))
	b.WithTags("ssh")
	return b
}

// WithHostIP adds the host IP as a tag
func (b *SSHProfileBuilder) WithHostIP(ip string) *SSHProfileBuilder {
	if ip != "" {
		b.WithTags(ip)
	}
	return b
}

// WithTmuxDetach adds tmux detach keyboard shortcut
func (b *SSHProfileBuilder) WithTmuxDetach() *SSHProfileBuilder {
	b.WithKeyboardShortcut("0x77-0x100000-0xd", 25, "Detach\ntmux.Detach")
	return b
}

// K8sProfileBuilder provides Kubernetes-specific profile building functionality
type K8sProfileBuilder struct {
	*Builder
}

// NewK8sProfileBuilder creates a builder for Kubernetes profiles
func NewK8sProfileBuilder(clusterName string) *K8sProfileBuilder {
	return &K8sProfileBuilder{
		Builder: NewBuilder(fmt.Sprintf("k8s-%s", clusterName)),
	}
}

// WithKubeConfig sets up the profile with KUBECONFIG environment variable
func (b *K8sProfileBuilder) WithKubeConfig(kubeconfigPath string) *K8sProfileBuilder {
	user, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot find current user")
	}
	
	command := fmt.Sprintf("/usr/bin/env KUBECONFIG=%s /usr/bin/login -fp %s", kubeconfigPath, user.Username)
	b.WithCommand(command)
	b.WithTags("k8s")
	return b
}

// WithAWSProfile adds AWS profile to the Kubernetes profile
func (b *K8sProfileBuilder) WithAWSProfile(awsProfile string) *K8sProfileBuilder {
	if awsProfile != "" {
		user, err := user.Current()
		if err != nil {
			log.Fatal().Err(err).Msg("cannot find current user")
		}
		
		// Get the current kubeconfig path from existing command or use a default
		kubeconfigPath := b.extractKubeConfigFromCommand()
		
		// Build new command with both KUBECONFIG and AWS_PROFILE
		command := fmt.Sprintf("/usr/bin/env KUBECONFIG=%s AWS_PROFILE=%s /usr/bin/login -fp %s", 
			kubeconfigPath, awsProfile, user.Username)
		
		b.WithCommand(command)
		b.WithTags(fmt.Sprintf("aws-profile=%s", awsProfile))
	}
	return b
}

// extractKubeConfigFromCommand extracts the KUBECONFIG path from the existing command
func (b *K8sProfileBuilder) extractKubeConfigFromCommand() string {
	if command, exists := b.config["Command"]; exists {
		// Look for KUBECONFIG= in the command
		if idx := strings.Index(command, "KUBECONFIG="); idx != -1 {
			start := idx + len("KUBECONFIG=")
			end := strings.Index(command[start:], " ")
			if end == -1 {
				end = len(command)
			} else {
				end += start
			}
			return command[start:end]
		}
	}
	return ""
}

// SSMProfileBuilder provides SSM-specific profile building functionality
type SSMProfileBuilder struct {
	*Builder
}

// NewSSMProfileBuilder creates a builder for SSM profiles
func NewSSMProfileBuilder(accountAlias, region, instanceName string) *SSMProfileBuilder {
	name := fmt.Sprintf("%s:%s:ssm-%s", accountAlias, region, instanceName)
	return &SSMProfileBuilder{
		Builder: NewBuilder(name),
	}
}

// WithSSMCommand sets up the SSM command
func (b *SSMProfileBuilder) WithSSMCommand(awsProfile, instanceName string) *SSMProfileBuilder {
	bashCommand := fmt.Sprintf("bash -c 'AWS_PROFILE=%s ssm %s'", awsProfile, instanceName)
	b.WithInitialText(bashCommand)
	
	// Add Alt+A shortcut for SSO login
	loginText := fmt.Sprintf("AWS_PROFILE=%s aws sso login && %s\n", awsProfile, bashCommand)
	b.WithAltAShortcut(loginText)
	
	return b
}

// Build creates the final iTerm profile and handles SSM-specific settings
func (b *SSMProfileBuilder) Build() *iterm.Profile {
	profile := b.Builder.Build()
	
	// SSM profiles should not have CustomCommand set to "Yes"
	// They use InitialText instead
	profile.CustomCommand = "No"
	
	return profile
}

// WithAWSAccountInfo adds AWS account and region information as tags
func (b *SSMProfileBuilder) WithAWSAccountInfo(accountAlias, accountID, region string, regionTags []string) *SSMProfileBuilder {
	tags := fmt.Sprintf("AWS, %s,account=%s", accountAlias, accountID)
	if len(regionTags) > 2 {
		tags += ",region_id=" + regionTags[2]
	}
	b.WithTagsString(tags)
	return b
}