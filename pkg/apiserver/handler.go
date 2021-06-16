// Copyright 2020 Chaos Mesh Authors.
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

package apiserver

import (
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/archive"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/event"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/experiment"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/schedule"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/workflow"
)

var handlerModule = fx.Options(
	fx.Provide(
		common.NewService,
		experiment.NewService,
		event.NewService,
		archive.NewService,
		workflow.NewService,
		schedule.NewService,
	),
	fx.Invoke(
		common.Register,
		experiment.Register,
		event.Register,
		archive.Register,
		workflow.Register,
		schedule.Register,
	),
)
