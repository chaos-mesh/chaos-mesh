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
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/finalizers"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
)

var log = utils.Log.WithName("experiments")

// Service defines a handler service for experiments.
type Service struct {
	archive core.ExperimentStore
	event   core.EventStore
	conf    *config.ChaosDashboardConfig
	scheme  *runtime.Scheme
}

// NewService returns an experiment service instance.
func NewService(
	archive core.ExperimentStore,
	event core.EventStore,
	conf *config.ChaosDashboardConfig,
	scheme *runtime.Scheme,
) *Service {
	return &Service{
		archive: archive,
		event:   event,
		conf:    conf,
		scheme:  scheme,
	}
}

// Register mounts HTTP handler on the mux.
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

type allChaosStatus struct {
	Injecting int `json:"injecting"`
	Running   int `json:"running"`
	Finished  int `json:"finished"`
	Paused    int `json:"paused"`
}

// Experiment defines the information of an experiment.
type Experiment struct {
	core.ObjectBase `json:",inline"`
	Status          utils.ChaosStatus `json:"status"`
	FailedMessage   string            `json:"failed_message,omitempty"`
}

// Detail represents an experiment instance.
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
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	ns := c.Query("namespace")
	name := c.Query("name")
	kind := c.Query("kind")

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
			utils.SetApimachineryError(c, err)

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
				Status: utils.GetChaosStatus(item),
			})
		}
	}

	c.JSON(http.StatusOK, exps)
}

// @Summary Create a new chaos experiment.
// @Description Pass a JSON object to create a new chaos experiment. The schema for JSON is the same as the YAML schema for the Kubernetes object.
// @Tags experiments
// @Accept json
// @Produce json
// @Param chaos body map[string]interface{} true "the chaos definition"
// @Success 200 {object} core.ExperimentInfo
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments [post]
func (s *Service) create(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	var exp map[string]interface{}
	if err = c.ShouldBindBodyWith(&exp, binding.JSON); err != nil {
		c.Status(http.StatusInternalServerError)
		c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))

		return
	}

	switch kind := exp["kind"].(string); kind {
	case v1alpha1.KindPodChaos:
		var chaos v1alpha1.PodChaos
		utils.ShouldBindBodyWithJSON(c, &chaos)
		err = kubeCli.Create(context.Background(), &chaos)
	case v1alpha1.KindNetworkChaos:
		var chaos v1alpha1.NetworkChaos
		utils.ShouldBindBodyWithJSON(c, &chaos)
		err = kubeCli.Create(context.Background(), &chaos)
	case v1alpha1.KindIOChaos:
		var chaos v1alpha1.IOChaos
		utils.ShouldBindBodyWithJSON(c, &chaos)
		err = kubeCli.Create(context.Background(), &chaos)
	case v1alpha1.KindStressChaos:
		var chaos v1alpha1.StressChaos
		utils.ShouldBindBodyWithJSON(c, &chaos)
		err = kubeCli.Create(context.Background(), &chaos)
	case v1alpha1.KindKernelChaos:
		var chaos v1alpha1.KernelChaos
		utils.ShouldBindBodyWithJSON(c, &chaos)
		err = kubeCli.Create(context.Background(), &chaos)
	case v1alpha1.KindTimeChaos:
		var chaos v1alpha1.TimeChaos
		utils.ShouldBindBodyWithJSON(c, &chaos)
		err = kubeCli.Create(context.Background(), &chaos)
	case v1alpha1.KindDNSChaos:
		var chaos v1alpha1.DNSChaos
		utils.ShouldBindBodyWithJSON(c, &chaos)
		err = kubeCli.Create(context.Background(), &chaos)
	case v1alpha1.KindAWSChaos:
		var chaos v1alpha1.AWSChaos
		utils.ShouldBindBodyWithJSON(c, &chaos)
		err = kubeCli.Create(context.Background(), &chaos)
	case v1alpha1.KindGCPChaos:
		var chaos v1alpha1.GCPChaos
		utils.ShouldBindBodyWithJSON(c, &chaos)
		err = kubeCli.Create(context.Background(), &chaos)
	case v1alpha1.KindJVMChaos:
		var chaos v1alpha1.JVMChaos
		utils.ShouldBindBodyWithJSON(c, &chaos)
		err = kubeCli.Create(context.Background(), &chaos)
	case v1alpha1.KindHTTPChaos:
		var chaos v1alpha1.HTTPChaos
		utils.ShouldBindBodyWithJSON(c, &chaos)
		err = kubeCli.Create(context.Background(), &chaos)
	default:
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrBadRequest.New("Kind " + kind + " is not supported"))

		return
	}

	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))

		return
	}

	c.JSON(http.StatusOK, exp)
}

