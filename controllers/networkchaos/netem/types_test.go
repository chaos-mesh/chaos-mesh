package netem

import (
	"testing"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
)

func TestMergenetem(t *testing.T) {
	spec := v1alpha1.NetworkChaosSpec{
		Action: "netem",
	}
	_, err := mergeNetem(spec)
	if err == nil {
		t.Errorf("expect invalid spec failed with message %s but got nil", invalidNetemSpecMsg)
	}
	if err != nil && err.Error() != invalidNetemSpecMsg {
		t.Errorf("expect merge failed with message %s but got %v", invalidNetemSpecMsg, err)
	}
}
