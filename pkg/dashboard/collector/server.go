// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package collector

import (
	"context"
	"net"
	"os"
	"strconv"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = v1alpha1.AddToScheme(scheme)
}

// Server defines a server to manage collectors.
type Server struct {
	Manager ctrl.Manager
	logger  logr.Logger
}

// NewServer returns a CollectorServer and Client.
func NewServer(
	conf *config.ChaosDashboardConfig,
	experimentArchive core.ExperimentStore,
	scheduleArchive core.ScheduleStore,
	event core.EventStore,
	workflowStore core.WorkflowStore,
	logger logr.Logger,
) (*Server, client.Client, client.Reader, *runtime.Scheme) {
	s := &Server{logger: logger}

	// namespace scoped
	options := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: net.JoinHostPort(conf.MetricHost, strconv.Itoa(conf.MetricPort)),
		LeaderElection:     conf.EnableLeaderElection,
		Port:               9443,
	}
	if conf.ClusterScoped {
		logger.Info("Chaos controller manager is running in cluster scoped mode.")
	} else {
		logger.Info("Chaos controller manager is running in namespace scoped mode.", "targetNamespace", conf.TargetNamespace)
		options.Namespace = conf.TargetNamespace
	}

	var err error

	cfg := ctrl.GetConfigOrDie()

	if conf.QPS > 0 {
		cfg.QPS = conf.QPS
		cfg.Burst = conf.Burst
	}

	s.Manager, err = ctrl.NewManager(cfg, options)
	if err != nil {
		logger.Error(err, "unable to start collector")
		os.Exit(1)
	}

	if conf.SecurityMode {
		clientpool.K8sClients, err = clientpool.NewClientPool(cfg, scheme, 100)
		if err != nil {
			// this should never happen
			logger.Error(err, "fail to create client pool")
			os.Exit(1)
		}
	} else {
		clientpool.K8sClients, err = clientpool.NewLocalClient(cfg, scheme)
		if err != nil {
			logger.Error(err, "fail to create client pool")
			os.Exit(1)
		}
	}

	for kind, chaosKind := range v1alpha1.AllKinds() {
		if err = (&ChaosCollector{
			Client:  s.Manager.GetClient(),
			Log:     logger.WithName(kind),
			archive: experimentArchive,
			event:   event,
		}).Setup(s.Manager, chaosKind.SpawnObject()); err != nil {
			logger.Error(err, "unable to create collector", "collector", kind)
			os.Exit(1)
		}
	}

	if err = (&ScheduleCollector{
		Client:  s.Manager.GetClient(),
		Log:     logger.WithName("schedule-collector").WithName(v1alpha1.KindSchedule),
		archive: scheduleArchive,
	}).Setup(s.Manager, &v1alpha1.Schedule{}); err != nil {
		logger.Error(err, "unable to create collector", "collector", v1alpha1.KindSchedule)
		os.Exit(1)
	}

	if err = (&EventCollector{
		Client: s.Manager.GetClient(),
		Log:    logger.WithName("event-collector").WithName("Event"),
		event:  event,
	}).Setup(s.Manager, &v1.Event{}); err != nil {
		logger.Error(err, "unable to create collector", "collector", v1alpha1.KindSchedule)
		os.Exit(1)
	}

	if err = (&WorkflowCollector{
		kubeClient: s.Manager.GetClient(),
		Log:        logger.WithName("workflow-collector").WithName(v1alpha1.KindWorkflow),
		store:      workflowStore,
	}).Setup(s.Manager, &v1alpha1.Workflow{}); err != nil {
		logger.Error(err, "unable to create collector", "collector", v1alpha1.KindWorkflow)
		os.Exit(1)
	}

	return s, s.Manager.GetClient(), s.Manager.GetAPIReader(), s.Manager.GetScheme()
}

// Register starts collectors manager.
func Register(ctx context.Context, s *Server) {
	go func() {
		s.logger.Info("Starting collector")
		if err := s.Manager.Start(ctx); err != nil {
			s.logger.Error(err, "could not start collector")
			os.Exit(1)
		}
	}()
}
