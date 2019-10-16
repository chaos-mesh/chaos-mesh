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

package podchaos

import (
	"fmt"
	"time"

	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/client/clientset/versioned"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

const (
	// AnnotationPrefix defines the prefix of annotation key for chaos-operator.
	AnnotationPrefix = "chaos-operator"
)

func GenAnnotationKeyForImage(pc *v1alpha1.PodChaos, containerName string) string {
	return fmt.Sprintf("%s-%s-%s-%s-image", AnnotationPrefix, pc.Name, pc.Spec.Action, containerName)
}

func cleanExpiredExperimentRecords(cli versioned.Interface, podChaos *v1alpha1.PodChaos) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pc, err := cli.PingcapV1alpha1().PodChaoses(podChaos.Namespace).Get(podChaos.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		retentionTime := v1alpha1.DefaultStatusRetentionTime
		if podChaos.Spec.StatusRetentionTime != "" {
			retentionTime, err = time.ParseDuration(podChaos.Spec.StatusRetentionTime)
			if err != nil {
				return err
			}
		}

		pc.Status.CleanExpiredStatusRecords(retentionTime)

		_, err = cli.PingcapV1alpha1().PodChaoses(pc.Namespace).Update(pc)
		return err
	})
}

func setExperimentRecord(
	cli versioned.Interface,
	podChaos *v1alpha1.PodChaos,
	record *v1alpha1.PodChaosExperimentStatus,
) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pc, err := cli.PingcapV1alpha1().PodChaoses(podChaos.Namespace).Get(podChaos.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		pc.Status.SetExperimentRecord(*record)

		pc.Status.Phase = v1alpha1.ChaosPhaseNormal
		pc.Status.Reason = ""

		if record.Phase == v1alpha1.ExperimentPhaseFailed {
			pc.Status.Phase = v1alpha1.ChaosPahseAbnormal
			pc.Status.Reason = "the last chaos experiment failed to execute"
		}

		_, err = cli.PingcapV1alpha1().PodChaoses(pc.Namespace).Update(pc)
		return err
	})
}

func setRecordPods(record *v1alpha1.PodChaosExperimentStatus, action v1alpha1.PodChaosAction, msg string, pods ...v1.Pod) {
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			StartTime: metav1.Now(),
			EndTime:   metav1.Now(),
			Action:    string(action),
			Message:   msg,
		}

		record.SetPods(ps)
	}
}
