package config

import (
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
)

type SidecarInjectionSpec struct {
	// RewriteHTTPProbe indicates whether Kubernetes HTTP prober in the PodSpec
	// will be rewritten to be redirected by pilot agent.
	PodRedirectAnnot    map[string]string             `yaml:"podRedirectAnnot"`
	RewriteAppHTTPProbe bool                          `yaml:"rewriteAppHTTPProbe"`
	InitContainers      []corev1.Container            `yaml:"initContainers"`
	Containers          []corev1.Container            `yaml:"containers"`
	Volumes             []corev1.Volume               `yaml:"volumes"`
	DNSConfig           *corev1.PodDNSConfig          `yaml:"dnsConfig"`
	ImagePullSecrets    []corev1.LocalObjectReference `yaml:"imagePullSecrets"`
}

func ReadSidecarConfig(path string) (SidecarInjectionSpec,error) {
	var c SidecarInjectionSpec
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return c, errors.Wrap(err,"Reading Injector Config File")
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return c, errors.Wrap(err,"Unmarshal Injector Config")
	}
	return c,nil
}