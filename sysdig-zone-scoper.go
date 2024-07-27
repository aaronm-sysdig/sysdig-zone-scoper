package main

import (
	"encoding/csv"
	"fmt"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/config"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/dataManipulation"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/mdsNamespaces"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/sysdighttp"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/teamPayload"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/teamZoneMapping"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/zonePayload"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
)

type customFormatter struct {
	logrus.TextFormatter
}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Ensure log level is displayed with a fixed width of 5 characters
	levelText := fmt.Sprintf("%-6s", strings.ToUpper(entry.Level.String()))

	// Simplify and format the function name and line number directly after the function name
	functionName := runtime.FuncForPC(entry.Caller.PC).Name()
	lastSlash := strings.LastIndex(functionName, "/")
	shortFunctionName := functionName

	if lastSlash != -1 {
		// Extract the part after the last slash
		afterSlash := functionName[lastSlash+1:]

		// Check if the extracted part contains parentheses (indicating a method)
		firstParen := strings.Index(afterSlash, "(")
		if firstParen != -1 {
			// Extract the struct and method name
			structAndMethodName := afterSlash[:firstParen]
			parts := strings.Split(structAndMethodName, ".")
			if len(parts) >= 2 {
				// Combine the first part (struct) and the last part (method)
				shortFunctionName = parts[0] + "." + parts[len(parts)-1]
			} else {
				shortFunctionName = structAndMethodName
			}
		} else {
			shortFunctionName = afterSlash
		}
		shortFunctionName = afterSlash
	}

	formattedCaller := fmt.Sprintf("%s:%d", shortFunctionName, entry.Caller.Line) // Combine function name and line

	// Right-align the caller info to ensure that it occupies a fixed width
	rightAlignedCaller := fmt.Sprintf("%-40s", formattedCaller)

	// Create the formatted log entry
	logMessage := fmt.Sprintf("%s[%s] %s %s\n", levelText, entry.Time.Format("2006-01-02 15:04:05"), rightAlignedCaller, entry.Message)

	return []byte(logMessage), nil
}

func getMDSNamespaces(appConfig *config.Configuration, logger *logrus.Logger, mdsNs *mdsNamespaces.NamespacePayload) (err error) {
	configMdsNamespaces := sysdighttp.DefaultSysdigRequestConfig(appConfig.SysdigApiEndpoint, appConfig.SecureApiToken)
	return mdsNs.GetNamespaces(logger, &configMdsNamespaces)
}

func getZones(appConfig *config.Configuration, logger *logrus.Logger, zones *zonePayload.ZonePayload) (err error) {
	configZones := sysdighttp.DefaultSysdigRequestConfig(appConfig.SysdigApiEndpoint, appConfig.SecureApiToken)
	return zones.GetZones(logger, &configZones)
}

