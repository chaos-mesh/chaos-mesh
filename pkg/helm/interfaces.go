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
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

// ReleaseService introduces all the operations about Helm Release
type ReleaseService interface {
	/*GetRelease would fetch the installed release.
	 */
	GetRelease(namespace string, releaseName string) (*release.Release, error)

	/*UpgradeOrInstall would upgrade the existed release or install a new one.
	namespace is the namespace of the release, it should be an existed namespace.
	releaseName introduces the name of the release.
	chart is the chart with certain version to be installed.
	values is the values to be used in the chart, it is also so-called Config in helm's codes.
	It will return the installed/upgraded release and error if any.
	*/
	UpgradeOrInstall(namespace string, releaseName string, chart *chart.Chart, values map[string]interface{}) (*release.Release, error)

	UninstallRelease(namespace string, releaseName string) (*release.UninstallReleaseResponse, error)
}
