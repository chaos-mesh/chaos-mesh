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

package experiment

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/jinzhu/gorm"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/finalizers"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config"
	apiservertypes "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/types"
	u "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/status"
)

// Service defines a handler service for experiments.
type Service struct {
	archive core.ExperimentStore
	event   core.EventStore
	config  *config.ChaosDashboardConfig
	scheme  *runtime.Scheme
	log     logr.Logger
}

func NewService(
	archive core.ExperimentStore,
	event core.EventStore,
	config *config.ChaosDashboardConfig,
	scheme *runtime.Scheme,
	log logr.Logger,
) *Service {
	return &Service{
		archive: archive,
		event:   event,
		config:  config,
		scheme:  scheme,
		log:     log,
	}
}

// Register experiments RouterGroup.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/experiments")

	endpoint.GET("", s.list)
	endpoint.POST("", s.create)
	endpoint.GET("/:uid", s.get)
	endpoint.DELETE("/:uid", s.delete)
	endpoint.DELETE("", s.batchDelete)
	endpoint.PUT("/pause/:uid", s.pause)
	endpoint.PUT("/start/:uid", s.start)
	endpoint.GET("/state", s.state)
}

// @Summary List chaos experiments.
// @Description Get chaos experiments from k8s clusters in real time.
// @Tags experiments
// @Produce json
// @Param namespace query string false "filter exps by namespace"
// @Param name query string false "filter exps by name"
// @Param kind query string false "filter exps by kind" Enums(PodChaos, NetworkChaos, IOChaos, StressChaos, KernelChaos, TimeChaos, DNSChaos, AWSChaos, GCPChaos, JVMChaos, HTTPChaos)
// @Param status query string false "filter exps by status" Enums(Injecting, Running, Finished, Paused)
// @Success 200 {array} apiservertypes.Experiment
// @Failure 400 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /experiments [get]
func (s *Service) list(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	ns, name, kind := c.Query("namespace"), c.Query("name"), c.Query("kind")

	if ns == "" && !s.config.ClusterScoped && s.config.TargetNamespace != "" {
		ns = s.config.TargetNamespace

		s.log.V(1).Info("Replace query namespace with", ns)
	}

	exps := make([]*apiservertypes.Experiment, 0)
	for k, chaosKind := range v1alpha1.AllKinds() {
		if kind != "" && k != kind {
			continue
		}

		list := chaosKind.SpawnList()
		if err := kubeCli.List(context.Background(), list, &client.ListOptions{Namespace: ns}); err != nil {
			u.SetAPImachineryError(c, err)

			return
		}

		for _, item := range list.GetItems() {
			chaosName := item.GetName()

			if name != "" && chaosName != name {
				continue
			}

			exps = append(exps, &apiservertypes.Experiment{
				ObjectBase: core.ObjectBase{
					Namespace: item.GetNamespace(),
					Name:      chaosName,
					Kind:      item.GetObjectKind().GroupVersionKind().Kind,
					UID:       string(item.GetUID()),
					Created:   item.GetCreationTimestamp().Format(time.RFC3339),
				},
				Status: status.GetChaosStatus(item.(v1alpha1.InnerObject)),
			})
		}
	}

	sort.Slice(exps, func(i, j int) bool {
		return exps[i].Created > exps[j].Created
	})

	c.JSON(http.StatusOK, exps)
}

// @Summary Create a new chaos experiment.
// @Description Pass a JSON object to create a new chaos experiment. The schema for JSON is the same as the YAML schema for the Kubernetes object.
// @Tags experiments
// @Accept json
// @Produce json
// @Param chaos body map[string]interface{} true "the chaos definition"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /experiments [post]
func (s *Service) create(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	var exp map[string]interface{}
	if err = u.ShouldBindBodyWithJSON(c, &exp); err != nil {
		return
	}
	kind := exp["kind"].(string)

	if chaosKind, ok := v1alpha1.AllKinds()[kind]; ok {
		chaos := chaosKind.SpawnObject()
		reflect.ValueOf(chaos).Elem().FieldByName("ObjectMeta").Set(reflect.ValueOf(metav1.ObjectMeta{}))

		if err = u.ShouldBindBodyWithJSON(c, chaos); err != nil {
			return
		}

		if err = kubeCli.Create(context.Background(), chaos); err != nil {
			u.SetAPImachineryError(c, err)

			return
		}
	} else {
		u.SetAPIError(c, u.ErrBadRequest.New("Kind "+kind+" is not supported"))

		return
	}

	c.JSON(http.StatusOK, exp)
}

