package helm

import (
	"context"
	"fmt"
	"os"
)

func ExampleDownloadChaosMeshChartTgz() {
	path, err := DownloadChaosMeshChartTgz(context.Background(), "2.2.0")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, path)
	// Output:
}

func ExampleFetchChaosMeshChart() {
	chart, err := FetchChaosMeshChart(context.Background(), "2.2.0")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, chart.Name())
	fmt.Fprintln(os.Stderr, chart.Metadata.Version)
	fmt.Fprintln(os.Stderr, chart.AppVersion())
	// Output:
}
