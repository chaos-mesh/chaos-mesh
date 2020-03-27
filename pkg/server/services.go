// Copyright 2020 PingCAP, Inc.
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
	"fmt"
	"net/http"

	"github.com/unrolled/render"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"

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

	var services corev1.ServiceList

	err := s.client.List(ctx, &services, &listOptions)
	if err != nil {
		s.log.Error(err, "error while listing services")
	}

	names := make([]string, 0)

	tailLen := len("-chaos-grafana")
	for _, service := range services.Items {
		if len(service.Name) > tailLen {
			name := service.Name[:len(service.Name)-tailLen]

			// Check whether this namespace is still alive
			var prometheusService corev1.Service
			err := s.client.Get(ctx, types.NamespacedName{
				Namespace: service.Labels["prometheus/namespace"],
				Name:      service.Labels["prometheus/name"],
			}, &prometheusService)

			if err != nil {
				s.log.Error(err, "cannot get prometheus", "labels", service.Labels)

				s.log.Info("Destroying namespace related grafana and service", "namespace", name)

				var service corev1.Service
				err := s.client.Get(ctx, types.NamespacedName{
					Namespace: utils.DashboardNamespace,
					Name:      fmt.Sprintf("%s-chaos-grafana", name),
				}, &service)
				if err != nil {
					s.log.Error(err, "get service error")
				}
				err = s.client.Delete(ctx, &service)
				if err != nil {
					s.log.Error(err, "delete service error", "service", service)
				}

				var deployment v1.Deployment
				err = s.client.Get(ctx, types.NamespacedName{
					Namespace: utils.DashboardNamespace,
					Name:      fmt.Sprintf("%s-chaos-grafana", name),
				}, &deployment)
				if err != nil {
					s.log.Error(err, "get deployment error")
				}
				err = s.client.Delete(ctx, &deployment)
				if err != nil {
					s.log.Error(err, "delete deployment error", "deployment", deployment)
				}
			} else {
				s.log.Info("Namespace does exist", "namespace", name)

				names = append(names, name) // remove tailing "-chaos-grafana"
			}
		}
	}

	err = rdr.JSON(w, 200, names)
	if err != nil {
		s.log.Error(err, "error while rendering response")
	}
}