func createZone(appConfig *config.Configuration,
	logger *logrus.Logger,
	zones *zonePayload.ZonePayload,
	productName string) (createdZone *zonePayload.Zone, err error) {

	var newZone = &zonePayload.CreateZone{
		Name:        productName,
		Description: fmt.Sprintf("Zone for '%s'", productName),
		Scopes: []zonePayload.Scope{{
			Rules:      "",
			TargetType: "kubernetes",
		}},
	}

	configCreateZone := sysdighttp.DefaultSysdigRequestConfig(appConfig.SysdigApiEndpoint, appConfig.SecureApiToken)
	logger.Infof("Creating zone %s'", productName)
	createdZone, err = zones.CreateNewZone(logger, &configCreateZone, newZone)

	return
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func updateZone(appConfig *config.Configuration,
	logger *logrus.Logger,
	zones *zonePayload.ZonePayload,
	distinctProductNames map[string][]mdsNamespaces.ClusterNamespace,
	productName string,
	createdZone *zonePayload.Zone) (err error) {

	var clusters []string
	var namespaces []string
	//Generate the comma lists of clusters and namespaces
	for _, cn := range distinctProductNames[productName] {
		logger.Debugf("Cluster: %s, Namespace: %s", cn.Cluster, cn.Namespace)
		// Append cluster if not already in clusters
		if !contains(clusters, cn.Cluster) {
			clusters = append(clusters, cn.Cluster)
		}

		// Append namespace if not already in namespaces
		if !contains(namespaces, cn.Namespace) {
			namespaces = append(namespaces, cn.Namespace)
		}
	}

	var joinedClusters = fmt.Sprintf("\"%s\"", strings.Join(clusters, "\",\""))
	var joinedNamespaces = fmt.Sprintf("\"%s\"", strings.Join(namespaces, "\",\""))
	logger.Debugf("Joined cluster string: '%s'", joinedClusters)
	logger.Debugf("Joined namespace string: '%s'", joinedNamespaces)

	// Create a new scope without kubernetes
	var newScope []zonePayload.Scope
	for _, scpe := range createdZone.Scopes {
		if scpe.TargetType != "kubernetes" {
			newScope = append(newScope, scpe)
		}
	}
	// Add in our kubernetes scopes
	newScope = append(newScope, zonePayload.Scope{
		Rules:      fmt.Sprintf("clusterId in (%s) and namespace in (%s)", joinedClusters, joinedNamespaces),
		TargetType: "kubernetes"},
	)

	//Update Zone
	var updateZone = &zonePayload.UpdateZone{
		ID:     createdZone.ID,
		Name:   createdZone.Name,
		Scopes: newScope,
	}
	configUpdate := sysdighttp.DefaultSysdigRequestConfig(appConfig.SysdigApiEndpoint, appConfig.SecureApiToken)
	configUpdate.JSON = updateZone
	logger.Debugf("Updating zone '%s', zoneID %d", productName, createdZone.ID)
	logger.Debugf("Cluster: '%s', Namespace: '%s'", updateZone.Scopes[len(updateZone.Scopes)-2].Rules, updateZone.Scopes[len(updateZone.Scopes)-1].Rules)
	if err = zones.UpdateZone(logger, &configUpdate, updateZone); err != nil {
		logger.Errorf("Could not update zoneId '%d' for '%s'", createdZone.ID, productName)
	}
	// Retrieve, modify, and set the Zone to mark it as kept
	zone := zones.Zones[productName]
	zone.Keep = true
	zones.Zones[productName] = zone

	logger.Debugf("Zone '%s' updated and marked as kept", productName)

	return
}

func createClusterNSString(distinctProductNames map[string][]mdsNamespaces.ClusterNamespace, productName string) (joinedClusters string, joinedNamespaces string) {
	var clusters []string
	var namespaces []string
	//Generate the comma lists of clusters and namespaces
	for _, cn := range distinctProductNames[productName] {
		// Append cluster if not already in clusters
		if !contains(clusters, cn.Cluster) {
			clusters = append(clusters, cn.Cluster)
		}

		// Append namespace if not already in namespaces
		if !contains(namespaces, cn.Namespace) {
			namespaces = append(namespaces, cn.Namespace)
		}
	}
	return fmt.Sprintf(strings.Join(clusters, ",")), fmt.Sprintf(strings.Join(namespaces, ","))
}

func processDryRun() {
	// Inform the user that the file has been written
	fmt.Println("\"dry-run.csv\" has been written. Do you wish to continue? [Y/N]")

	// Function to read user input
	var response string
	fmt.Scanln(&response)

	if strings.TrimSpace(strings.ToUpper(response)) == "Y" {
		fmt.Println("Are you SURE? [Y/N]")
		fmt.Scanln(&response)
		if strings.TrimSpace(strings.ToUpper(response)) == "Y" {
			// User confirmed twice, continue the program
			fmt.Println("Continuing...")
		} else {
			fmt.Println("Exiting...")
			os.Exit(0)
		}
	} else {
		os.Exit(0)
	}
}

func getTeamZoneMapping(logger *logrus.Logger, appConfig *config.Configuration) (error, *teamZoneMapping.TeamZones) {
	teamZoneMappingFile, err := os.Open(appConfig.TeamZoneMappingFile)
	if err != nil {
		logger.Errorf("Error opening team zone mapping file: %v", err)
		return err, nil // Return nil to indicate an error occurred
	}
	defer func(teamZoneMappingFile *os.File) {
		_ = teamZoneMappingFile.Close()
	}(teamZoneMappingFile)

	teamZones := teamZoneMapping.NewTeamZones() // Initialize the TeamZones structure
	if err := teamZones.ParseCSV(teamZoneMappingFile); err != nil {
		logger.Errorf("Error parsing team zone mapping CSV: %v", err)
		return err, nil // Return nil to indicate an error occurred
	}

	// Print the map to verify the contents
	for team, zones := range *teamZones {
		formattedZones := fmt.Sprintf("[\"%s\"]", strings.Join(zones, "\", \""))
		logger.Infof("Team: %s, Zones: %v", team, formattedZones)
	}
	return nil, teamZones
}

func createOrUpdateTeam(logger *logrus.Logger,
	appConfig *config.Configuration,
	teamName string,
	zoneIds []int64,
	teamMapping *teamPayload.TeamPayload) (err error) {
	tz := &teamPayload.TeamPayload{}
	configCreateTeam := sysdighttp.DefaultSysdigRequestConfig(appConfig.SysdigApiEndpoint, appConfig.SecureApiToken)

	// Check if the team already exists, if so we will update (PUT) the team, else we will create (POST) it
	tb := &teamPayload.TeamBase{}
	configGetTeamByName := sysdighttp.DefaultSysdigRequestConfig(appConfig.SysdigApiEndpoint, appConfig.SecureApiToken)
	if err = tb.GetTeamByName(logger, &configGetTeamByName, teamName); err != nil {
		logger.Errorf("Could not execute query to ascertain if team '%s' exists. Error %v", teamName, err)
		return err
	}

	if len(tb.Data) > 0 {
		// Means we found the team and we need to run an update not a create
		if err = tz.UpdateTeamZoneMapping(logger, teamName, zoneIds, &configCreateTeam, &tb.Data[0]); err != nil {
			return err
		}
	} else {
		if err = tz.CreateTeamZoneMapping(logger, appConfig, teamName, zoneIds, &configCreateTeam, teamMapping); err != nil {
			return err
		}
	}

	return nil
}

func cRUDTeamMonitor(logger *logrus.Logger,
	appConfig *config.Configuration,
	teamName string,
	keyName string,
	teamMapping *teamPayload.TeamPayload) (err error) {
	tz := &teamPayload.TeamPayload{}
	configCreateTeam := sysdighttp.DefaultSysdigRequestConfig(appConfig.SysdigApiEndpoint, appConfig.SecureApiToken)

	// Check if the team already exists, if so we will update (PUT) the team, else we will create (POST) it
	tb := &teamPayload.TeamBase{}
	configGetTeamByName := sysdighttp.DefaultSysdigRequestConfig(appConfig.SysdigApiEndpoint, appConfig.SecureApiToken)
	if err = tb.GetTeamByName(logger, &configGetTeamByName, teamName); err != nil {
		logger.Errorf("Could not execute query to ascertain if team '%s' exists. Error %v", teamName, err)
		return err
	}

	if len(tb.Data) == 0 {
		teamMapping.Scopes = append(tz.Scopes, teamPayload.Scope{
			Expression: "container",
			Type:       "HOST_CONTAINER",
		}, teamPayload.Scope{
			Expression: fmt.Sprintf("%s = \"%s\"", appConfig.GroupingLabel, keyName),
			Type:       "AGENT",
		})
		logger.Infof("Creating team: %s", teamName)
		if !appConfig.DryRun {
			if err = tz.CreateTeamZoneMapping(logger, appConfig, teamName, nil, &configCreateTeam, teamMapping); err != nil {
				return err
			}
		}
	} else {
		logger.Infof("Skipping existing team: %s", teamName)
	}

	return nil
}

func setLogLevel(logger *logrus.Logger, appConfig *config.Configuration) {
	if strings.ToUpper(appConfig.LogLevel) == "INFO" {
		logger.SetLevel(logrus.InfoLevel)
	} else if strings.ToUpper(appConfig.LogLevel) == "DEBUG" {
		logger.SetLevel(logrus.DebugLevel)
	} else if strings.ToUpper(appConfig.LogLevel) == "ERROR" {
		logger.SetLevel(logrus.ErrorLevel)
	}
	logger.Infof("Setting LogLevel = '%v'", strings.ToUpper(logger.Level.String()))
}

func main() {
	var err error
	logger := logrus.New()
	logger.SetFormatter(&customFormatter{logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	}})
	logger.SetReportCaller(true) // Enables reporting of file, function, and line number
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)

	logger.Info("Sysdig Zone Scoper v9.5.9\n")

	appConfig := &config.Configuration{}
	if err := appConfig.Build(logger); err != nil {
		logger.Fatalf("Could not build configuration. Error %s", err)
	}

	// Set logging level based off configuration
	setLogLevel(logger, appConfig)

	// We need zones for both the teams and zones operations so run this either way
	fmt.Println("")
	zones := zonePayload.NewZonePayload()
	logger.Info("Getting list of Zones")
	if err = getZones(appConfig, logger, zones); err != nil {
		logger.Fatalf("Failed to retrieve zones. Error %v", err)
	}

	if strings.Contains(strings.ToUpper(appConfig.Mode), "ZONE") {
		fmt.Println("")
		logger.Info("------------------------------")
		logger.Info("Running in 'Create Zones' mode")
		logger.Info("------------------------------")

		logger.Info("Running in 'Create Zone' mode")
		// Build distinct mapping list for cluster and namespaces
		mdsNs := &mdsNamespaces.NamespacePayload{}
		logger.Infof("Getting mds Namespace list")
		if err = getMDSNamespaces(appConfig, logger, mdsNs); err != nil {
			logger.Fatalf("Failed to retrieve mds namespaces. Error %v", err)
		}

		// Custom data manipulation
		_ = dataManipulation.Manipulate(logger, mdsNs)
		distinctProducts := mdsNs.DistinctClusterNamespaceByLabel(logger, appConfig.GroupingLabel)

		// Create a dry run data of sorts to output to CSV to confirm before running
		file, err := os.Create("dry-run.csv")
		if err != nil {
			panic(err)
		}
		defer func(file *os.File) {
			_ = file.Close()
		}(file)
		writer := csv.NewWriter(file)
		defer writer.Flush()
		_ = writer.Write([]string{"Mode", "Zone Name", "Cluster", "Namespace"})
		for productName := range distinctProducts {
			joinedClusters, joinedNamespaces := createClusterNSString(distinctProducts, productName)
			if _, exists := zones.Zones[productName]; !exists {
				_ = writer.Write([]string{"Create", productName, joinedClusters, joinedNamespaces})
			} else {
				_ = writer.Write([]string{"Update", productName, joinedClusters, joinedNamespaces})
			}
		}
		writer.Flush()

		//Process Dry run input
		if appConfig.DryRun {
			fmt.Println("\"dry-run.csv\" has been written, exiting")
			os.Exit(0)
		} else {
			if !appConfig.Silent {
				processDryRun()
			}
		}

		//Iterate through zones, if it does not already exist, we will create a blank one (update later all at once)
		for productName := range distinctProducts {
			fmt.Println("")
			if _, exists := zones.Zones[productName]; !exists {
				var createdZone *zonePayload.Zone
				logger.Debugf("Zone with product name '%s' does NOT exist, creating zone", productName)

				if createdZone, err = createZone(appConfig, logger, zones, productName); err != nil {
					logger.Fatalf("Failed to create new zone '%s'. Error %v", productName, err)
				}

				if err = updateZone(appConfig, logger, zones, distinctProducts, productName, createdZone); err != nil {
					logger.Fatalf("Failed to update zone '%s'. Error %v", productName, err)
				}
			} else {
				logger.Infof("Zone '%s' EXISTS, will update zone", productName)

				zone := zones.Zones[productName]
				if err = updateZone(appConfig, logger, zones, distinctProducts, productName, &zone); err != nil {
					logger.Fatalf("Failed to update zone '%s'. Error %v", productName, err)
				}
			}
		}

		fmt.Println("")
		//Setting static zones to keep
		for key := range appConfig.StaticZones {
			zone := zones.Zones[key]
			zone.Keep = true
			zones.Zones[key] = zone
		}
		//Now we sync/cleanup our zones, deleting any that we have not decided to keep
		for key, zone := range zones.Zones {
			if !zone.Keep {
				logger.Infof("Zone '%s' not marked to keep. Deleting...", key)
			}
		}
	}

	if strings.Contains(strings.ToUpper(appConfig.Mode), "TEAM") {
		fmt.Println("")
		logger.Info("------------------------------")
		logger.Info("Running in 'Create Teams' mode")
		logger.Info("------------------------------")

		// Now lets create some teams
		fmt.Println("")

		// First get the template team to use and re-use
		tb := &teamPayload.TeamBase{}
		configGetTeamByName := sysdighttp.DefaultSysdigRequestConfig(appConfig.SysdigApiEndpoint, appConfig.SecureApiToken)
		if err = tb.GetTeamByName(logger, &configGetTeamByName, appConfig.TeamTemplateName); err != nil {
			logger.Fatalf("Could not retreive team template to use '%s'. Error %v", appConfig.TeamTemplateName, err)
		}

		//Process Team to Zone mapping
		err, tzMapping := getTeamZoneMapping(logger, appConfig)
		if tzMapping == nil || err != nil {
			logger.Fatalf("Failed to load team zone mapping. Error: %v", err)
		}

		// Now create the team(s)
		fmt.Println("")
		for keyName, keyValue := range *tzMapping {
			var teamZoneIDs []int64
			for _, val := range keyValue {
				if zone, exists := zones.Zones[val]; exists {
					teamZoneIDs = append(teamZoneIDs, zone.ID)
				}
			}
			logger.Infof("Team: '%s', ZoneIds %v", keyName, teamZoneIDs)
			if err = createOrUpdateTeam(logger, appConfig, keyName, teamZoneIDs, &tb.Data[0]); err != nil {
				logger.Errorf("Could not create or update team '%s'. Error: %v", keyName, err)
			}
			fmt.Println("")
		}

	}

	if strings.Contains(strings.ToUpper(appConfig.Mode), "MONITOR") {

		fmt.Println("")
		logger.Info("--------------------------------------")
		logger.Info("Running in 'Create Monitor Teams' mode")
		logger.Info("--------------------------------------")

		// Now lets create some teams
		fmt.Println("")

		// Build distinct mapping list for cluster and namespaces
		mdsNs := &mdsNamespaces.NamespacePayload{}
		logger.Infof("Getting mds Namespace list")
		if err = getMDSNamespaces(appConfig, logger, mdsNs); err != nil {
			logger.Fatalf("Failed to retrieve mds namespaces. Error %v", err)
		}

		// Custom data manipulation
		_ = dataManipulation.Manipulate(logger, mdsNs)
		distinctProducts := mdsNs.DistinctClusterNamespaceByLabel(logger, appConfig.GroupingLabel)

		// First get the template team to use and re-use
		tb := &teamPayload.TeamBase{}
		configGetTeamByName := sysdighttp.DefaultSysdigRequestConfig(appConfig.SysdigApiEndpoint, appConfig.SecureApiToken)
		if err = tb.GetTeamByName(logger, &configGetTeamByName, appConfig.TeamTemplateName); err != nil {
			logger.Fatalf("Could not retreive team template to use '%s'. Error %v", appConfig.TeamTemplateName, err)
		}

		for keyName := range distinctProducts {
			teamName := fmt.Sprintf("%s%s", appConfig.TeamPrefix, keyName)
			logger.Infof("Team: '%s'", teamName)
			if err = cRUDTeamMonitor(logger, appConfig, teamName, keyName, &tb.Data[0]); err != nil {
				logger.Errorf("Could not create or update team '%s'. Error: %v", keyName, err)
			}
		}
	}
	logger.Print("Finished...")
}

//TODO Implement chunking for scoping
//TODO Implement dryrun for team creation
//TODO Implement mode va create teams/zones
