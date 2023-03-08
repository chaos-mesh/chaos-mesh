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

package podhttpchaos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tproxyconfig"
)

// Reconciler applys podhttpchaos
type Reconciler struct {
	client.Client

	Recorder                 record.EventRecorder
	Log                      logr.Logger
	ChaosDaemonClientBuilder *chaosdaemon.ChaosDaemonClientBuilder
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	obj := &v1alpha1.PodHttpChaos{}

	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("chaos not found")
		} else {
			// TODO: handle this error
			r.Log.Error(err, "unable to get chaos")
		}
		return ctrl.Result{}, nil
	}

	if obj.ObjectMeta.Generation <= obj.Status.ObservedGeneration && obj.Status.FailedMessage == "" {
		r.Log.Info("the target pod has been up to date", "pod", obj.Namespace+"/"+obj.Name)
		return ctrl.Result{}, nil
	}

	r.Log.Info("updating http chaos", "pod", obj.Namespace+"/"+obj.Name, "spec", obj.Spec)

	pod := &v1.Pod{}

	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}, pod)
	if err != nil {
		err = errors.Wrapf(err, "failed to apply for pod %s/%s", pod.Namespace, pod.Name)
		r.Log.Error(err, "fail to find pod")
		return ctrl.Result{}, nil
	}

	observedGeneration := obj.ObjectMeta.Generation
	pid := obj.Status.Pid
	startTime := obj.Status.StartTime

	defer func() {
		var failedMessage string
		if err != nil {
			failedMessage = err.Error()
		}

		updateError := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			obj := &v1alpha1.PodHttpChaos{}

			if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
				r.Log.Error(err, "unable to get chaos")
				return err
			}

			obj.Status.FailedMessage = failedMessage
			obj.Status.ObservedGeneration = observedGeneration
			obj.Status.Pid = pid
			obj.Status.StartTime = startTime

			return r.Client.Status().Update(context.TODO(), obj)
		})

		if updateError != nil {
			updateError = errors.Wrapf(updateError, "failed to apply for pod %s/%s", pod.Namespace, pod.Name)
			r.Log.Error(updateError, "fail to update")
			r.Recorder.Eventf(obj, "Normal", "Failed", "Failed to update status: %s", updateError.Error())
		}
	}()

	pbClient, err := r.ChaosDaemonClientBuilder.Build(ctx, pod, &types.NamespacedName{
		Namespace: obj.Namespace,
		Name:      obj.Name,
	})
	if err != nil {
		err = errors.Wrapf(err, "failed to apply for pod %s/%s", pod.Namespace, pod.Name)
		r.Recorder.Event(obj, "Warning", "Failed", err.Error())
		return ctrl.Result{Requeue: true}, nil
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		err = errors.Wrapf(utils.ErrContainerNotFound, "pod %s/%s has empty container status", pod.Namespace, pod.Name)
		r.Recorder.Event(obj, "Warning", "Failed", err.Error())
		return ctrl.Result{}, nil
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	rules := make([]v1alpha1.PodHttpChaosBaseRule, 0)
	proxyPortsMap := make(map[uint32]bool)

	for _, rule := range obj.Spec.Rules {
		proxyPortsMap[uint32(rule.Port)] = true
		rules = append(rules, rule.PodHttpChaosBaseRule)
	}

	var proxyPorts []uint32
	for port := range proxyPortsMap {
		proxyPorts = append(proxyPorts, port)
	}

	inputRules, err := json.Marshal(rules)
	if err != nil {
		err = errors.Wrapf(err, "failed to apply for pod %s/%s", pod.Namespace, pod.Name)
		r.Recorder.Event(obj, "Warning", "Failed", err.Error())
		return ctrl.Result{}, nil
	}

	inputTLS := []byte("")
	if obj.Spec.TLS != nil {
		tlsKeys := obj.Spec.TLS
		secret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tlsKeys.SecretName,
				Namespace: tlsKeys.SecretNamespace,
			},
		}
		if err := r.Client.Get(context.TODO(), req.NamespacedName, &secret); err != nil {
			r.Log.Error(err, "unable to get secret")
			return ctrl.Result{}, nil
		}

		cert, ok := secret.Data[tlsKeys.CertName]
		if !ok {
			err = errors.Wrapf(err, "get cert %s", tlsKeys.CertName)
			r.Recorder.Event(obj, "Warning", "Failed", err.Error())
			return ctrl.Result{}, nil
		}

		key, ok := secret.Data[tlsKeys.KeyName]
		if !ok {
			err = errors.Wrapf(err, "get key %s", tlsKeys.KeyName)
			r.Recorder.Event(obj, "Warning", "Failed", err.Error())
			return ctrl.Result{}, nil
		}

		var ca []byte
		if tlsKeys.CAName != nil {
			ca, ok = secret.Data[*tlsKeys.CAName]
			if !ok {
				err = errors.Wrapf(err, "get ca %s", *tlsKeys.CAName)
				r.Recorder.Event(obj, "Warning", "Failed", err.Error())
				return ctrl.Result{}, nil
			}
		}

		tlsConfig := tproxyconfig.TLSConfig{
			CertFile: tproxyconfig.TLSConfigItem{
				Type:  "Contents",
				Value: cert,
			},
			KeyFile: tproxyconfig.TLSConfigItem{
				Type:  "Contents",
				Value: key,
			},
		}

		if ca != nil {
			tlsConfig.CAFile = &tproxyconfig.TLSConfigItem{
				Type:  "Contents",
				Value: ca,
			}
		}

		inputTLS, err = json.Marshal(tlsConfig)
		if err != nil {
			err = errors.Wrapf(err, "apply for pod %s/%s", pod.Namespace, pod.Name)
			r.Recorder.Event(obj, "Warning", "Failed", err.Error())
			return ctrl.Result{}, nil
		}
	}

	r.Log.Info("input with", "rules", string(inputRules))

	res, err := pbClient.ApplyHttpChaos(ctx, &pb.ApplyHttpChaosRequest{
		Rules:       string(inputRules),
		Tls:         string(inputTLS),
		ProxyPorts:  proxyPorts,
		ContainerId: containerID,

		Instance:  obj.Status.Pid,
		StartTime: obj.Status.StartTime,
		EnterNS:   true,
	})
	if err != nil {
		err = errors.Wrapf(err, "failed to apply for pod %s/%s", pod.Namespace, pod.Name)
		r.Recorder.Event(obj, "Warning", "Failed", err.Error())
		return ctrl.Result{Requeue: true}, nil
	}

	if res.StatusCode != http.StatusOK {
		err = errors.Wrapf(fmt.Errorf("%s", res.Error),
			"failed to apply for pod %s/%s, status(%d)",
			pod.Namespace, pod.Name, res.StatusCode)
		r.Recorder.Event(obj, "Warning", "Failed", err.Error())
		return ctrl.Result{Requeue: true}, nil
	}

	pid = res.Instance
	startTime = res.StartTime

	return ctrl.Result{}, nil
}
