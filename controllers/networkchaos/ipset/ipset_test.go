package ipset

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmv1alpha1 "github.com/pingcap/chaos-mesh/api/v1alpha1"
)

func Test_generateIpSetName(t *testing.T) {
	g := NewWithT(t)
	postfix := "alongpostfix"

	t.Run("name with postfix", func(t *testing.T) {
		chaosName := "test"

		networkChaos := &cmv1alpha1.NetworkChaos{
			ObjectMeta: metav1.ObjectMeta{
				Name: chaosName,
			},
		}

		name := generateIpSetName(networkChaos, postfix)

		g.Expect(name).Should(Equal(chaosName + "_" + postfix))
	})

	t.Run("length equal 27", func(t *testing.T) {
		networkChaos := &cmv1alpha1.NetworkChaos{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-metav1object",
			},
		}

		name := generateIpSetName(networkChaos, postfix)

		g.Expect(len(name)).Should(Equal(27))
	})
}
