package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
	"strings"
)

type Configuration struct {
	MyPAT               string
	GitRepo             string
	ConfigFile          string
	SecureApiToken      string
	SysdigApiEndpoint   string
	GroupingLabel       string
	Silent              bool
	StaticZones         map[string]bool
	TeamZoneMappingFile string
	TeamTemplateName    string
	LogLevel            string
	Mode                string
	TeamPrefix          string
	DryRun              bool
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

/*func getOSEnvBool(logger *logrus.Logger, environmentVariable string, optional bool) bool {
	env := os.Getenv(environmentVariable)
	if env == "" {
		if !optional {
			logger.Fatalf("Fatal Error: Could not find %s environment variable, exiting status code (1)", environmentVariable)
		} else {
			logger.Printf("Warning: Could not find %s environment variable, continuing anyway...", environmentVariable)
		}
		return false // Default to false or another sensible default for your use case
	}

	boolVal, err := strconv.ParseBool(env)
	if err != nil {
		logger.Errorf("Error parsing %s environment variable: %v", environmentVariable, err)
		return false // Default to false or handle the error according to your requirements
	}

	logger.Printf("Found %s Variable with value %t, continuing ...", environmentVariable, boolVal)
	return boolVal
}*/

func (c *Configuration) Build(logger *logrus.Logger) error {
	c.SecureApiToken = getOSEnvString(logger, "SECURE_API_TOKEN", false)
	c.SysdigApiEndpoint = getOSEnvString(logger, "SYSDIG_API_ENDPOINT", false)

	// Setup Label to group from
	var groupingLabel string
	var boolSilent bool
	var boolDryRun bool
	var teamZoneMappingFile string
	var teamTemplateName string
	var LogLevel string
	var mode string
	var teamPrefix string

	pflag.StringVarP(&groupingLabel, "grouping-label", "l", "", "Label to group by")
	pflag.StringVarP(&teamZoneMappingFile, "team-zone-mapping", "m", "", "CSV file to load for team to zone mapping")
	pflag.StringVarP(&teamTemplateName, "template-team", "e", "", "Template Team name")
	pflag.StringVarP(&LogLevel, "log-level", "d", "", "Logging Level. INFO, DEBUG or ERROR")
	pflag.StringVarP(&mode, "mode", "o", "", "Operation mode.  ZONE or TEAM")
	pflag.StringVarP(&teamPrefix, "team-prefix", "t", "", "Team Name Prefix")

	pflag.BoolVarP(&boolSilent, "silent", "s", false, "Run Silently without dryrun prompt")
	pflag.BoolVarP(&boolDryRun, "dryrun", "r", false, "DryRun mode.  Will not actually do anything irrespective of even --silent/-s ")

	pflag.Parse()

	if groupingLabel == "" {
		logger.Info("'grouping-label' not  found on the command line.  Checking 'GROUPING_LABEL' environment variable instead")
		c.GroupingLabel = getOSEnvString(logger, "GROUPING_LABEL", false)
	} else {
		c.GroupingLabel = groupingLabel
	}

	if teamZoneMappingFile == "" {
		logger.Info("'team-zone-mapping' not  found on the command line.  Checking 'TEAM_ZONE_MAPPING' environment variable instead")
		c.TeamZoneMappingFile = getOSEnvString(logger, "TEAM_ZONE_MAPPING", true)
	} else {
		c.TeamZoneMappingFile = teamZoneMappingFile
	}

	if teamTemplateName == "" {
		logger.Info("'team-template-name' not  found on the command line.  Checking 'TEAM_TEMPLATE_NAME' environment variable instead")
		c.TeamTemplateName = getOSEnvString(logger, "TEAM_TEMPLATE_NAME", true)
	} else {
		c.TeamTemplateName = teamTemplateName
	}

	if mode == "" {
		logger.Info("'mode' not found on the command line.  Checking 'MODE' environment variable instead")
		c.Mode = getOSEnvString(logger, "MODE", false)
	} else {
		c.Mode = mode
	}

	if LogLevel == "" {
		logger.Info("'log-mode' not  found on the command line.  Checking 'LOG_MODE' environment variable instead")
		c.LogLevel = getOSEnvString(logger, "LOG_LEVEL", true)
	} else {
		c.LogLevel = "INFO"
	}

	if teamPrefix == "" {
		logger.Info("'team-prefix' not  found on the command line.  Checking 'TEAM_PREFIX' environment variable instead")
		c.TeamPrefix = getOSEnvString(logger, "TEAM_PREFIX", true)
	} else {
		c.TeamPrefix = teamPrefix
	}

	c.Silent = boolSilent
	c.DryRun = boolDryRun
	if c.DryRun {
		logger.Infof("Dryrun mode enabled")
	}

	//Get our static list of zones to keep even if we did not create or update them
	envStaticZones := getOSEnvString(logger, "STATIC_ZONES", true)
	c.StaticZones = make(map[string]bool)
	if len(envStaticZones) > 0 {
		staticZones := strings.Split(envStaticZones, ",")
		for _, sliceZone := range staticZones {
			c.StaticZones[strings.TrimSpace(sliceZone)] = true
		}
	}
	// Add in 'Entire Infrastrucutre' and 'Entire Git' as they are immutable
	c.StaticZones["Entire Infrastructure"] = true
	c.StaticZones["Entire Git"] = true

	for sliceZone := range c.StaticZones {
		logger.Debugf("Static Zone '%s'", sliceZone)
	}
	return nil
}
