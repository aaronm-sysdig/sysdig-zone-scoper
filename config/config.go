package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
	"strings"
)

type Configuration struct {
	MyPAT             string
	GitRepo           string
	ConfigFile        string
	SecureApiToken    string
	SysdigApiEndpoint string
	GroupingLabel     string
	Silent            bool
	StaticZones       map[string]bool
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
	c.SecureApiToken = getOSEnvString(logger, "SECURE_API_TOKEN", false)
	c.SysdigApiEndpoint = getOSEnvString(logger, "SYSDIG_API_ENDPOINT", false)

	// Setup Label to group from
	var groupingLabel string
	var silent bool
	pflag.StringVarP(&groupingLabel, "grouping-label", "l", os.Getenv("GROUPING-LABEL"), "Label to group by")
	pflag.BoolVarP(&silent, "silent", "s", false, "Run Silently without dryrun prompt")
	pflag.Parse()
	c.Silent = true
	c.GroupingLabel = getOSEnvString(logger, "GROUPING_LABEL", false)

	//Get our static list of zones to keep even if we did not create or update them
	envStaticZones := getOSEnvString(logger, "STATIC_ZONES", true)
	staticZones := strings.Split(envStaticZones, ",")
	c.StaticZones = make(map[string]bool)
	for _, sliceZone := range staticZones {
		c.StaticZones[sliceZone] = true
	}
	fmt.Println("")
	return nil
}
