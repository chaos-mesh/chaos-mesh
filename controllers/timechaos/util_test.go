package timechaos

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

type SecAndNSecFromDurationTestCase struct {
	Duration time.Duration
	Sec      int64
	NSec     int64
}

func TestSecAndNSecFromDuration(t *testing.T) {
	g := NewGomegaWithT(t)
	cases := []SecAndNSecFromDurationTestCase{
		{time.Second * 100, 100, 0},
		{time.Second * -100, -100, 0},
		{time.Second*-100 + time.Microsecond*-20, -100, -20000},
		{time.Second*-100 + time.Microsecond*20, -99, -999980000},
		{time.Second*100 + time.Microsecond*20, 100, 20000},
		{time.Second*100 + time.Microsecond*-20, 99, 999980000},
	}

	for _, c := range cases {
		sec, nsec := secAndNSecFromDuration(c.Duration)
		g.Expect(sec).Should(Equal(c.Sec))
		g.Expect(nsec).Should(Equal(c.NSec))
	}
}
