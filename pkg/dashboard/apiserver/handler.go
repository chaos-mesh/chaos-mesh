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

package apiserver

import (
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/archive"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/auth/gcp"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/event"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/experiment"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/schedule"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/workflow"
)

var handlerModule = fx.Options(
	fx.Provide(
		common.NewService,
		experiment.Bootstrap,
		schedule.Bootstrap,
		workflow.Bootstrap,
		event.NewService,
		archive.NewService,
		gcp.NewService,
		template.Bootstrap,
	),
	fx.Invoke(
		// gcp should register at the first, because it registers a middleware
		gcp.Register,
		common.Register,
		experiment.Register,
		schedule.Register,
		workflow.Register,
		event.Register,
		archive.Register,
		template.Register,
	),
)
