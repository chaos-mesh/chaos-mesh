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

package experiment

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/core"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Service defines a handler service for experiments.
type Service struct {
	conf    *config.ChaosServerConfig
	kubeCli client.Client
	archive core.ExperimentStore
	event   core.EventStore
}

// NewService returns an experiment service instance.
func NewService(
	conf *config.ChaosServerConfig,
	cli client.Client,
	archive core.ExperimentStore,
	event core.EventStore,
) *Service {
	return &Service{
		conf:    conf,
		kubeCli: cli,
		archive: archive,
		event:   event,
	}
}

// Register mounts our HTTP handler on the mux.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/experiments")

	// TODO: add more api handlers
	endpoint.GET("/", s.listExperiments)
	endpoint.GET("/detail/:name", s.getExperimentDetail)
	endpoint.DELETE("/delete/:ns/:name", s.deleteExperiment)
	endpoint.GET("/state", s.state)
}

// TODO: need to be implemented
func (s *Service) listExperiments(c *gin.Context) {}

// TODO: need to be implemented
func (s *Service) getExperimentDetail(c *gin.Context) {}

// TODO: need to be implemented
func (s *Service) deleteExperiment(c *gin.Context) {}

// getPodChaosState returns the state of PodChaos
func (s *Service) getPodChaosState (stateInfo map[string]int) error {
	var chaosList v1alpha1.PodChaosList
	err := s.kubeCli.List(context.Background(), &chaosList)
	if err != nil {
		return err
	}
	for _, chaos := range chaosList.Items {
		stateInfo["Total"]++
		switch chaos.Status.ChaosStatus.Experiment.Phase {
		case v1alpha1.ExperimentPhaseRunning:
			stateInfo["running"]++
		case v1alpha1.ExperimentPhasePaused:
			stateInfo["Paused"]++
		case v1alpha1.ExperimentPhaseFailed:
			stateInfo["Failed"]++
		case v1alpha1.ExperimentPhaseFinished:
			stateInfo["Finished"]++
		}
	}
	return nil
}

// getIoChaosState returns the state of IoChaos
func (s *Service) getIoChaosState (stateInfo map[string]int) error {
	var chaosList v1alpha1.IoChaosList
	err := s.kubeCli.List(context.Background(), &chaosList)
	if err != nil {
		return err
	}
	for _, chaos := range chaosList.Items {
		stateInfo["Total"]++
		switch chaos.Status.ChaosStatus.Experiment.Phase {
		case v1alpha1.ExperimentPhaseRunning:
			stateInfo["running"]++
		case v1alpha1.ExperimentPhasePaused:
			stateInfo["Paused"]++
		case v1alpha1.ExperimentPhaseFailed:
			stateInfo["Failed"]++
		case v1alpha1.ExperimentPhaseFinished:
			stateInfo["Finished"]++
		}
	}
	return nil
}

// getNetworkChaosState returns the state of NetworkChaos
func (s *Service) getNetworkChaosState (stateInfo map[string]int) error {
	var chaosList v1alpha1.NetworkChaosList
	err := s.kubeCli.List(context.Background(), &chaosList)
	if err != nil {
		return err
	}
	for _, chaos := range chaosList.Items {
		stateInfo["Total"]++
		switch chaos.Status.ChaosStatus.Experiment.Phase {
		case v1alpha1.ExperimentPhaseRunning:
			stateInfo["running"]++
		case v1alpha1.ExperimentPhasePaused:
			stateInfo["Paused"]++
		case v1alpha1.ExperimentPhaseFailed:
			stateInfo["Failed"]++
		case v1alpha1.ExperimentPhaseFinished:
			stateInfo["Finished"]++
		}
	}
	return nil
}

// getTimeChaosState returns the state of TimeChaos
func (s *Service) getTimeChaosState (stateInfo map[string]int) error {
	var chaosList v1alpha1.TimeChaosList
	err := s.kubeCli.List(context.Background(), &chaosList)
	if err != nil {
		return err
	}
	for _, chaos := range chaosList.Items {
		stateInfo["Total"]++
		switch chaos.Status.ChaosStatus.Experiment.Phase {
		case v1alpha1.ExperimentPhaseRunning:
			stateInfo["running"]++
		case v1alpha1.ExperimentPhasePaused:
			stateInfo["Paused"]++
		case v1alpha1.ExperimentPhaseFailed:
			stateInfo["Failed"]++
		case v1alpha1.ExperimentPhaseFinished:
			stateInfo["Finished"]++
		}
	}
	return nil
}

// getKernelChaosState returns the state of KernelChaos
func (s *Service) getKernelChaosState (stateInfo map[string]int) error {
	var chaosList v1alpha1.KernelChaosList
	err := s.kubeCli.List(context.Background(), &chaosList)
	if err != nil {
		return err
	}
	for _, chaos := range chaosList.Items {
		stateInfo["Total"]++
		switch chaos.Status.ChaosStatus.Experiment.Phase {
		case v1alpha1.ExperimentPhaseRunning:
			stateInfo["running"]++
		case v1alpha1.ExperimentPhasePaused:
			stateInfo["Paused"]++
		case v1alpha1.ExperimentPhaseFailed:
			stateInfo["Failed"]++
		case v1alpha1.ExperimentPhaseFinished:
			stateInfo["Finished"]++
		}
	}
	return nil
}

func (s *Service) state (c *gin.Context) {
	data := make(map[string]int)
	data["Total"] = 0
	data["Running"] = 0
	data["Paused"] = 0
	data["Failed"] = 0
	data["Finished"] = 0
	getChaosWrong := gin.H{
		"status": 1003,
		"message": "get chaos wrong",
		"data": make(map[string]int),
	}

	err := s.getPodChaosState(data)
	if err != nil {
		c.JSON(200, getChaosWrong)
		return
	}
	err = s.getIoChaosState(data)
	if err != nil {
		c.JSON(200, getChaosWrong)
		return
	}
	err = s.getNetworkChaosState(data)
	if err != nil {
		c.JSON(200, getChaosWrong)
		return
	}
	err = s.getTimeChaosState(data)
	if err != nil {
		c.JSON(200, getChaosWrong)
		return
	}
	err = s.getKernelChaosState(data)
	if err != nil {
		c.JSON(200, getChaosWrong)
		return
	}

	c.JSON(200, gin.H{
		"status": 0,
		"message": "success",
		"data": data,
	})
}