// @Summary Get a chaos experiment.
// @Description Get the chaos experiment's detail by uid.
// @Tags experiments
// @Produce json
// @Param uid path string true "the experiment uid"
// @Success 200 {object} apiservertypes.ExperimentDetail
// @Failure 400 {object} u.APIError
// @Failure 404 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /experiments/{uid} [get]
func (s *Service) get(c *gin.Context) {
	var (
		exp       *core.Experiment
		expDetail *apiservertypes.ExperimentDetail
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrNotFound.New("Experiment "+uid+" not found"))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	ns, name, kind := exp.Namespace, exp.Name, exp.Kind

	if chaosKind, ok := v1alpha1.AllKinds()[kind]; ok {
		expDetail = s.findChaosInCluster(c, kubeCli, types.NamespacedName{Namespace: ns, Name: name}, chaosKind.SpawnObject())

		if expDetail == nil {
			return
		}
	} else {
		u.SetAPIError(c, u.ErrBadRequest.New("Kind "+kind+" is not supported"))

		return
	}

	c.JSON(http.StatusOK, expDetail)
}

func (s *Service) findChaosInCluster(c *gin.Context, kubeCli client.Client, namespacedName types.NamespacedName, chaos client.Object) *apiservertypes.ExperimentDetail {
	if err := kubeCli.Get(context.Background(), namespacedName, chaos); err != nil {
		u.SetAPImachineryError(c, err)

		return nil
	}

	gvk, err := apiutil.GVKForObject(chaos, s.scheme)
	if err != nil {
		u.SetAPImachineryError(c, err)

		return nil
	}

	kind := gvk.Kind

	return &apiservertypes.ExperimentDetail{
		Experiment: apiservertypes.Experiment{
			ObjectBase: core.ObjectBase{
				Namespace: reflect.ValueOf(chaos).MethodByName("GetNamespace").Call(nil)[0].String(),
				Name:      reflect.ValueOf(chaos).MethodByName("GetName").Call(nil)[0].String(),
				Kind:      kind,
				UID:       reflect.ValueOf(chaos).MethodByName("GetUID").Call(nil)[0].String(),
				Created:   reflect.ValueOf(chaos).MethodByName("GetCreationTimestamp").Call(nil)[0].Interface().(metav1.Time).Format(time.RFC3339),
			},
			Status: status.GetChaosStatus(chaos.(v1alpha1.InnerObject)),
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       kind,
			},
			Meta: core.KubeObjectMeta{
				Namespace:   reflect.ValueOf(chaos).Elem().FieldByName("Namespace").String(),
				Name:        reflect.ValueOf(chaos).Elem().FieldByName("Name").String(),
				Labels:      reflect.ValueOf(chaos).Elem().FieldByName("Labels").Interface().(map[string]string),
				Annotations: reflect.ValueOf(chaos).Elem().FieldByName("Annotations").Interface().(map[string]string),
			},
			Spec: reflect.ValueOf(chaos).Elem().FieldByName("Spec").Interface(),
		},
	}
}

// @Summary Delete a chaos experiment.
// @Description Delete the chaos experiment by uid.
// @Tags experiments
// @Produce json
// @Param uid path string true "the experiment uid"
// @Param force query string false "force" Enums(true, false)
// @Success 200 {object} u.Response
// @Failure 400 {object} u.APIError
// @Failure 404 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /experiments/{uid} [delete]
func (s *Service) delete(c *gin.Context) {
	var (
		exp *core.Experiment
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrNotFound.New("Experiment "+uid+" not found"))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	ns, name, kind, force := exp.Namespace, exp.Name, exp.Kind, c.DefaultQuery("force", "false")
	if ok := checkAndDeleteChaos(c, kubeCli, types.NamespacedName{Namespace: ns, Name: name}, kind, force); !ok {
		return
	}

	c.JSON(http.StatusOK, u.ResponseSuccess)
}

// @Summary Batch delete chaos experiments.
// @Description Batch delete chaos experiments by uids.
// @Tags experiments
// @Produce json
// @Param uids query string true "the experiment uids, split with comma. Example: ?uids=uid1,uid2"
// @Param force query string false "force" Enums(true, false)
// @Success 200 {object} u.Response
// @Failure 400 {object} u.APIError
// @Failure 404 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /experiments [delete]
func (s *Service) batchDelete(c *gin.Context) {
	var (
		exp *core.Experiment
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uids := c.Query("uids")
	if uids == "" {
		u.SetAPIError(c, u.ErrInternalServer.New("The uids cannot be empty"))

		return
	}

	uidSlice, force := strings.Split(uids, ","), c.DefaultQuery("force", "false")

	if len(uidSlice) > 100 {
		u.SetAPIError(c, u.ErrInternalServer.New("Too many uids, please delete less than 100 at a time"))

		return
	}

	for _, uid := range uidSlice {
		if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
			if gorm.IsRecordNotFoundError(err) {
				u.SetAPIError(c, u.ErrNotFound.New("Experiment "+uid+" not found"))
			} else {
				u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
			}

			return
		}

		ns, name, kind := exp.Namespace, exp.Name, exp.Kind
		if ok := checkAndDeleteChaos(c, kubeCli, types.NamespacedName{Namespace: ns, Name: name}, kind, force); !ok {
			return
		}
	}

	c.JSON(http.StatusOK, u.ResponseSuccess)
}

