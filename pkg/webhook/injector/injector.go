package injector

import (
	"encoding/json"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/injector/config"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path"
	ctrl "sigs.k8s.io/controller-runtime"
)
var log = ctrl.Log.WithName("inject-webhook")

var configDirPath = "/usr/local/bin/config/webhook/"

func Inject(res *v1beta1.AdmissionRequest) *v1beta1.AdmissionResponse {
	var pod corev1.Pod
	if err := json.Unmarshal(res.Object.Raw, &pod); err != nil {
		log.Error(err, "Could not unmarshal raw object")
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	podName := potentialPodName(&pod.ObjectMeta)
	if pod.ObjectMeta.Namespace == "" {
		pod.ObjectMeta.Namespace = res.Namespace
	}

	log.Info("AdmissionReview for",
		"Kind", res.Kind, "Namespace", res.Namespace, "Name", res.Name, "podName", podName, "UID", res.UID, "patchOperation", res.Operation, "UserInfo", res.UserInfo)
	log.V(4).Info("Object", "Object", string(res.Object.Raw))
	log.V(4).Info("OldObject", "OldObject", string(res.OldObject.Raw))
	log.V(4).Info("Pod", "Pod", pod)

	cfg, err := config.ReadInjectConfig(path.Join(configDirPath, "injector.yaml"))
	if err != nil {
		log.Info("Skipping injection due to Read Config fail",
			err,"namespace", pod.ObjectMeta.Namespace, "name", podName)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	ok, err := PodFilter.CheckAll(PodItem{Pod: &pod, Config: &cfg})
	if !ok||err!=nil {
		log.Info("Skipping injection due to policy checkALL",err, "namespace", pod.ObjectMeta.Namespace, "name", podName)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	spec,err := config.ReadSidecarConfig(path.Join(configDirPath,"sidecar.yaml"))
	if err != nil {
		log.Info("Skipping injection due to Read sidecar Config fail",
			err,"namespace", pod.ObjectMeta.Namespace, "name", podName)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	patchBytes, err := createPatch(&pod,&spec)
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	log.Info("AdmissionResponse: patch", "patchBytes", string(patchBytes))
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func potentialPodName(metadata *metav1.ObjectMeta) string {
	if metadata.Name != "" {
		return metadata.Name
	}
	if metadata.GenerateName != "" {
		return metadata.GenerateName + "***** (actual name not yet known)"
	}
	return ""
}