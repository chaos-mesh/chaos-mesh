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

package inject

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"

	controllerCfg "github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("webhook inject", func() {

	Context("Inject", func() {
		It("should return unexpected end of JSON input", func() {
			var testClient client.Client
			var cfg *config.Config
			var controllerCfg *controllerCfg.ChaosControllerConfig
			res := Inject(&admissionv1beta1.AdmissionRequest{}, testClient, cfg, controllerCfg, nil)
			Expect(res.Result.Message).To(ContainSubstring("unexpected end of JSON input"))
		})
	})

	Context("checkInjectStatus", func() {
		It("should return false", func() {
			var metadata metav1.ObjectMeta
			metadata.Annotations = make(map[string]string)
			var cfg config.Config
			res := checkInjectStatus(&metadata, &cfg)
			Expect(res).To(Equal(false))
		})

		It("should return true", func() {
			var metadata metav1.ObjectMeta
			metadata.Annotations = make(map[string]string)
			metadata.Annotations["testNamespace/status"] = StatusInjected
			var cfg config.Config
			cfg.AnnotationNamespace = "testNamespace"
			res := checkInjectStatus(&metadata, &cfg)
			Expect(res).To(Equal(true))
		})
	})

	Context("injectByPodRequired", func() {
		It("should return false", func() {
			var metadata metav1.ObjectMeta
			metadata.Annotations = make(map[string]string)
			var cfg config.Config
			str, flag := injectByPodRequired(&metadata, &cfg)
			Expect(str).To(Equal(""))
			Expect(flag).To(Equal(false))
		})

		It("should return false", func() {
			var metadata metav1.ObjectMeta
			metadata.Annotations = make(map[string]string)
			metadata.Annotations["testNamespace/request"] = "test"
			var cfg config.Config
			cfg.AnnotationNamespace = "testNamespace"
			str, flag := injectByPodRequired(&metadata, &cfg)
			Expect(str).To(Equal("test"))
			Expect(flag).To(Equal(true))
		})
	})

	Context("injectRequired", func() {
		It("should return ignore", func() {
			var metadata metav1.ObjectMeta
			metadata.Annotations = make(map[string]string)
			metadata.Namespace = "kube-system"
			var cli client.Client
			var cfg config.Config
			var controllerCfg controllerCfg.ChaosControllerConfig
			str, flag := injectRequired(&metadata, cli, &cfg, &controllerCfg)
			Expect(str).To(Equal(""))
			Expect(flag).To(Equal(false))
		})

		It("should return ignore", func() {
			var metadata metav1.ObjectMeta
			metadata.Annotations = make(map[string]string)
			metadata.Annotations["testNamespace/status"] = StatusInjected
			var cfg config.Config
			var controllerCfg controllerCfg.ChaosControllerConfig
			cfg.AnnotationNamespace = "testNamespace"
			var cli client.Client
			str, flag := injectRequired(&metadata, cli, &cfg, &controllerCfg)
			Expect(str).To(Equal(""))
			Expect(flag).To(Equal(false))
		})

		It("should return ignore", func() {
			var metadata metav1.ObjectMeta
			metadata.Annotations = make(map[string]string)
			metadata.Annotations["testNamespace/status"] = StatusInjected
			var cfg config.Config
			var controllerCfg controllerCfg.ChaosControllerConfig
			cfg.AnnotationNamespace = "testNamespace"
			var cli client.Client
			str, flag := injectRequired(&metadata, cli, &cfg, &controllerCfg)
			Expect(str).To(Equal(""))
			Expect(flag).To(Equal(false))
		})

		It("should return Pod annotation requesting sidecar config", func() {
			var metadata metav1.ObjectMeta
			metadata.Annotations = make(map[string]string)
			metadata.Annotations["testNamespace/request"] = "test"
			metadata.Namespace = "testNamespace"
			var cfg config.Config
			var controllerCfg controllerCfg.ChaosControllerConfig
			cfg.AnnotationNamespace = "testNamespace"
			str, flag := injectRequired(&metadata, k8sClient, &cfg, &controllerCfg)
			Expect(str).To(Equal("test"))
			Expect(flag).To(Equal(true))
		})

		It("should return false", func() {
			var metadata metav1.ObjectMeta
			metadata.Annotations = make(map[string]string)
			var cfg config.Config
			var controllerCfg controllerCfg.ChaosControllerConfig
			_, flag := injectRequired(&metadata, k8sClient, &cfg, &controllerCfg)
			Expect(flag).To(Equal(false))
		})
	})

	Context("injectByNamespaceRequired", func() {
		It("should return nil and false", func() {
			var metadata metav1.ObjectMeta
			metadata.Annotations = make(map[string]string)
			metadata.Namespace = "testNamespace"
			var cfg config.Config
			str, flag := injectByNamespaceRequired(&metadata, k8sClient, &cfg)
			Expect(str).To(Equal(""))
			Expect(flag).To(Equal(false))
		})
	})

	Context("createPatch", func() {
		It("should return nil and false", func() {
			var pod corev1.Pod
			var inj config.InjectionConfig
			annotations := make(map[string]string)
			_, err := createPatch(&pod, &inj, annotations)
			Expect(err).To(BeNil())
		})
	})

	Context("setCommands", func() {
		It("should return", func() {
			var target []corev1.Container = []corev1.Container{
				{
					Name: "testContainerName",
				}}
			postStart := make(map[string]config.ExecAction)
			patch := setCommands(target, postStart)
			Expect(patch).To(BeNil())
		})

		It("should return nil", func() {
			var target []corev1.Container = []corev1.Container{
				{
					Name: "testContainerName",
				}}
			postStart := make(map[string]config.ExecAction)
			var ce config.ExecAction = config.ExecAction{
				Command: []string{"nil"},
			}
			postStart["testContainerName"] = ce
			patch := setCommands(target, postStart)
			Expect(patch).ToNot(BeNil())
		})
	})

	Context("setEnvironment", func() {
		It("should return not nil", func() {
			var target []corev1.Container = []corev1.Container{
				{
					Name: "testContainerName",
				}}
			var addEnv []corev1.EnvVar = []corev1.EnvVar{
				{
					Name: "testContainerName",
				}}
			patch := setEnvironment(target, addEnv)
			Expect(patch).ToNot(BeNil())
		})

		It("should return not nil", func() {
			var Env []corev1.EnvVar = []corev1.EnvVar{
				{
					Name: "_testContainerName_",
				}}
			var target []corev1.Container = []corev1.Container{
				{
					Name: "testContainerName",
					Env:  Env,
				}}
			var addEnv []corev1.EnvVar = []corev1.EnvVar{
				{
					Name: "testContainerName",
				}}
			patch := setEnvironment(target, addEnv)
			Expect(patch).ToNot(BeNil())
		})

		It("should return nil", func() {
			var Env []corev1.EnvVar = []corev1.EnvVar{
				{
					Name: "testContainerName",
				}}
			var target []corev1.Container = []corev1.Container{
				{
					Name: "testContainerName",
					Env:  Env,
				}}
			var addEnv []corev1.EnvVar = []corev1.EnvVar{
				{
					Name: "testContainerName",
				}}
			patch := setEnvironment(target, addEnv)
			Expect(patch).To(BeNil())
		})
	})

	Context("addContainers", func() {
		It("should return not nil", func() {
			var target []corev1.Container = []corev1.Container{
				{
					Name: "testContainerName",
				}}
			var added []corev1.Container = []corev1.Container{
				{
					Name: "testContainerName",
				}}
			basePath := "/test"
			patch := addContainers(target, added, basePath)
			Expect(patch).ToNot(BeNil())
		})

		It("should return not nil", func() {
			var target []corev1.Container = []corev1.Container{}
			var added []corev1.Container = []corev1.Container{
				{
					Name: "testContainerName",
				}}
			basePath := "/test"
			patch := addContainers(target, added, basePath)
			Expect(patch).ToNot(BeNil())
		})
	})

	Context("addVolumes", func() {
		It("should return not nil", func() {
			var target []corev1.Volume = []corev1.Volume{
				{
					Name: "test",
				}}
			var added []corev1.Volume = []corev1.Volume{
				{
					Name: "test",
				}}
			basePath := "/test"
			patch := addVolumes(target, added, basePath)
			Expect(patch).ToNot(BeNil())
		})

		It("should return not nil", func() {
			var target []corev1.Volume = []corev1.Volume{}
			var added []corev1.Volume = []corev1.Volume{
				{
					Name: "test",
				}}
			basePath := "/test"
			patch := addVolumes(target, added, basePath)
			Expect(patch).ToNot(BeNil())
		})
	})

	Context("setVolumeMounts", func() {
		It("should return not nil", func() {
			var vm []corev1.VolumeMount = []corev1.VolumeMount{
				{
					Name: "test",
				}}
			var target []corev1.Container = []corev1.Container{
				{
					Name:         "test",
					VolumeMounts: vm,
				}}
			var added []corev1.VolumeMount = []corev1.VolumeMount{
				{
					Name: "test",
				}}
			basePath := "/test"
			patch := setVolumeMounts(target, added, basePath)
			Expect(patch).ToNot(BeNil())
		})
	})

	Context("addHostAliases", func() {
		It("should return not nil", func() {
			var target []corev1.HostAlias = []corev1.HostAlias{
				{
					IP: "testip",
				}}
			var added []corev1.HostAlias = []corev1.HostAlias{
				{
					IP: "testip",
				}}
			basePath := "/test"
			patch := addHostAliases(target, added, basePath)
			Expect(patch).ToNot(BeNil())
		})

		It("should return not nil", func() {
			var target []corev1.HostAlias = []corev1.HostAlias{}
			var added []corev1.HostAlias = []corev1.HostAlias{
				{
					IP: "testip",
				}}
			basePath := "/test"
			patch := addHostAliases(target, added, basePath)
			Expect(patch).ToNot(BeNil())
		})
	})

	Context("mergeEnvVars", func() {
		It("should return not nil", func() {
			var envs []corev1.EnvVar = []corev1.EnvVar{
				{
					Name: "test",
				}}
			var containers []corev1.Container = []corev1.Container{
				{
					Name: "test",
				}}
			mutatedContainers := mergeEnvVars(envs, containers)
			Expect(mutatedContainers).ToNot(BeNil())
		})

		It("should return not nil", func() {
			var envs []corev1.EnvVar = []corev1.EnvVar{
				{
					Name: "test",
				}}
			var env []corev1.EnvVar = []corev1.EnvVar{
				{
					Name: "test",
				}}
			var containers []corev1.Container = []corev1.Container{
				{
					Name: "test",
					Env:  env,
				}}
			mutatedContainers := mergeEnvVars(envs, containers)
			Expect(mutatedContainers).ToNot(BeNil())
		})
	})

	Context("mergeVolumeMounts", func() {
		It("should return not nil", func() {
			var volumeMounts []corev1.VolumeMount = []corev1.VolumeMount{
				{
					Name: "test",
				}}
			var containers []corev1.Container = []corev1.Container{
				{
					Name: "test",
				}}
			mutatedContainers := mergeVolumeMounts(volumeMounts, containers)
			Expect(mutatedContainers).ToNot(BeNil())
		})

		It("should return not nil", func() {
			var volumeMounts []corev1.VolumeMount = []corev1.VolumeMount{
				{
					Name: "test",
				}}
			var vm []corev1.VolumeMount = []corev1.VolumeMount{
				{
					Name: "test",
				}}
			var containers []corev1.Container = []corev1.Container{
				{
					Name:         "test",
					VolumeMounts: vm,
				}}
			mutatedContainers := mergeVolumeMounts(volumeMounts, containers)
			Expect(mutatedContainers).ToNot(BeNil())
		})
	})

	Context("updateAnnotations", func() {
		It("should return not nil", func() {
			target := make(map[string]string)
			added := make(map[string]string)
			added["testKey"] = "testValue"
			patch := updateAnnotations(target, added)
			Expect(patch).ToNot(BeNil())
		})

		It("should return not nil", func() {
			target := make(map[string]string)
			added := make(map[string]string)
			added["testKey"] = "testValue"
			target["testKey"] = "testValue"
			patch := updateAnnotations(target, added)
			Expect(patch).ToNot(BeNil())
		})
	})

	Context("potentialPodName", func() {
		It("should return testName", func() {
			var metadata metav1.ObjectMeta
			metadata.Name = "testName"
			name := potentialPodName(&metadata)
			Expect(name).ToNot(BeNil())
			Expect(name).To(Equal("testName"))
		})

		It("should return (actual name not yet known)", func() {
			var metadata metav1.ObjectMeta
			metadata.GenerateName = "testName"
			name := potentialPodName(&metadata)
			Expect(name).ToNot(BeNil())
			Expect(name).To(ContainSubstring("(actual name not yet known)"))
		})

		It("should return nil", func() {
			var metadata metav1.ObjectMeta
			name := potentialPodName(&metadata)
			Expect(name).To(Equal(""))
		})
	})
})
