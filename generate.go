// Copyright 2021 Chaos Mesh Authors.
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

package chaosmesh

//go:generate go tool client-gen --input=github.com/chaos-mesh/chaos-mesh/api/v1alpha1 --input-base= --output-dir=./pkg/client --output-pkg=github.com/chaos-mesh/chaos-mesh/pkg/client/ --clientset-name=versioned --go-header-file=./hack/boilerplate/boilerplate.generatego.txt --fake-clientset=true --plural-exceptions=PodChaos:podchaos -v=2
