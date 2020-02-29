package utils

import (
	. "github.com/onsi/gomega"

	"testing"
)

func TestEncodeClkIds(t *testing.T) {
	g := NewGomegaWithT(t)

	mask, err := EncodeClkIds([]string{"CLOCK_REALTIME"})
	g.Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)
	g.Expect(mask).Should(Equal(uint64(1)))

	mask, err = EncodeClkIds([]string{"CLOCK_REALTIME", "CLOCK_MONOTONIC"})
	g.Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)
	g.Expect(mask).Should(Equal(uint64(3)))

	mask, err = EncodeClkIds([]string{"CLOCK_MONOTONIC"})
	g.Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)
	g.Expect(mask).Should(Equal(uint64(2)))
}
