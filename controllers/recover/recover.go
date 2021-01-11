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

package recover

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/finalizer"
)

type Delegate struct {
	client.Client
	Log logr.Logger
	RecoverIntf
}

type RecoverIntf interface {
	RecoverPod(context.Context, *v1.Pod, v1alpha1.InnerObject) error
}

func (r *Delegate) CleanFinalizersAndRecover(ctx context.Context, chaos v1alpha1.InnerObject, finalizers []string, annotations map[string]string) ([]string, error) {
	var result error

	for _, key := range finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		var pod v1.Pod
		err = r.Client.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, &pod)

		if err != nil {
			if !k8serror.IsNotFound(err) {
				result = multierror.Append(result, err)
				continue
			}

			r.Log.Info("Pod not found", "namespace", ns, "name", name)
			finalizers = finalizer.RemoveFromFinalizer(finalizers, key)
			continue
		}

		err = r.RecoverPod(ctx, &pod, chaos)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		finalizers = finalizer.RemoveFromFinalizer(finalizers, key)
	}

	if annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", chaos)
		finalizers = finalizers[:0]
		return finalizers, nil
	}

	return finalizers, result
}