func (s *Service) findChaosInCluster(c *gin.Context, kubeCli client.Client, namespace string, name string, chaos runtime.Object) *Detail {
	if err := kubeCli.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, chaos); err != nil {
		utils.SetApimachineryError(c, err)

		return nil
	}

	gvk, err := apiutil.GVKForObject(chaos, s.scheme)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))

		return nil
	}

	return &Detail{
		Experiment: Experiment{
			ObjectBase: core.ObjectBase{
				Namespace: reflect.ValueOf(chaos).Elem().FieldByName("Namespace").String(),
				Name:      reflect.ValueOf(chaos).Elem().FieldByName("Name").String(),
				Kind:      gvk.Kind,
				UID:       reflect.ValueOf(chaos).MethodByName("GetChaos").Call(nil)[0].Elem().FieldByName("UID").String(),
				Created:   reflect.ValueOf(chaos).MethodByName("GetChaos").Call(nil)[0].Elem().FieldByName("StartTime").Interface().(time.Time).Format(time.RFC3339),
			},
			Status: utils.GetChaosStatus(chaos.(v1alpha1.InnerObject)),
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       gvk.Kind,
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
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusNotFound)
			c.Error(utils.ErrNotFound.New("Experiment " + uid + " not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	ns, name, kind := exp.Namespace, exp.Name, exp.Kind

	switch kind {
	case v1alpha1.KindPodChaos:
		expDetail = s.findChaosInCluster(c, kubeCli, ns, name, &v1alpha1.PodChaos{})
	case v1alpha1.KindNetworkChaos:
		expDetail = s.findChaosInCluster(c, kubeCli, ns, name, &v1alpha1.NetworkChaos{})
	case v1alpha1.KindIOChaos:
		expDetail = s.findChaosInCluster(c, kubeCli, ns, name, &v1alpha1.IOChaos{})
	case v1alpha1.KindStressChaos:
		expDetail = s.findChaosInCluster(c, kubeCli, ns, name, &v1alpha1.StressChaos{})
	case v1alpha1.KindKernelChaos:
		expDetail = s.findChaosInCluster(c, kubeCli, ns, name, &v1alpha1.KernelChaos{})
	case v1alpha1.KindTimeChaos:
		expDetail = s.findChaosInCluster(c, kubeCli, ns, name, &v1alpha1.TimeChaos{})
	case v1alpha1.KindDNSChaos:
		expDetail = s.findChaosInCluster(c, kubeCli, ns, name, &v1alpha1.DNSChaos{})
	case v1alpha1.KindAWSChaos:
		expDetail = s.findChaosInCluster(c, kubeCli, ns, name, &v1alpha1.AWSChaos{})
	case v1alpha1.KindGCPChaos:
		expDetail = s.findChaosInCluster(c, kubeCli, ns, name, &v1alpha1.GCPChaos{})
	case v1alpha1.KindJVMChaos:
		expDetail = s.findChaosInCluster(c, kubeCli, ns, name, &v1alpha1.JVMChaos{})
	case v1alpha1.KindHTTPChaos:
		expDetail = s.findChaosInCluster(c, kubeCli, ns, name, &v1alpha1.HTTPChaos{})
	default:
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrBadRequest.New("Kind " + kind + " is not supported"))

		return
	}

	if expDetail == nil {
		return
	}

	c.JSON(http.StatusOK, expDetail)
}

func checkAndDeleteChaos(c *gin.Context, kubeCli client.Client, namespace string, name string, kind string, force string) bool {
	var (
		chaosKind *v1alpha1.ChaosKind
		ok        bool
		err       error
	)

	if chaosKind, ok = v1alpha1.AllKinds()[kind]; !ok {
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrBadRequest.New("Kind " + kind + " is not supported"))

		return false
	}

	ctx := context.Background()

	if err = kubeCli.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, chaosKind.Chaos); err != nil {
		utils.SetApimachineryError(c, err)

		return false
	}

	if force == "true" {
		if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return forceClean(kubeCli, namespace, name, kind)
		}); err != nil {
			c.Status(http.StatusInternalServerError)
			c.Error(utils.ErrInternalServer.New("Forced deletion failed because the setAnnotations of chaos could not be updated"))

			return false
		}
	}

	if err := kubeCli.Delete(ctx, chaosKind.Chaos, &client.DeleteOptions{}); err != nil {
		utils.SetApimachineryError(c, err)

		return false
	}

	return true
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
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusNotFound)
			c.Error(utils.ErrNotFound.New("Experiment " + uid + " not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	ns, name, kind, force := exp.Namespace, exp.Name, exp.Kind, c.DefaultQuery("force", "false")
	if ok := checkAndDeleteChaos(c, kubeCli, ns, name, kind, force); !ok {
		return
	}

	c.JSON(http.StatusOK, utils.Response{Status: "success"})
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
		exp      *core.Experiment
		uidSlice []string
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uids := c.Query("uids")
	if uids == "" {
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrInternalServer.New("The uids cannot be empty"))

		return
	}

	uidSlice = strings.Split(uids, ",")
	force := c.DefaultQuery("force", "false")

	if len(uidSlice) > 100 {
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrInternalServer.New("Too many uids, please delete less than 100 at a time"))

		return
	}

	for _, uid := range uidSlice {
		if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
			if gorm.IsRecordNotFoundError(err) {
				c.Status(http.StatusNotFound)
				c.Error(utils.ErrNotFound.New("Experiment " + uid + " not found"))
			} else {
				c.Status(http.StatusInternalServerError)
				c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			}

			return
		}

		ns, name, kind := exp.Namespace, exp.Name, exp.Kind
		if ok := checkAndDeleteChaos(c, kubeCli, ns, name, kind, force); !ok {
			return
		}
	}

	c.JSON(http.StatusOK, utils.Response{Status: "success"})
}

