package main

import (
	"fmt"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/config"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/dataManipulation"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/mdsNamespaces"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/sysdighttp"
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
	shortFunctionName := functionName[lastSlash+1:] // Default to full string post-last slash if no parenthesis found

	// Try to find the parenthesis and refine shortFunctionName
	if lastSlash != -1 {
		firstParen := strings.Index(functionName[lastSlash:], "(") + lastSlash
		if firstParen > lastSlash {
			shortFunctionName = functionName[lastSlash+1 : firstParen]
		}
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
		clusters = append(clusters, cn.Cluster)
		namespaces = append(namespaces, cn.Namespace)
	}

	var joinedClusters = fmt.Sprintf("\"%s\"", strings.Join(clusters, "\",\""))
	var joinedNamespaces = fmt.Sprintf("\"%s\"", strings.Join(namespaces, "\",\""))
	logger.Debugf("Joined cluster string: '%s'", joinedClusters)
	logger.Debugf("Joined namespace string: '%s'", joinedNamespaces)

	//Update Zone
	var updateZone = &zonePayload.UpdateZone{
		ID:   createdZone.ID,
		Name: createdZone.Name,
		Scopes: []zonePayload.Scope{{
			Rules:      fmt.Sprintf("clusterId in (%s)", joinedClusters),
			TargetType: "kubernetes",
		},
			{
				Rules:      fmt.Sprintf("namespace in (%s)", joinedNamespaces),
				TargetType: "kubernetes",
			},
		},
	}
	configUpdate := sysdighttp.DefaultSysdigRequestConfig(appConfig.SysdigApiEndpoint, appConfig.SecureApiToken)
	configUpdate.JSON = updateZone
	logger.Debugf("Updating zone '%s', zoneID %d", productName, createdZone.ID)
	logger.Debugf(" Cluster: '%s', Namespace: '%s'", updateZone.Scopes[0].Rules, updateZone.Scopes[1].Rules)
	if err = zones.UpdateZone(logger, &configUpdate, updateZone); err != nil {
		logger.Errorf("Could not update zoneId '%d' for '%s'", createdZone.ID, productName)
	}
	return
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

	// we could send a username if we wanted, but no point
	appConfig := &config.Configuration{}
	if err := appConfig.Build(logger); err != nil {
		logger.Fatalf("Could not build configuration. Error %s", err)
	}

	mdsNs := &mdsNamespaces.NamespacePayload{}
	logger.Infof("Getting mds Namespace list")
	if err = getMDSNamespaces(appConfig, logger, mdsNs); err != nil {
		logger.Fatalf("Failed to retrieve mds namespaces. Error %v", err)
	}

	zones := zonePayload.NewZonePayload()
	fmt.Println("")
	logger.Info("Getting list of Zones")
	if err = getZones(appConfig, logger, zones); err != nil {
		logger.Fatalf("Failed to retrieve zones. Error %v", err)
	}

	// Custom data manipulation
	_ = dataManipulation.Manipulate(logger, mdsNs)
	distinctProducts := mdsNs.DistinctClusterNamespaceByLabel(logger, appConfig.GroupingLabel)

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

	logger.Print("Finished...")
}
