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

package workflow

//go:generate rm -rf ./mock
//go:generate mockgen -source ./model/workflow/workflow_frontend.go -copyright_file ../../hack/boilerplate/boilerplate.gomock.txt -destination ./mock/model/mock_workflow/mock_workflow.go

//go:generate mockgen -source ./model/node/node.go -copyright_file ../../hack/boilerplate/boilerplate.gomock.txt -destination ./mock/model/mock_node/mock_node.go

//go:generate mockgen -source ./model/template/template.go -copyright_file ../../hack/boilerplate/boilerplate.gomock.txt -destination ./mock/model/mock_template/mock_template.go
//go:generate mockgen -source ./model/template/serial.go -copyright_file ../../hack/boilerplate/boilerplate.gomock.txt -aux_files github.com/chaos-mesh/chaos-mesh/pkg/workflow/model/template=./model/template/template.go -destination ./mock/model/mock_template/mock_serial.go
//go:generate mockgen -source ./model/template/task.go -copyright_file ../../hack/boilerplate/boilerplate.gomock.txt -aux_files github.com/chaos-mesh/chaos-mesh/pkg/workflow/model/template=./model/template/template.go -destination ./mock/model/mock_template/mock_task.go
//go:generate mockgen -source ./model/template/parallel.go -copyright_file ../../hack/boilerplate/boilerplate.gomock.txt -aux_files github.com/chaos-mesh/chaos-mesh/pkg/workflow/model/template=./model/template/template.go -destination ./mock/model/mock_template/mock_parallel.go
//go:generate mockgen -source ./model/template/suspend.go -copyright_file ../../hack/boilerplate/boilerplate.gomock.txt -aux_files github.com/chaos-mesh/chaos-mesh/pkg/workflow/model/template=./model/template/template.go -destination ./mock/model/mock_template/mock_suspend.go
