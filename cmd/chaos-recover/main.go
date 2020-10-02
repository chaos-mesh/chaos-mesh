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

package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type metadata struct {
	Type          string
	Package       string
	Manager       string
	ImportManager string
}

var (
	log         = zap.Logger(true)
	withManager = map[string]string{
		"iochaos":                     "IoChaos",
		"networkchaos/partition":      "NetworkChaos",
		"networkchaos/trafficcontrol": "NetworkChaos",
	}
	withoutManager = map[string]string{
		"kernelchaos":         "KernelChaos",
		"stresschaos":         "StressChaos",
		"timechaos":           "TimeChaos",
		"podchaos/podfailure": "PodChaos",
	}
)

const recoverTemplate = `// Copyright 2020 Chaos Mesh Authors.
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

package {{.Package}}

import (
	"context"
	"errors"

	"github.com/hashicorp/go-multierror"
	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
	{{.ImportManager}}
)

// Recover means the reconciler recovers the chaos action
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	somechaos, ok := chaos.(*v1alpha1.{{.Type}})
	if !ok {
		err := errors.New("chaos is not {{.Type}}")
		r.Log.Error(err, "chaos is not {{.Type}}", "chaos", chaos)
		return err
	}

	if err := r.cleanFinalizersAndRecover(ctx, somechaos); err != nil {
		return err
	}
	r.Event(somechaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")

	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, chaos *v1alpha1.{{.Type}}) error {
	var result error
`

const cleanWithManagerTemplate = `
	source := chaos.Namespace + "/" + chaos.Name
	m := {{.Manager}}.New(source, r.Log, r.Client, r.Reader)

	for _, key := range chaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		_ = m.WithInit(types.NamespacedName{
			Namespace: ns,
			Name:      name,
		})

		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		err = m.Commit(ctx)
		// if pod not found or not running, directly return and giveup recover.
		if err != nil && err != {{.Manager}}.ErrPodNotFound && err != {{.Manager}}.ErrPodNotRunning {
			r.Log.Error(err, "fail to commit")
		}

		chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, key)
	}
	r.Log.Info("After recovering", "finalizers", chaos.Finalizers)

	if chaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", chaos)
		chaos.Finalizers = chaos.Finalizers[:0]
		return nil
	}

	return result
}
`

const cleanWithoutManagerTemplate = `
	for _, key := range chaos.Finalizers {
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
			chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, key)
			continue
		}

		err = r.recoverPod(ctx, &pod, chaos)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, key)
	}

	if chaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", chaos)
		chaos.Finalizers = chaos.Finalizers[:0]
		return nil
	}

	return result
}
`

func generateImpl(meta metadata, temp string) string {
	tmpl, err := template.New("impl").Parse(temp)
	if err != nil {
		log.Error(err, "fail to build template")
		return ""
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, &meta)
	if err != nil {
		log.Error(err, "fail to execute template")
		return ""
	}

	return buf.String()
}

func main() {
	// generate `Recover` for controllers with manager
	for path, chaos := range withManager {
		file, err := os.Create("./controllers/" + path + "/zz_generated.recover.go")
		if err != nil {
			log.Error(err, "fail to generate recover file")
		}

		chaosLower := strings.ToLower(chaos)
		chaosManager := "pod" + chaosLower + "manager"

		generatedCode := generateImpl(metadata{
			Type:          chaos,
			Package:       path[strings.LastIndex(path, "/")+1:],
			ImportManager: `"github.com/chaos-mesh/chaos-mesh/controllers/` + chaosLower + `/` + chaosManager + `"`,
		}, recoverTemplate)
		generatedCode += generateImpl(metadata{
			Manager: chaosManager,
		}, cleanWithManagerTemplate)
		fmt.Fprint(file, generatedCode)
	}

	// generate `Recover` for controllers without manager
	for path, chaos := range withoutManager {
		file, err := os.Create("./controllers/" + path + "/zz_generated.recover.go")
		if err != nil {
			log.Error(err, "fail to generate recover file")
		}

		generatedCode := generateImpl(metadata{
			Type:    chaos,
			Package: path[strings.LastIndex(path, "/")+1:],
		}, recoverTemplate)
		generatedCode += generateImpl(metadata{}, cleanWithoutManagerTemplate)
		fmt.Fprint(file, generatedCode)
	}
	return
}
