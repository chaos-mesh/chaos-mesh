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

package schedule

import (
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/active"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/cron"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/gc"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/pause"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/utils"
)

var Module = fx.Provide(
	fx.Annotated{
		Group:  "controller",
		Target: cron.NewController,
	},
	fx.Annotated{
		Group:  "controller",
		Target: active.NewController,
	},
	fx.Annotated{
		Group:  "controller",
		Target: gc.NewController,
	},

	fx.Annotated{
		Group:  "controller",
		Target: pause.NewController,
	},
	utils.NewActiveLister,
)
