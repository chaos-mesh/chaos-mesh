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
	"github.com/jinzhu/gorm"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/finalizers"
	u "github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/status"
)

var log = u.Log.WithName("experiments")

// Service defines a handler service for experiments.
type Service struct {
	archive core.ExperimentStore
	event   core.EventStore
	conf    *config.ChaosDashboardConfig
}

func NewService(
	archive core.ExperimentStore,
	event core.EventStore,
	conf *config.ChaosDashboardConfig,
) *Service {
	return &Service{
		archive: archive,
		event:   event,
		conf:    conf,
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
	endpoint.PUT("", s.update)
	endpoint.PUT("/pause/:uid", s.pause)
	endpoint.PUT("/start/:uid", s.start)
	endpoint.GET("/state", s.state)
}

// Experiment defines the information of an experiment.
type Experiment struct {
	core.ObjectBase
	Status        status.ChaosStatus `json:"status"`
	FailedMessage string             `json:"failed_message,omitempty"`
}

// Detail adds KubeObjectDesc on Experiment.
type Detail struct {
	Experiment
	KubeObject core.KubeObjectDesc `json:"kube_object"`
}

// @Summary List chaos experiments.
// @Description Get chaos experiments from k8s clusters in real time.
// @Tags experiments
// @Produce json
// @Param namespace query string false "filter exps by namespace"
// @Param name query string false "filter exps by name"
// @Param kind query string false "filter exps by kind" Enums(PodChaos, NetworkChaos, IOChaos, StressChaos, KernelChaos, TimeChaos, DNSChaos, AWSChaos, GCPChaos, JVMChaos, HTTPChaos)
// @Param status query string false "filter exps by status" Enums(Injecting, Running, Finished, Paused)
// @Success 200 {array} Experiment
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments [get]
func (s *Service) list(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	ns, name, kind := c.Query("namespace"), c.Query("name"), c.Query("kind")

	if ns == "" && !s.conf.ClusterScoped && s.conf.TargetNamespace != "" {
		ns = s.conf.TargetNamespace

		log.V(1).Info("Replace query namespace with", ns)
	}

	exps := make([]*Experiment, 0)
	for key, list := range v1alpha1.AllKinds() {
		if kind != "" && key != kind {
			continue
		}

		if err := kubeCli.List(context.Background(), list.ChaosList, &client.ListOptions{Namespace: ns}); err != nil {
			u.SetAPImachineryError(c, err)

			return
		}

		items := reflect.ValueOf(list.ChaosList).Elem().FieldByName("Items")
		for i := 0; i < items.Len(); i++ {
			item := items.Index(i).Addr().Interface().(v1alpha1.InnerObject)
			chaos := item.GetChaos()

			if name != "" && chaos.Name != name {
				continue
			}

			exps = append(exps, &Experiment{
				ObjectBase: core.ObjectBase{
					Namespace: chaos.Namespace,
					Name:      chaos.Name,
					Kind:      chaos.Kind,
					UID:       chaos.UID,
					Created:   chaos.StartTime.Format(time.RFC3339),
				},
				Status: status.GetChaosStatus(item),
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
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments [post]
func (s *Service) create(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	var exp map[string]interface{}
	u.ShouldBindBodyWithJSON(c, &exp)
	kind := exp["kind"].(string)

	if chaosKind, ok := v1alpha1.AllKinds()[kind]; ok {
		reflect.ValueOf(chaosKind.Chaos).Elem().FieldByName("ObjectMeta").Set(reflect.ValueOf(metav1.ObjectMeta{}))
		u.ShouldBindBodyWithJSON(c, chaosKind.Chaos)

		if err = kubeCli.Create(context.Background(), chaosKind.Chaos); err != nil {
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
// @Success 200 {object} Detail
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/{uid} [get]
func (s *Service) get(c *gin.Context) {
	var (
		exp       *core.Experiment
		expDetail *Detail
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
		expDetail = s.findChaosInCluster(c, kubeCli, types.NamespacedName{Namespace: ns, Name: name}, chaosKind.Chaos)

		if expDetail == nil {
			return
		}
	} else {
		u.SetAPIError(c, u.ErrBadRequest.New("Kind "+kind+" is not supported"))

		return
	}

	c.JSON(http.StatusOK, expDetail)
}

func (s *Service) findChaosInCluster(c *gin.Context, kubeCli client.Client, namespacedName types.NamespacedName, chaos runtime.Object) *Detail {
	if err := kubeCli.Get(context.Background(), namespacedName, chaos); err != nil {
		u.SetAPImachineryError(c, err)

		return nil
	}

	getChaosResult := reflect.ValueOf(chaos).MethodByName("GetChaos").Call(nil)[0].Elem()

	return &Detail{
		Experiment: Experiment{
			ObjectBase: core.ObjectBase{
				Namespace: reflect.ValueOf(chaos).Elem().FieldByName("Namespace").String(),
				Name:      reflect.ValueOf(chaos).Elem().FieldByName("Name").String(),
				Kind:      getChaosResult.FieldByName("Kind").String(),
				UID:       getChaosResult.FieldByName("UID").String(),
				Created:   getChaosResult.FieldByName("StartTime").Interface().(time.Time).Format(time.RFC3339),
			},
			Status: status.GetChaosStatus(chaos.(v1alpha1.InnerObject)),
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha1.GroupVersion.String(),
				Kind:       getChaosResult.FieldByName("Kind").String(),
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
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
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

	c.JSON(http.StatusOK, u.Response{Status: "success"})
}

// @Summary Batch delete chaos experiments.
// @Description Batch delete chaos experiments by uids.
// @Tags experiments
// @Produce json
// @Param uids query string true "the experiment uids, split with comma. Example: ?uids=uid1,uid2"
// @Param force query string false "force" Enums(true, false)
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
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

	c.JSON(http.StatusOK, u.Response{Status: "success"})
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

	if err = kubeCli.Get(ctx, namespacedName, chaosKind.Chaos); err != nil {
		u.SetAPImachineryError(c, err)

		return false
	}

	if force == "true" {
		if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return forceClean(kubeCli, chaosKind.Chaos)
		}); err != nil {
			u.SetAPIError(c, u.ErrInternalServer.New("Forced deletion failed"))

			return false
		}
	}

	if err := kubeCli.Delete(ctx, chaosKind.Chaos); err != nil {
		u.SetAPImachineryError(c, err)

		return false
	}

	return true
}

func forceClean(kubeCli client.Client, chaos runtime.Object) error {
	annotations := chaos.(metav1.Object).GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[finalizers.AnnotationCleanFinalizer] = finalizers.AnnotationCleanFinalizerForced
	chaos.(metav1.Object).SetAnnotations(annotations)

	return kubeCli.Update(context.Background(), chaos)
}

// @Summary Update a chaos experiment.
// @Description Update a chaos experiment.
// @Tags experiments
// @Produce json
// @Param request body map[string]interface{} true "Request body"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments [put]
func (s *Service) update(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	var exp map[string]interface{}
	u.ShouldBindBodyWithJSON(c, &exp)
	kind := exp["kind"].(string)

	if chaosKind, ok := v1alpha1.AllKinds()[kind]; ok {
		u.ShouldBindBodyWithJSON(c, chaosKind.Chaos)

		if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return internalUpdate(kubeCli, chaosKind.Chaos)
		}); err != nil {
			u.SetAPImachineryError(c, err)

			return
		}
	} else {
		u.SetAPIError(c, u.ErrBadRequest.New("Kind "+kind+" is not supported"))

		return
	}

	c.JSON(http.StatusOK, exp)
}

func internalUpdate(kubeCli client.Client, chaos runtime.Object) error {
	namespace := reflect.ValueOf(chaos).Elem().FieldByName("Namespace").String()
	name := reflect.ValueOf(chaos).Elem().FieldByName("Name").String()

	if err := kubeCli.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, chaos); err != nil {
		return err
	}

	reflect.ValueOf(chaos).Elem().FieldByName("ObjectMeta").FieldByName("ResourceVersion").SetString(reflect.ValueOf(chaos).Elem().FieldByName("ObjectMeta").FieldByName("ResourceVersion").String())

	return kubeCli.Update(context.Background(), chaos)
}

// @Summary Pause a chaos experiment.
// @Description Pause a chaos experiment.
// @Tags experiments
// @Produce json
// @Param uid path string true "the experiment uid"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
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

	c.JSON(http.StatusOK, u.Response{Status: "success"})
}

// @Summary Start a chaos experiment.
// @Description Start a chaos experiment.
// @Tags experiments
// @Produce json
// @Param uid path string true "the experiment uid"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
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
	chaos := v1alpha1.AllKinds()[exp.Kind].Chaos

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
// @Success 200 {object} utils.AllChaosStatus
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/state [get]
func (s *Service) state(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	ns := c.Query("namespace")
	if ns == "" && !s.conf.ClusterScoped && s.conf.TargetNamespace != "" {
		ns = s.conf.TargetNamespace

		log.V(1).Info("Replace query namespace with", ns)
	}

	allChaosStatus := status.AllChaosStatus{}

	g, ctx := errgroup.WithContext(context.Background())
	m := &sync.Mutex{}
	kinds := v1alpha1.AllKinds()

	var listOptions []client.ListOption
	listOptions = append(listOptions, &client.ListOptions{Namespace: ns})

	for index := range kinds {
		list := kinds[index]

		g.Go(func() error {
			if err := kubeCli.List(ctx, list.ChaosList, listOptions...); err != nil {
				return err
			}
			m.Lock()

			items := reflect.ValueOf(list.ChaosList).Elem().FieldByName("Items")
			for i := 0; i < items.Len(); i++ {
				item := items.Index(i).Addr().Interface().(v1alpha1.InnerObject)
				s := status.GetChaosStatus(item)

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
