package maws

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/kayac/go-config"
	"github.com/pkg/errors"
)

var DefaultAllowedCommandPrefixes = []string{
	"list-",
	"get-",
	"describe-",
}

type Config struct {
	Roles                  []string `yaml:"roles"`
	AllowedCommandPrefixes []string `yaml:"allowed_command_prefixes"`
}

func LoadConfig(filename string) (*Config, error) {
	var c Config
	err := config.LoadWithEnv(&c, filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config file")
	}

	// validate roles
	for _, role := range c.Roles {
		an, err := arn.Parse(role)
		if err != nil {
			return nil, errors.Errorf("invalid ARN %s", role)
		}
		if an.Service != "iam" || !strings.HasPrefix(an.Resource, "role/") {
			return nil, errors.Errorf("not a IAM role ARN %s", role)
		}
	}

	if len(c.AllowedCommandPrefixes) == 0 {
		c.AllowedCommandPrefixes = DefaultAllowedCommandPrefixes
	}

	return &c, err
}

func (c *Config) restrictCommand(commands []string) error {
	if len(commands) < 2 {
		return errors.New("insufficient commands")
	}

	// example: maws -- sts get-caller-identity
	// commands[0]: sts
	// commands[1]: get-caller-identity
	for _, prefix := range c.AllowedCommandPrefixes {
		if strings.HasPrefix(commands[1], prefix) {
			return nil
		}
	}
	return errors.Errorf(
		"%s %s is restricted. allowed command prefixes are %v",
		commands[0],
		commands[1],
		c.AllowedCommandPrefixes,
	)
}
