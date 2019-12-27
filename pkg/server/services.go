package server

import (
	"github.com/pingcap/chaos-mesh/pkg/utils"
	"github.com/unrolled/render"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	tailing_len := len("-chaos-grafana")
	for _, service := range services.Items {
		if len(service.Name) > tailing_len {
			names = append(names, service.Name[:len(service.Name)-tailing_len]) // remove tailing "-chaos-grafana"
		}
	}

	err = rdr.JSON(w, 200, names)
	if err != nil {
		s.log.Error(err, "error while rendering response")
	}
}
