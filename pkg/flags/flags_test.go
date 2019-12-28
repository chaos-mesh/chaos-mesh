package flags

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestNewMapStringStringFlag(t *testing.T) {
	g := NewGomegaWithT(t)
	flag := NewMapStringStringFlag()
	g.Expect(flag.Values).ShouldNot(BeNil())
}

func TestMapStringStringFlag_String(t *testing.T) {

	g := NewGomegaWithT(t)
	flag := NewMapStringStringFlag()

	g.Expect(flag.String()).Should(BeEmpty())

	var err error

	err = flag.Set("flag1")
	g.Expect(err).ShouldNot(BeNil())
	g.Expect(flag.String()).Should(BeEmpty())

	//err = flag.Set("flag1=")
	//err = flag.Set("=")

	err = flag.Set("")
	g.Expect(err).ShouldNot(BeNil())
	g.Expect(flag.String()).Should(BeEmpty())

	err = flag.Set("flag2=key2")
	g.Expect(err).Should(BeNil())
	g.Expect(flag.String()).Should(Equal("flag2=key2"))

	//err = flag.Set("    flag3=key3     ")
	err = flag.Set("flag2=key2")
	g.Expect(err).Should(BeNil())
	g.Expect(flag.String()).Should(Equal("flag2=key2"))

	err = flag.Set("flag3=key3,flag4=key4,flag2=key22")

	g.Expect(err).Should(BeNil())
	g.Expect(strings.Contains(flag.String(), "flag2=key22")).To(Equal(true))
	g.Expect(strings.Contains(flag.String(), "flag4=key4")).To(Equal(true))
	g.Expect(strings.Contains(flag.String(), "flag3=key3")).To(Equal(true))
	g.Expect(strings.Contains(flag.String(), ",")).To(Equal(true))

	g.Expect(len(flag.String())).To(Equal(len("flag3=key3,flag4=key4,flag2=key22")))
}
