// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package graph

import (
	"strings"

	"k8s.io/apimachinery/pkg/types"

	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/model"
)

const DefaultNamespace = "default"

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