func checkAndDeleteChaos(c *gin.Context, kubeCli client.Client, namespacedName types.NamespacedName, kind string, force string) bool {
	var (
		chaosKind *v1alpha1.ChaosKind
		ok        bool
		err       error
	)

	if chaosKind, ok = v1alpha1.AllKinds()[kind]; !ok {
		u.SetAPIError(c, u.ErrBadRequest.New("Kind "+kind+" is not supported"))

		return false
	}

	ctx := context.Background()
	chaos := chaosKind.SpawnObject()

	if err = kubeCli.Get(ctx, namespacedName, chaos); err != nil {
		u.SetAPImachineryError(c, err)

		return false
	}

	if force == "true" {
		if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return forceClean(kubeCli, chaos)
		}); err != nil {
			u.SetAPIError(c, u.ErrInternalServer.New("Forced deletion failed"))

			return false
		}
	}

	if err := kubeCli.Delete(ctx, chaos); err != nil {
		u.SetAPImachineryError(c, err)

		return false
	}

	return true
}

func forceClean(kubeCli client.Client, chaos client.Object) error {
	annotations := chaos.(metav1.Object).GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[finalizers.AnnotationCleanFinalizer] = finalizers.AnnotationCleanFinalizerForced
	chaos.(metav1.Object).SetAnnotations(annotations)

	return kubeCli.Update(context.Background(), chaos)
}

// @Summary Pause a chaos experiment.
// @Description Pause a chaos experiment.
// @Tags experiments
// @Produce json
// @Param uid path string true "the experiment uid"
// @Success 200 {object} u.Response
// @Failure 400 {object} u.APIError
// @Failure 404 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /experiments/pause/{uid} [put]
func (s *Service) pause(c *gin.Context) {
	var exp *core.Experiment

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrNotFound.New("Experiment "+uid+" not found"))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	annotations := map[string]string{
		v1alpha1.PauseAnnotationKey: "true",
	}
	if err = patchExperiment(kubeCli, exp, annotations); err != nil {
		u.SetAPImachineryError(c, err)

		return
	}

	c.JSON(http.StatusOK, u.ResponseSuccess)
}

// @Summary Start a chaos experiment.
// @Description Start a chaos experiment.
// @Tags experiments
// @Produce json
// @Param uid path string true "the experiment uid"
// @Success 200 {object} u.Response
// @Failure 400 {object} u.APIError
// @Failure 404 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /experiments/start/{uid} [put]
func (s *Service) start(c *gin.Context) {
	var exp *core.Experiment

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrNotFound.New("Experiment "+uid+" not found"))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	annotations := map[string]string{
		v1alpha1.PauseAnnotationKey: "false",
	}
	if err = patchExperiment(kubeCli, exp, annotations); err != nil {
		u.SetAPImachineryError(c, err)

		return
	}

	c.JSON(http.StatusOK, u.ResponseSuccess)
}

func patchExperiment(kubeCli client.Client, exp *core.Experiment, annotations map[string]string) error {
	chaos := v1alpha1.AllKinds()[exp.Kind].SpawnObject()

	if err := kubeCli.Get(context.Background(), types.NamespacedName{Namespace: exp.Namespace, Name: exp.Name}, chaos); err != nil {
		return err
	}

	var mergePatch []byte
	mergePatch, _ = json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": annotations,
		},
	})

	return kubeCli.Patch(context.Background(), chaos, client.RawPatch(types.MergePatchType, mergePatch))
}

// @Summary Get the status of all experiments.
// @Description Get the status of all experiments.
// @Tags experiments
// @Produce json
// @Param namespace query string false "namespace"
// @Success 200 {object} status.AllChaosStatus
// @Failure 400 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /experiments/state [get]
func (s *Service) state(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	ns := c.Query("namespace")
	if ns == "" && !s.config.ClusterScoped && s.config.TargetNamespace != "" {
		ns = s.config.TargetNamespace

		s.log.V(1).Info("Replace query namespace with", ns)
	}

	allChaosStatus := status.AllChaosStatus{}

	g, ctx := errgroup.WithContext(context.Background())
	m := &sync.Mutex{}

	var listOptions []client.ListOption
	listOptions = append(listOptions, &client.ListOptions{Namespace: ns})

	for _, chaosKind := range v1alpha1.AllKinds() {
		list := chaosKind.SpawnList()

		g.Go(func() error {
			if err := kubeCli.List(ctx, list, listOptions...); err != nil {
				return err
			}
			m.Lock()

			for _, item := range list.GetItems() {
				s := status.GetChaosStatus(item.(v1alpha1.InnerObject))

				switch s {
				case status.Injecting:
					allChaosStatus.Injecting++
				case status.Running:
					allChaosStatus.Running++
				case status.Finished:
					allChaosStatus.Finished++
				case status.Paused:
					allChaosStatus.Paused++
				}
			}

			m.Unlock()
			return nil
		})
	}

	if err = g.Wait(); err != nil {
		u.SetAPImachineryError(c, err)

		return
	}

	c.JSON(http.StatusOK, allChaosStatus)
}
