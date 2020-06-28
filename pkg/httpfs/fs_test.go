package httpfs

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	vfs "golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

func ExampleFileServer() {
	http.Handle("/assets/", http.StripPrefix("/assets", FileServer(
		http.Dir("assets"),
		FileServerOptions{
			IndexHTML: true,
		},
	)))
}

func TestFileServerImplicitLeadingSlash(t *testing.T) {
	g := NewGomegaWithT(t)

	fs := vfs.New(mapfs.New(map[string]string{
		"foo.txt": "Hello world",
	}))

	ts := httptest.NewServer(http.StripPrefix("/bar/", FileServer(fs, FileServerOptions{})))
	defer ts.Close()
	get := func(suffix string) string {
		res, err := http.Get(ts.URL + suffix)
		if err != nil {
			t.Fatalf("Get %s: %v", suffix, err)
		}
		defer res.Body.Close()
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("ReadAll %s: %v", suffix, err)
		}
		return string(b)
	}

	g.Expect(strings.Contains(get("/bar/"), ">foot.txt<")).Should(Equal(false))
	g.Expect(get("/bar/foo.txt")).Should(Equal("Hello world"))
}
