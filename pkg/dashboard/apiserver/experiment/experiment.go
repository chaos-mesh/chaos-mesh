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
	"errors"
	"maps"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	experimentStore core.ExperimentStore
	eventStore      core.EventStore
	config          *config.ChaosDashboardConfig
	scheme          *runtime.Scheme
	log             logr.Logger
}

func NewService(
	experimentStore core.ExperimentStore,
	eventStore core.EventStore,
	config *config.ChaosDashboardConfig,
	scheme *runtime.Scheme,
	log logr.Logger,
) *Service {
	return &Service{
		experimentStore: experimentStore,
		eventStore:      eventStore,
		config:          config,
		scheme:          scheme,
		log:             log,
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

		s.log.V(1).Info("Replace query namespace", "ns", ns)
	}

	exps := make([]*apiservertypes.Experiment, 0)
	for k, chaosKind := range v1alpha1.AllKinds() {
		if kind != "" && k != kind {
			continue
		}

		list := chaosKind.SpawnList()
		if err := kubeCli.List(c.Request.Context(), list, &client.ListOptions{Namespace: ns}); err != nil {
			u.SetAPImachineryError(c, err)

			return
		}

		for _, item := range list.GetItems() {
			chaosName := item.GetName()

			if name != "" && chaosName != name {
				continue
			}

			exps = append(exps, &apiservertypes.Experiment{
				Namespace: item.GetNamespace(),
				Name:      chaosName,
				Kind:      item.GetObjectKind().GroupVersionKind().Kind,
				UID:       string(item.GetUID()),
				Created:   item.GetCreationTimestamp().Format(time.RFC3339),
				Status:    status.GetChaosStatus(item.(v1alpha1.InnerObject)),
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
// @Param chaos body map[string]any true "the chaos definition"
// @Success 200 {object} map[string]any
// @Failure 400 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /experiments [post]
func (s *Service) create(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	var exp map[string]any
	if err = u.ShouldBindBodyWithJSON(c, &exp); err != nil {
		return
	}
	kind, ok := exp["kind"].(string)
	if !ok || kind == "" {
		u.SetAPIError(c, u.ErrBadRequest.New("kind must be a non-empty string"))
		return
	}

	if chaosKind, ok := v1alpha1.AllKinds()[kind]; ok {
		chaos := chaosKind.SpawnObject()
		reflect.ValueOf(chaos).Elem().FieldByName("ObjectMeta").Set(reflect.ValueOf(metav1.ObjectMeta{}))

		if err = u.ShouldBindBodyWithJSON(c, chaos); err != nil {
			return
		}

		if err = kubeCli.Create(c.Request.Context(), chaos); err != nil {
			u.SetAPImachineryError(c, err)

			return
		}
	} else {
		u.SetAPIError(c, u.ErrBadRequest.New("Kind %s is not supported", kind))

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
	var exp *core.Experiment

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if exp, err = s.experimentStore.FindByUID(c.Request.Context(), uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.SetAPIError(c, u.ErrNotFound.New("Experiment %s not found", uid))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	chaos, err := getChaosForExperiment(c.Request.Context(), kubeCli, exp)
	if err != nil {
		u.SetAPImachineryError(c, err)
		return
	}
	expDetail := s.chaosDetail(c, chaos)
	if expDetail != nil {
		c.JSON(http.StatusOK, expDetail)
	}
}

func getChaosForExperiment(ctx context.Context, kubeCli client.Client, record *core.Experiment) (client.Object, error) {
	kind, ok := v1alpha1.AllKinds()[record.Kind]
	if !ok {
		return nil, apierrors.NewBadRequest("Kind " + record.Kind + " is not supported")
	}
	chaos := kind.SpawnObject()
	if err := kubeCli.Get(ctx, types.NamespacedName{Namespace: record.Namespace, Name: record.Name}, chaos); err != nil {
		return nil, err
	}
	if chaos.GetUID() != types.UID(record.UID) {
		resource := v1alpha1.SchemeGroupVersion.WithResource(strings.ToLower(record.Kind)).GroupResource()
		return nil, apierrors.NewNotFound(resource, record.Name)
	}
	return chaos, nil
}

func (s *Service) chaosDetail(c *gin.Context, chaos client.Object) *apiservertypes.ExperimentDetail {
	gvk, err := apiutil.GVKForObject(chaos, s.scheme)
	if err != nil {
		u.SetAPImachineryError(c, err)

		return nil
	}

	kind := gvk.Kind

	return &apiservertypes.ExperimentDetail{
		Namespace: reflect.ValueOf(chaos).MethodByName("GetNamespace").Call(nil)[0].String(),
		Name:      reflect.ValueOf(chaos).MethodByName("GetName").Call(nil)[0].String(),
		Kind:      kind,
		UID:       reflect.ValueOf(chaos).MethodByName("GetUID").Call(nil)[0].String(),
		Created:   reflect.ValueOf(chaos).MethodByName("GetCreationTimestamp").Call(nil)[0].Interface().(metav1.Time).Format(time.RFC3339),
		Status:    status.GetChaosStatus(chaos.(v1alpha1.InnerObject)),
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
	if exp, err = s.experimentStore.FindByUID(c.Request.Context(), uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.SetAPIError(c, u.ErrNotFound.New("Experiment %s not found", uid))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	if ok := checkAndDeleteChaos(c, kubeCli, exp, c.DefaultQuery("force", "false")); !ok {
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
		if exp, err = s.experimentStore.FindByUID(c.Request.Context(), uid); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				u.SetAPIError(c, u.ErrNotFound.New("Experiment %s not found", uid))
			} else {
				u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
			}

			return
		}

		if ok := checkAndDeleteChaos(c, kubeCli, exp, force); !ok {
			return
		}
	}

	c.JSON(http.StatusOK, u.ResponseSuccess)
}

func checkAndDeleteChaos(c *gin.Context, kubeCli client.Client, record *core.Experiment, force string) bool {
	ctx := c.Request.Context()
	chaos, err := getChaosForExperiment(ctx, kubeCli, record)
	if err != nil {
		u.SetAPImachineryError(c, err)
		return false
	}

	if force == "true" {
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			updateErr := forceClean(ctx, kubeCli, chaos)
			if !apierrors.IsConflict(updateErr) {
				return updateErr
			}
			// Retry against the current version of the same experiment.
			fresh, err := getChaosForExperiment(ctx, kubeCli, record)
			if err != nil {
				return err
			}
			chaos = fresh
			return updateErr
		})
		if apierrors.IsNotFound(err) {
			u.SetAPImachineryError(c, err)
			return false
		}
		if err != nil || ctx.Err() != nil {
			u.SetAPIError(c, u.ErrInternalServer.New("Forced deletion failed"))
			return false
		}
	}

	uid := types.UID(record.UID)
	if err := kubeCli.Delete(ctx, chaos, client.Preconditions{UID: &uid}); err != nil {
		u.SetAPImachineryError(c, err)
		return false
	}
	return true
}

func forceClean(ctx context.Context, kubeCli client.Client, chaos client.Object) error {
	annotations := chaos.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[finalizers.AnnotationCleanFinalizer] = finalizers.AnnotationCleanFinalizerForced
	chaos.SetAnnotations(annotations)

	return kubeCli.Update(ctx, chaos)
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
	if exp, err = s.experimentStore.FindByUID(c.Request.Context(), uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.SetAPIError(c, u.ErrNotFound.New("Experiment %s not found", uid))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	annotations := map[string]string{
		v1alpha1.PauseAnnotationKey: "true",
	}
	if err = patchExperiment(c.Request.Context(), kubeCli, exp, annotations); err != nil {
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
	if exp, err = s.experimentStore.FindByUID(c.Request.Context(), uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.SetAPIError(c, u.ErrNotFound.New("Experiment %s not found", uid))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	annotations := map[string]string{
		v1alpha1.PauseAnnotationKey: "false",
	}
	if err = patchExperiment(c.Request.Context(), kubeCli, exp, annotations); err != nil {
		u.SetAPImachineryError(c, err)

		return
	}

	c.JSON(http.StatusOK, u.ResponseSuccess)
}

func patchExperiment(ctx context.Context, kubeCli client.Client, record *core.Experiment, annotations map[string]string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		chaos, err := getChaosForExperiment(ctx, kubeCli, record)
		if err != nil {
			return err
		}

		before := chaos.DeepCopyObject().(client.Object)
		currentAnnotations := chaos.GetAnnotations()
		if currentAnnotations == nil {
			currentAnnotations = make(map[string]string, len(annotations))
		}
		maps.Copy(currentAnnotations, annotations)
		chaos.SetAnnotations(currentAnnotations)
		return kubeCli.Patch(ctx, chaos, client.MergeFromWithOptions(before, client.MergeFromWithOptimisticLock{}))
	})
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

		s.log.V(1).Info("Replace query namespace", "ns", ns)
	}

	allChaosStatus := status.AllChaosStatus{}

	g, ctx := errgroup.WithContext(c.Request.Context())
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
