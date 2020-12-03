// Copyright 2020 Chaos Mesh Authors.
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

package timechaos

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// get pod current time in nanosecond
func getPodTimeNS(c http.Client, port uint16) (*time.Time, error) {
	resp, err := c.Get(fmt.Sprintf("http://localhost:%d/time", port))
	if err != nil {
		return nil, err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	t, err := time.Parse(time.RFC3339Nano, string(out))
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func createTimeChaos(ns string, duration string, cron string) *v1alpha1.TimeChaos {
	return &v1alpha1.TimeChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "timer-time-chaos",
			Namespace: ns,
		},
		Spec: v1alpha1.TimeChaosSpec{
			Selector: v1alpha1.SelectorSpec{
				Namespaces:     []string{ns},
				LabelSelectors: map[string]string{"app": "timer"},
			},
			Mode:       v1alpha1.OnePodMode,
			Duration:   pointer.StringPtr(duration),
			TimeOffset: "-1h",
			Scheduler: &v1alpha1.SchedulerSpec{
				Cron: cron,
			},
		},
	}
}

func waitChaosWorking(initTime *time.Time, c http.Client, port uint16) error {
	return wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		podTime, err := getPodTimeNS(c, port)
		framework.ExpectNoError(err, "failed to get pod time")
		if podTime.Before(*initTime) {
			return true, nil
		}
		return false, nil
	})
}

func waitChaosStatus(ctx context.Context, status v1alpha1.ExperimentPhase, chaosKey types.NamespacedName, initTime *time.Time, cli client.Client) error {
	return wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		chaos := &v1alpha1.TimeChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get time chaos error")
		if chaos.Status.Experiment.Phase == status {
			return true, nil
		}
		return false, err
	})
}
