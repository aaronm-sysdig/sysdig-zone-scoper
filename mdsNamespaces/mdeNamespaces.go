package mdsNamespaces

import (
	"github.com/aaronm-sysdig/sysdig-zone-scoper/sysdighttp"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type NamespacePayload struct {
	Entities []Entity `json:"entities"`
}

type Entity struct {
	UID         string            `json:"uid"`
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	CustomerID  string            `json:"customerId"`
	TimestampNs int64             `json:"timestampNs"`
	Labels      map[string]string `json:"labels"`
}

type ClusterNamespace struct {
	Cluster   string
	Namespace string
}

func (p *NamespacePayload) GetNamespaces(logger *logrus.Logger, configNS *sysdighttp.SysdigRequestConfig) (err error) {
	var objFetchNamespaceResponse *http.Response
	configNS.Path = "/api/mds/getEntities"
	configNS.Params = map[string]interface{}{
		"type": "k8s_namespace",
	}

	if objFetchNamespaceResponse, err = sysdighttp.SysdigRequest(logger, *configNS); err != nil {
		logger.Fatalf("Could not retrieve namespaces")
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(objFetchNamespaceResponse.Body)

	if err = sysdighttp.ResponseBodyToJson(objFetchNamespaceResponse, &p); err != nil {
		logger.Fatalf("Could not unmarshal namespace payload")
	}
	return nil
}

// DistinctClusterNamespaceByLabel organizes unique clusters and namespaces by product name.
func (p *NamespacePayload) DistinctClusterNamespaceByLabel(logging *logrus.Logger, groupingLabel string) map[string][]ClusterNamespace {
	result := make(map[string][]ClusterNamespace)

	for _, entity := range p.Entities {
		if entity.Labels[groupingLabel] == "" {
			//logging.Debugf("Grouping Label '%s' == '' for namespace '%s'. Skipping...", groupingLabel, entity.Name)
			continue // Skip entities without a product name
		}

		cluster := entity.Labels["kubernetes.cluster.name"]
		namespace := entity.Labels["kubernetes.namespace.name"]
		if cluster == "" || namespace == "" {
			logging.Infof("Cluster == '%s', Namespace == '%s'.  Skipping...", cluster, namespace)
			continue // Skip entities without complete cluster or namespace info
		}

		clusterNamespace := ClusterNamespace{Cluster: cluster, Namespace: namespace}

		// Check if the cluster-namespace pair is already in the slice
		found := false
		for _, cn := range result[entity.Labels[groupingLabel]] {
			if cn.Cluster == cluster && cn.Namespace == namespace {
				found = true
				break
			}
		}
		if !found {
			// Append the cluster-namespace pair to the slice if it's not already present
			logging.Infof("Adding Cluster '%s', Namespace '%s' to distinct slice", cluster, namespace)
			result[entity.Labels[groupingLabel]] = append(result[entity.Labels[groupingLabel]], clusterNamespace)
		}
	}

	return result
}
