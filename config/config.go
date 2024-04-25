package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
)

type Configuration struct {
	MyPAT             string
	GitRepo           string
	ConfigFile        string
	SecureApiToken    string
	SysdigApiEndpoint string
	GroupingLabel     string
}

func getOSEnvString(logger *logrus.Logger, environmentVariable string, optional bool) string {
	env := os.Getenv(environmentVariable)
	if env == "" {
		if !optional {
			logger.Fatalf("Fatal Error: Could not find %s environment variable, exiting status code (1)", environmentVariable)
		} else {
			logger.Printf("Warning: Could not find %s environment variable, continuing anyway...", environmentVariable)
		}
	} else {
		logger.Printf("Found %s Variable, continuing ...", environmentVariable)
	}
	return env
}

func (c *Configuration) Build(logger *logrus.Logger) error {
	//c.MyPAT = base64.StdEncoding.EncodeToString([]byte(":" + getOSEnvString(logger, "MY_PAT", false)))
	//c.GitRepo = getOSEnvString(logger, "GIT_REPO", true)
	//c.ConfigFile = getOSEnvString(logger, "CONFIG_FILE", true)
	c.SecureApiToken = getOSEnvString(logger, "SECURE_API_TOKEN", false)
	c.SysdigApiEndpoint = getOSEnvString(logger, "SYSDIG_API_ENDPOINT", false)
	// Setup Label to group from
	var groupingLabel string
	pflag.StringVarP(&groupingLabel, "grouping-label", "l", os.Getenv("GROUPING-LABEL"), "Label to group by")

	pflag.Parse()
	c.GroupingLabel = getOSEnvString(logger, "GROUPING_LABEL", false)

	fmt.Println("")
	return nil
}
