package dataManipulation

import (
	"github.com/sirupsen/logrus"
	"strings"
	"sysdig-zone-scoper/mdsNamespaces"
)

func Manipulate(logger *logrus.Logger, mdsNs *mdsNamespaces.NamespacePayload) error {
	for i := range mdsNs.Entities {
		entity := &mdsNs.Entities[i]

		if supportGroup, exists := entity.Labels["kubernetes.namespace.label.SupportGroup"]; exists {
			modifiedSupportGroup := strings.Replace(supportGroup, "_", " ", -1)
			modifiedSupportGroup = strings.Replace(modifiedSupportGroup, "API SUPPORT", "API Support", -1)
			entity.Labels["kubernetes.namespace.label.SupportGroup"] = modifiedSupportGroup
			if supportGroup != modifiedSupportGroup {
				logger.Debugf("Replaced '%s' with '%s'", supportGroup, modifiedSupportGroup)
			}
		}

		if supportGroup, exists := entity.Labels["kubernetes.namespace.label.SupportGroup"]; exists && supportGroup == "" {
			entity.Labels["kubernetes.namespace.label.SupportGroup"] = "KubeOperations"
			logger.Debug("SupportGroup is empty, replacing with 'KubeOperations")
		}

		productName, exists := entity.Labels["kubernetes.namespace.label.ProductName"]
		if exists && productName == "" {
			entity.Labels["kubernetes.namespace.label.ProductName"] = "KubeOperations"
			logger.Debug("ProductName is empty, replacing with 'KubeOperations")
		}
	}
	return nil
}