func (s *Service) patchExperiment(c *gin.Context, kubeCli client.Client, exp *core.Experiment, annotations map[string]string) error {
	var chaos runtime.Object

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
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusNotFound)
			c.Error(utils.ErrNotFound.New("Experiment " + uid + " not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	annotations := map[string]string{
		v1alpha1.PauseAnnotationKey: "true",
	}
	if err := s.patchExperiment(c, kubeCli, exp, annotations); err != nil {
		utils.SetApimachineryError(c, err)

		return
	}

	c.JSON(http.StatusOK, utils.Response{Status: "success"})
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
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusNotFound)
			c.Error(utils.ErrNotFound.New("Experiment " + uid + " not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	annotations := map[string]string{
		v1alpha1.PauseAnnotationKey: "false",
	}
	if err := s.patchExperiment(c, kubeCli, exp, annotations); err != nil {
		utils.SetApimachineryError(c, err)

		return
	}

	c.JSON(http.StatusOK, utils.Response{Status: "success"})
}

// @Summary Get the status of all experiments.
// @Description Get the status of all experiments.
// @Tags experiments
// @Produce json
// @Param namespace query string false "namespace"
// @Success 200 {object} allChaosStatus
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/state [get]
func (s *Service) state(c *gin.Context) {
	var (
		err error
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	ns := c.Query("namespace")
	if ns == "" && !s.conf.ClusterScoped && s.conf.TargetNamespace != "" {
		ns = s.conf.TargetNamespace

		log.V(1).Info("Replace query namespace with", ns)
	}

	status := allChaosStatus{}

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
				state := utils.GetChaosStatus(item)
				if err != nil {
					c.Status(http.StatusInternalServerError)
					_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
					return err
				}
				switch state {
				case utils.Paused:
					status.Paused++
				case utils.Running:
					status.Running++
				case utils.Injecting:
					status.Injecting++
				case utils.Finished:
					status.Finished++
				}
			}

			m.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		c.Status(http.StatusInternalServerError)
		utils.SetApimachineryError(c, err)

		return
	}

	c.JSON(http.StatusOK, status)
}

