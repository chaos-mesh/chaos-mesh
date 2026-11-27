// Copyright 2022 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package helm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	chartutil "helm.sh/helm/v4/pkg/chart/v2/util"
)

func TestHelmChart(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helm chart tests")
}

type chartTransport func(*http.Request) (*http.Response, error)

func (transport chartTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

type chartBody struct {
	io.Reader
	closed bool
}

func (body *chartBody) Close() error {
	body.closed = true
	return nil
}

var _ = Describe("Downloading a Chaos Mesh chart", func() {
	const version = "2.2.0"
	var directory string

	BeforeEach(func() {
		directory = GinkgoT().TempDir()
		GinkgoT().Setenv("TMPDIR", directory)
	})

	It("requests the expected chart and saves the response", func(ctx SpecContext) {
		body := &chartBody{Reader: strings.NewReader("chart archive")}
		client := &http.Client{Transport: chartTransport(func(request *http.Request) (*http.Response, error) {
			Expect(request.Method).To(Equal(http.MethodGet))
			Expect(request.URL.String()).To(Equal(ChaosMeshHelmRepo + "/chaos-mesh-2.2.0.tgz"))
			Expect(request.Context()).To(BeIdenticalTo(ctx))
			return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
		})}

		path, err := downloadChaosMeshChartTgz(ctx, version, client)
		Expect(err).NotTo(HaveOccurred())
		Expect(filepath.Dir(path)).To(Equal(directory))
		Expect(filepath.Base(path)).To(And(HavePrefix("chaos-mesh-2.2.0-"), HaveSuffix(".tgz")))
		Expect(os.ReadFile(path)).To(Equal([]byte("chart archive")))
		Expect(body.closed).To(BeTrue())
	})

	It("reports a non-success HTTP status without saving the error body", func() {
		body := &chartBody{Reader: strings.NewReader("chart not found")}
		client := &http.Client{Transport: chartTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusNotFound, Body: body}, nil
		})}

		path, err := downloadChaosMeshChartTgz(context.Background(), version, client)
		Expect(path).To(BeEmpty())
		Expect(err).To(MatchError(ContainSubstring("HTTP status 404")))
		Expect(os.ReadDir(directory)).To(BeEmpty())
		Expect(body.closed).To(BeTrue())
	})

	It("propagates a transport error", func() {
		unavailable := errors.New("transport unavailable")
		client := &http.Client{Transport: chartTransport(func(*http.Request) (*http.Response, error) {
			return nil, unavailable
		})}
		path, err := downloadChaosMeshChartTgz(context.Background(), version, client)
		Expect(path).To(BeEmpty())
		Expect(err).To(MatchError(unavailable))
		Expect(os.ReadDir(directory)).To(BeEmpty())
	})

	It("removes an incomplete archive and closes the response when downloading fails", func() {
		disconnected := errors.New("body interrupted")
		body := &chartBody{Reader: io.MultiReader(strings.NewReader("incomplete"), iotest.ErrReader(disconnected))}
		client := &http.Client{Transport: chartTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
		})}

		path, err := downloadChaosMeshChartTgz(context.Background(), version, client)
		Expect(path).To(BeEmpty())
		Expect(err).To(MatchError(disconnected))
		Expect(os.ReadDir(directory)).To(BeEmpty())
		Expect(body.closed).To(BeTrue())
	})
})

var _ = Describe("Loading a local Chaos Mesh chart", func() {
	const version = "2.2.0"

	It("reads the requested chart's metadata", func() {
		directory := GinkgoT().TempDir()
		_, err := chartutil.Save(&chart.Chart{Metadata: &chart.Metadata{
			Name: "chaos-mesh", Version: version, APIVersion: chart.APIVersionV2, AppVersion: "2.2.0",
		}}, directory)
		Expect(err).NotTo(HaveOccurred())
		result, err := FetchChaosMeshChart(context.Background(), version, directory)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Name()).To(Equal("chaos-mesh"))
		Expect(result.Metadata.Version).To(Equal(version))
		Expect(result.AppVersion()).To(Equal("2.2.0"))
	})

	It("reports missing and invalid archives", func() {
		directory := GinkgoT().TempDir()
		Expect(FetchChaosMeshChart(context.Background(), version, directory)).Error().To(MatchError(os.ErrNotExist))
		path := filepath.Join(directory, "chaos-mesh-2.2.0.tgz")
		Expect(os.WriteFile(path, []byte("not a chart"), 0o600)).To(Succeed())
		Expect(FetchChaosMeshChart(context.Background(), version, directory)).Error().To(MatchError(ContainSubstring("load helm chart")))
	})
})

// This example uses the published chart repository.
func ExampleDownloadChaosMeshChartTgz() {
	path, err := DownloadChaosMeshChartTgz(context.Background(), "2.2.0")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, path)
}

// This example uses the published chart repository.
func ExampleFetchChaosMeshChart() {
	chart, err := FetchChaosMeshChart(context.Background(), "2.2.0", "")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, chart.Name())
	fmt.Fprintln(os.Stderr, chart.Metadata.Version)
	fmt.Fprintln(os.Stderr, chart.AppVersion())
}

func TestFetchRemovesDownloadedArchive(t *testing.T) {
	g := NewWithT(t)
	directory := t.TempDir()
	t.Setenv("TMPDIR", directory)
	archive, err := chartutil.Save(&chart.Chart{Metadata: &chart.Metadata{
		Name: "chaos-mesh", Version: "1.0.0", APIVersion: chart.APIVersionV2,
	}}, t.TempDir())
	g.Expect(err).NotTo(HaveOccurred())
	valid, err := os.ReadFile(archive)
	g.Expect(err).NotTo(HaveOccurred())

	previous := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = previous })
	for _, test := range []struct {
		name    string
		payload []byte
		valid   bool
	}{
		{"valid", valid, true},
		{"corrupt", []byte("incomplete archive"), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			body := &chartBody{Reader: bytes.NewReader(test.payload)}
			http.DefaultClient = &http.Client{Transport: chartTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
			})}
			result, err := FetchChaosMeshChart(t.Context(), "1.0.0", "")
			if test.valid {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(result.Name()).To(Equal("chaos-mesh"))
			} else {
				g.Expect(err).To(MatchError(ContainSubstring("load helm chart")))
			}
			g.Expect(body.closed).To(BeTrue())
			g.Expect(os.ReadDir(directory)).To(BeEmpty())
		})
	}
}
