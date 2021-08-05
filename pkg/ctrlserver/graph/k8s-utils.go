package graph

import (
	"strings"

	"k8s.io/apimachinery/pkg/types"

	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/model"
)

func componentLabels(component model.Component) map[string]string {
	var componentLabel string
	switch component {
	case model.ComponentManager:
		componentLabel = "controller-manager"
	case model.ComponentDaemon:
		componentLabel = "chaos-daemon"
	case model.ComponentDashboard:
		componentLabel = "chaos-dashboard"
	case model.ComponentDNSServer:
		componentLabel = "chaos-dns-server"
	default:
		return nil
	}
	return map[string]string{
		"app.kubernetes.io/component": componentLabel,
	}
}

func parseNamespacedName(namespacedName string) types.NamespacedName {
	parts := strings.Split(namespacedName, "/")
	return types.NamespacedName{
		Namespace: parts[0],
		Name:      parts[1],
	}
}
