// Copyright 2019 PingCAP, Inc.
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

package server

import (
	"net/http"

	"github.com/unrolled/render"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/pkg/utils"
)

func (s *Server) services(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rdr := render.New()

	var listOptions = client.ListOptions{}
	listOptions.Namespace = utils.DashboardNamespace
	listOptions.LabelSelector = labels.SelectorFromSet(map[string]string{
		"app.kubernetes.io/component": "grafana",
	})

	var services v1.ServiceList
	err := s.client.List(ctx, &services, &listOptions)
	if err != nil {
		s.log.Error(err, "error while listing services")
	}

	var names []string

	tailingLen := len("-chaos-grafana")
	for _, service := range services.Items {
		if len(service.Name) > tailingLen {
			names = append(names, service.Name[:len(service.Name)-tailingLen]) // remove tailing "-chaos-grafana"
		}
	}

	err = rdr.JSON(w, 200, names)
	if err != nil {
		s.log.Error(err, "error while rendering response")
	}
}
