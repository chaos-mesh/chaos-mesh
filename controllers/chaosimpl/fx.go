// Copyright 2021 Chaos Mesh Authors.
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

package chaosimpl

import (
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/awschaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/dnschaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/gcpchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/httpchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/iochaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/jvmchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/kernelchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/podchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/stresschaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/timechaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
)

var AllImpl = fx.Options(
	awschaos.Module,
	dnschaos.Module,
	httpchaos.Module,
	iochaos.Module,
	kernelchaos.Module,
	networkchaos.Module,
	podchaos.Module,
	gcpchaos.Module,
	stresschaos.Module,
	jvmchaos.Module,
	timechaos.Module,

	utils.Module)
