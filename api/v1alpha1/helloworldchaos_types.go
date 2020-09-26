package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KindHelloWorldChaos is the kind for hello-world chaos
/*const KindHelloWorldChaos = "HelloWorldChaos"

func init() {
	all.register(KindHelloWorldChaos, &ChaosKind{
		Chaos:     &HelloWorldChaos{},
		ChaosList: &HelloWorldChaosList{},
	})
}*/

// +kubebuilder:object:root=true

// HelloWorldChaos is the Schema for the helloworldchaos API
type HelloWorldChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// +kubebuilder:object:root=true

// HelloWorldChaosList contains a list of HelloWorldChaos
type HelloWorldChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelloWorldChaos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelloWorldChaos{}, &HelloWorldChaosList{})
}
