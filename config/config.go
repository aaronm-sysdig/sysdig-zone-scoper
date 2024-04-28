package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
	"strconv"
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
	CreateZones         bool
	CreateTeams         bool
	LogLevel            string
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

func getOSEnvBool(logger *logrus.Logger, environmentVariable string, optional bool) bool {
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
}

func (c *Configuration) Build(logger *logrus.Logger) error {
	c.SecureApiToken = getOSEnvString(logger, "SECURE_API_TOKEN", false)
	c.SysdigApiEndpoint = getOSEnvString(logger, "SYSDIG_API_ENDPOINT", false)

	// Setup Label to group from
	var groupingLabel string
	var boolSilent bool
	var teamZoneMappingFile string
	var teamTemplateName string
	var boolCreateTeams bool
	var boolCreateZones bool
	var LogLevel string
	pflag.StringVarP(&groupingLabel, "grouping-label", "l", "", "Label to group by")
	pflag.StringVarP(&teamZoneMappingFile, "team-zone-mapping", "m", "", "CSV file to load for team to zone mapping")
	pflag.StringVarP(&teamTemplateName, "template-team", "e", "", "Template Team name")
	pflag.StringVarP(&LogLevel, "log-level", "d", "", "Logging Level. INFO, DEBUG or ERROR")

	pflag.BoolVarP(&boolSilent, "silent", "s", false, "Run Silently without dryrun prompt")
	pflag.BoolVarP(&boolCreateTeams, "create-teams", "t", false, "Create Teams")
	pflag.BoolVarP(&boolCreateZones, "create-zones", "z", false, "Create Zones")

	pflag.Parse()

	if groupingLabel == "" {
		logger.Info("'grouping-label' not  found on the command line.  Checking 'GROUPING-LABEL' environment variable instead")
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

	if !boolCreateTeams {
		logger.Info("'create-teams' not  found on the command line.  Checking 'CREATE_TEAMS' environment variable instead")
		c.CreateTeams = getOSEnvBool(logger, "CREATE_TEAMS", true)
	} else {
		c.CreateTeams = true
	}

	if !boolCreateZones {
		logger.Info("'create-zones' not  found on the command line.  Checking 'CREATE_ZONES' environment variable instead")
		c.CreateZones = getOSEnvBool(logger, "CREATE_ZONES", true)
	} else {
		c.CreateZones = true
	}

	if LogLevel == "" {
		logger.Info("'log-mode' not  found on the command line.  Checking 'LOG_MODE' environment variable instead")
		c.LogLevel = getOSEnvString(logger, "LOG_LEVEL", true)
	} else {
		c.LogLevel = "INFO"
	}

	// Some logic for what you select

	if c.CreateZones && c.CreateTeams {
		logger.Fatal("Sorry, you cannot run Create Zones and Create teams in one execution. Exiting...")
	}

	if !c.CreateZones && !c.CreateTeams {
		logger.Fatal("You have not specified either to Create teams or zones, so I will just exit.  Goodbye. Exiting...")
	}

	c.Silent = boolSilent

	//Get our static list of zones to keep even if we did not create or update them
	envStaticZones := getOSEnvString(logger, "STATIC_ZONES", true)
	staticZones := strings.Split(envStaticZones, ",")
	c.StaticZones = make(map[string]bool)
	for _, sliceZone := range staticZones {
		c.StaticZones[sliceZone] = true
	}

	return nil
}
