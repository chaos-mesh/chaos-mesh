// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

// Parameters parameters
type Parameters struct {
	Addr                string // addr to serve on
	CertFile            string // path to the x509 certificate for https
	KeyFile             string // path to the x509 private key matching `CertFile`
	ConfigDirectory     string // path to sidecar injector configuration directory (contains yamls)
	AnnotationNamespace string // namespace used to scope annotations
}
