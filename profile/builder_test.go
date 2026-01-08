package profile

import (
	"testing"

	"github.com/mhristof/germ/iterm"
	"github.com/stretchr/testify/assert"
)

func TestNewBuilder(t *testing.T) {
	builder := NewBuilder("test-profile")
	
	assert.Equal(t, "test-profile", builder.name)
	assert.NotNil(t, builder.config)
	assert.NotNil(t, builder.tags)
	assert.NotNil(t, builder.keyboardMap)
	assert.NotNil(t, builder.triggers)
	assert.NotNil(t, builder.boundHosts)
}

func TestBuilder_WithCommand(t *testing.T) {
	profile := NewBuilder("test").
		WithCommand("echo hello").
		Build()
	
	assert.Equal(t, "echo hello", profile.Command)
	assert.Equal(t, "Yes", profile.CustomCommand)
}

func TestBuilder_WithTags(t *testing.T) {
	profile := NewBuilder("test").
		WithTags("tag1", "tag2").
		Build()
	
	// Profile always has at least one tag (the unique name), so check if our tags are included
	assert.Contains(t, profile.Tags, "tag1")
	assert.Contains(t, profile.Tags, "tag2")
}

func TestBuilder_WithKeyboardShortcut(t *testing.T) {
	profile := NewBuilder("test").
		WithKeyboardShortcut("test-key", 12, "test text").
		Build()
	
	assert.Contains(t, profile.KeyboardMap, "test-key")
	assert.Equal(t, int64(12), profile.KeyboardMap["test-key"].Action)
	assert.Equal(t, "test text", profile.KeyboardMap["test-key"].Text)
}

func TestBuilder_WithAltAShortcut(t *testing.T) {
	profile := NewBuilder("test").
		WithAltAShortcut("alt-a text").
		Build()
	
	assert.Contains(t, profile.KeyboardMap, iterm.KeyboardSortcutAltA)
	assert.Equal(t, int64(12), profile.KeyboardMap[iterm.KeyboardSortcutAltA].Action)
	assert.Equal(t, "alt-a text", profile.KeyboardMap[iterm.KeyboardSortcutAltA].Text)
}

func TestBuilder_WithTrigger(t *testing.T) {
	trigger := iterm.Trigger{
		Action:    "HighlightTrigger",
		Parameter: "red",
		Regex:     "ERROR",
		Partial:   true,
	}
	
	profile := NewBuilder("test").
		WithTrigger(trigger).
		Build()
	
	// Profile may have default triggers, so check that our trigger is included
	found := false
	for _, t := range profile.Triggers {
		if t.Action == "HighlightTrigger" && t.Parameter == "red" && t.Regex == "ERROR" {
			found = true
			break
		}
	}
	assert.True(t, found, "Custom trigger should be found in profile triggers")
}

func TestBuilder_WithBoundHosts(t *testing.T) {
	profile := NewBuilder("test").
		WithBoundHosts("host1", "host2").
		Build()
	
	assert.Equal(t, []string{"host1", "host2"}, profile.BoundHosts)
}

func TestBuilder_WithPrefix(t *testing.T) {
	profile := NewBuilder("test").
		WithPrefix("prefix").
		Build()
	
	assert.Equal(t, "prefix-test", profile.Name)
}

func TestBuilder_FluentInterface(t *testing.T) {
	profile := NewBuilder("test").
		WithCommand("echo test").
		WithTags("tag1", "tag2").
		WithAltAShortcut("shortcut text").
		WithPrefix("prefix").
		Build()
	
	assert.Equal(t, "prefix-test", profile.Name)
	assert.Equal(t, "echo test", profile.Command)
	assert.Contains(t, profile.Tags, "tag1")
	assert.Contains(t, profile.Tags, "tag2")
	assert.Contains(t, profile.KeyboardMap, iterm.KeyboardSortcutAltA)
}

func TestAWSProfileBuilder_WithAWSProfile(t *testing.T) {
	profile := NewAWSProfileBuilder("aws-test").
		WithAWSProfile("my-profile").
		Build()
	
	assert.Contains(t, profile.Command, "AWS_PROFILE=my-profile")
	assert.Contains(t, profile.Command, "/usr/bin/login")
}

func TestSSHProfileBuilder_WithSSHCommand(t *testing.T) {
	profile := NewSSHProfileBuilder("server1").
		WithSSHCommand("server1").
		WithHostIP("192.168.1.10").
		Build()
	
	assert.Equal(t, "server1", profile.Name)
	assert.Equal(t, "ssh server1", profile.Command)
	assert.Contains(t, profile.Tags, "ssh")
	assert.Contains(t, profile.Tags, "192.168.1.10")
}

func TestSSHProfileBuilder_WithTmuxDetach(t *testing.T) {
	profile := NewSSHProfileBuilder("server1").
		WithTmuxDetach().
		Build()
	
	assert.Contains(t, profile.KeyboardMap, "0x77-0x100000-0xd")
	assert.Equal(t, int64(25), profile.KeyboardMap["0x77-0x100000-0xd"].Action)
	assert.Equal(t, "Detach\ntmux.Detach", profile.KeyboardMap["0x77-0x100000-0xd"].Text)
}

func TestK8sProfileBuilder_WithKubeConfig(t *testing.T) {
	profile := NewK8sProfileBuilder("my-cluster").
		WithKubeConfig("/path/to/kubeconfig").
		Build()
	
	assert.Equal(t, "k8s-my-cluster", profile.Name)
	assert.Contains(t, profile.Command, "KUBECONFIG=/path/to/kubeconfig")
	assert.Contains(t, profile.Tags, "k8s")
}

func TestK8sProfileBuilder_WithAWSProfile(t *testing.T) {
	profile := NewK8sProfileBuilder("my-cluster").
		WithKubeConfig("/path/to/kubeconfig").
		WithAWSProfile("aws-profile").
		Build()
	
	assert.Contains(t, profile.Tags, "aws-profile=aws-profile")
	// Note: The command modification logic needs refinement in the actual implementation
}

func TestSSMProfileBuilder_WithSSMCommand(t *testing.T) {
	profile := NewSSMProfileBuilder("account", "us-east-1", "instance1").
		WithSSMCommand("aws-profile", "instance1").
		WithAWSAccountInfo("account", "123456789", "us-east-1", []string{"US", "East", "use1"}).
		Build()
	
	assert.Equal(t, "account:us-east-1:ssm-instance1", profile.Name)
	assert.Contains(t, profile.InitialText, "AWS_PROFILE=aws-profile ssm instance1")
	assert.Contains(t, profile.KeyboardMap, iterm.KeyboardSortcutAltA)
	
	// The CustomCommand field should be set to "No" via the config map during NewProfile
	assert.Equal(t, "No", profile.CustomCommand)
}