// @Summary Update a chaos experiment.
// @Description Update a chaos experiment.
// @Tags experiments
// @Produce json
// @Param request body map[string]interface{} true "Request body"
// @Success 200 {object} core.KubeObjectDesc
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/update [put]
func (s *Service) update(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(utils.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	var exp map[string]interface{}
	if err = c.ShouldBindJSON(&exp); err != nil {
		c.Status(http.StatusInternalServerError)
		c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))

		return
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		switch kind := exp["kind"].(string); kind {
		case v1alpha1.KindPodChaos:
			var chaos v1alpha1.PodChaos
			utils.ShouldBindBodyWithJSON(c, &chaos)
			err = internalUpdate(c, kubeCli, &v1alpha1.PodChaos{}, &chaos)
		case v1alpha1.KindNetworkChaos:
			var chaos v1alpha1.NetworkChaos
			utils.ShouldBindBodyWithJSON(c, &chaos)
			err = internalUpdate(c, kubeCli, &v1alpha1.NetworkChaos{}, &chaos)
			// case v1alpha1.KindIOChaos:
			// 	var chaos v1alpha1.IOChaos
			// 	internalUpdate(c, kubeCli, &v1alpha1.IOChaos{}, &chaos)
			// case v1alpha1.KindStressChaos:
			// 	var chaos v1alpha1.StressChaos
			// 	internalUpdate(c, kubeCli, &v1alpha1.StressChaos{}, &chaos)
			// case v1alpha1.KindKernelChaos:
			// 	var chaos v1alpha1.KernelChaos
			// 	internalUpdate(c, kubeCli, &v1alpha1.KernelChaos{}, &chaos)
			// case v1alpha1.KindTimeChaos:
			// 	var chaos v1alpha1.TimeChaos
			// 	internalUpdate(c, kubeCli, &v1alpha1.TimeChaos{}, &chaos)
			// case v1alpha1.KindDNSChaos:
			// 	var chaos v1alpha1.DNSChaos
			// 	internalUpdate(c, kubeCli, &chaos)
			// case v1alpha1.KindAwsChaos:
			// 	var chaos v1alpha1.AwsChaos
			// 	internalUpdate(c, kubeCli, &chaos)
			// case v1alpha1.KindGcpChaos:
			// 	var chaos v1alpha1.GcpChaos
			// 	internalUpdate(c, kubeCli, &chaos)
			// case v1alpha1.KindJVMChaos:
			// 	var chaos v1alpha1.JVMChaos
			// 	internalUpdate(c, kubeCli, &chaos)
			// case v1alpha1.KindHTTPChaos:
			// 	var chaos v1alpha1.HTTPChaos
			// 	internalUpdate(c, kubeCli, &chaos)
		}

		return err
	})

	if err != nil {
		utils.SetApimachineryError(c, err)

		return
	}

	c.JSON(http.StatusOK, exp)
}

func internalUpdate(c *gin.Context, kubeCli client.Client, current runtime.Object, chaos runtime.Object) error {
	namespace := reflect.ValueOf(chaos).Elem().FieldByName("Namespace").String()
	name := reflect.ValueOf(chaos).Elem().FieldByName("Name").String()

	if err := kubeCli.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, current); err != nil {
		return err
	}

	reflect.ValueOf(chaos).Elem().FieldByName("ObjectMeta").FieldByName("ResourceVersion").SetString(reflect.ValueOf(current).Elem().FieldByName("ObjectMeta").FieldByName("ResourceVersion").String())
	fmt.Printf("%#v\n", chaos)

	return kubeCli.Update(context.Background(), chaos)
}

func forceClean(kubeCli client.Client, ns string, name string, kind string) error {
	var chaos runtime.Object

	if err := kubeCli.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: name}, chaos); err != nil {
		return err
	}

	annotations := chaos.(metav1.Object).GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[finalizers.AnnotationCleanFinalizer] = finalizers.AnnotationCleanFinalizerForced
	chaos.(metav1.Object).SetAnnotations(annotations)

	return kubeCli.Update(context.Background(), chaos)
}
