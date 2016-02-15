// Package sdk provides the API clients for the AWS SDK.
package sdk

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/pflag"
)

var (
	AutoScalingConfigs    = []*aws.Config{}
	EC2Configs            = []*aws.Config{}
	CloudFormationConfigs = []*aws.Config{}
	ELBConfigs            = []*aws.Config{}
)

// SessionProvider represents the current AWS session.
type SessionProvider interface {
	GetSession() *session.Session
}

var (
	// DefaultSessionProvider represents a default SessionProvider.
	DefaultSessionProvider *ConfigurableSessionProvider = &ConfigurableSessionProvider{
		ConfigRegion:                "eu-west-1",
		ConfigDisableENVCredentials: false,
		ConfigAccessKey:             "",
		ConfigSecretKey:             "",
		ConfigSessionToken:          "",
	}
)

// ConfigurableSessionProvider represents a SessionProvider that can be configured.
type ConfigurableSessionProvider struct {
	_session     *session.Session
	sessionMutex sync.Mutex

	ConfigDisableENVCredentials bool
	ConfigRegion                string
	ConfigAccessKey             string
	ConfigSecretKey             string
	ConfigSessionToken          string
}

// RegisterFlagSet registers command line flags with the SessionProvider.
func (dsp *ConfigurableSessionProvider) RegisterFlagSet(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&dsp.ConfigRegion, "aws-region", dsp.ConfigRegion, "AWS region")

	flagSet.StringVar(&dsp.ConfigAccessKey, "aws-access-key", dsp.ConfigAccessKey, "AWS access key")
	flagSet.StringVar(&dsp.ConfigSecretKey, "aws-secret-key", dsp.ConfigSecretKey, "AWS secret key")
	flagSet.StringVar(&dsp.ConfigSessionToken, "aws-session-token", dsp.ConfigSessionToken, "AWS session token - empty most of the time")

	flagSet.BoolVar(&dsp.ConfigDisableENVCredentials, "disable-aws-env-credentials", dsp.ConfigDisableENVCredentials, "do not read credentials from environment variables")
}

// GetSessions returns the current session from the SessionProvider.
func (dsp *ConfigurableSessionProvider) GetSession() *session.Session {
	dsp.sessionMutex.Lock()
	defer dsp.sessionMutex.Unlock()

	if dsp._session != nil {
		return dsp._session
	}

	cfgs := []*aws.Config{
		aws.NewConfig().WithRegion(dsp.ConfigRegion),
	}

	if dsp.ConfigAccessKey != "" {
		cfgs = append(cfgs, aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(
			dsp.ConfigAccessKey,
			dsp.ConfigSecretKey,
			dsp.ConfigSessionToken,
		)))
	} else if !dsp.ConfigDisableENVCredentials {
		// NOTE: We may want to init all credential providers in one config, as they might overwrite each other
		cfgs = append(cfgs, aws.NewConfig().WithCredentials(credentials.NewCredentials(&credentials.EnvProvider{})))
	} else {
		panic("no valid configuration parameters for aws credentials found.")
	}

	dsp._session = session.New(cfgs...)
	return dsp._session
}
