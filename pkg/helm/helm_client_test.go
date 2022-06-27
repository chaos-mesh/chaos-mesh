package helm

import (
	"context"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"helm.sh/helm/v3/pkg/cli"
	"os"
)

func ExampleHelmClientUpgradeOrInstall() {
	chart, err := FetchChaosMeshChart(context.Background(), "2.2.0")
	if err != nil {
		panic(err)
	}

	settings := cli.New()
	restClientGetter := settings.RESTClientGetter()
	logger, err := log.NewDefaultZapLogger()
	if err != nil {
		panic(err)
	}
	client, err := NewHelmClient(restClientGetter, logger)
	if err != nil {
		panic(err)
	}
	_, err = client.UpgradeOrInstall(
		"chaos-mesh",
		"chaos-mesh-in-remote-cluster",
		chart,
		nil,
	)
	if err != nil {
		panic(err)
	}
	// Output:
}

func ExampleHelmClientGetRelease() {
	settings := cli.New()
	restClientGetter := settings.RESTClientGetter()
	logger, err := log.NewDefaultZapLogger()
	if err != nil {
		panic(err)
	}
	client, err := NewHelmClient(restClientGetter, logger)
	if err != nil {
		panic(err)
	}
	release, err := client.GetRelease(
		"chaos-mesh",
		"chaos-mesh-in-remote-cluster",
	)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, release.Name)
	// Output:
}

func ExampleHelmClientUninstallRelease() {
	settings := cli.New()
	restClientGetter := settings.RESTClientGetter()
	logger, err := log.NewDefaultZapLogger()
	if err != nil {
		panic(err)
	}
	client, err := NewHelmClient(restClientGetter, logger)
	if err != nil {
		panic(err)
	}
	_, err = client.UninstallRelease(
		"chaos-mesh",
		"chaos-mesh-in-remote-cluster",
	)
	if err != nil {
		panic(err)
	}
	// Output:
}
