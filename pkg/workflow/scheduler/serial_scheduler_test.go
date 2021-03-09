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

package scheduler

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	workflowerrors "github.com/chaos-mesh/chaos-mesh/pkg/workflow/errors"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/model/mock_node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/model/mock_template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/model/mock_workflow"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/model/template"
)

func TestScheduleWithSerial(t *testing.T) {

	tests := []struct {
		name                       string
		childrenTemplates          []string
		succeedChildren            int
		expectedScheduledTemplates []string
		expectedError              error
	}{
		{
			name:                       "no-schedule",
			childrenTemplates:          nil,
			succeedChildren:            0,
			expectedScheduledTemplates: nil,
			expectedError:              workflowerrors.ErrNoMoreTemplateInSerialTemplate,
		}, {
			name:                       "alternative-no-schedule",
			childrenTemplates:          []string{},
			succeedChildren:            0,
			expectedScheduledTemplates: nil,
			expectedError:              workflowerrors.ErrNoMoreTemplateInSerialTemplate,
		}, {
			name:                       "schedule-one-by-one-0",
			childrenTemplates:          []string{"child-0", "child-1", "child-2"},
			succeedChildren:            0,
			expectedScheduledTemplates: []string{"child-0"},
			expectedError:              nil,
		}, {
			name:                       "schedule-one-by-one-1",
			childrenTemplates:          []string{"child-0", "child-1", "child-2"},
			succeedChildren:            1,
			expectedScheduledTemplates: []string{"child-1"},
			expectedError:              nil,
		}, {
			name:                       "schedule-one-by-one-2",
			childrenTemplates:          []string{"child-0", "child-1", "child-2"},
			succeedChildren:            2,
			expectedScheduledTemplates: []string{"child-2"},
			expectedError:              nil,
		}, {
			name:                       "schedule-one-by-one-final",
			childrenTemplates:          []string{"child-0", "child-1", "child-2"},
			succeedChildren:            3,
			expectedScheduledTemplates: nil,
			expectedError:              workflowerrors.ErrNoMoreTemplateInSerialTemplate,
		}, {
			name:                       "schedule-duplicated-template-0",
			childrenTemplates:          []string{"child-0", "child-1", "child-0", "child-1"},
			succeedChildren:            0,
			expectedScheduledTemplates: []string{"child-0"},
			expectedError:              nil,
		}, {
			name:                       "schedule-duplicated-template-1",
			childrenTemplates:          []string{"child-0", "child-1", "child-0", "child-1"},
			succeedChildren:            1,
			expectedScheduledTemplates: []string{"child-1"},
			expectedError:              nil,
		}, {
			name:                       "schedule-duplicated-template-2",
			childrenTemplates:          []string{"child-0", "child-1", "child-0", "child-1"},
			succeedChildren:            2,
			expectedScheduledTemplates: []string{"child-0"},
			expectedError:              nil,
		}, {
			name:                       "schedule-duplicated-template-3",
			childrenTemplates:          []string{"child-0", "child-1", "child-0", "child-1"},
			succeedChildren:            3,
			expectedScheduledTemplates: []string{"child-1"},
			expectedError:              nil,
		}, {
			name:                       "schedule-duplicated-template-4",
			childrenTemplates:          []string{"child-0", "child-1", "child-0", "child-1"},
			succeedChildren:            4,
			expectedScheduledTemplates: nil,
			expectedError:              workflowerrors.ErrNoMoreTemplateInSerialTemplate,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockctl := gomock.NewController(t)
			const serialTemplateName = "mock-serial"
			const serialNodeName = serialTemplateName + "-0000"
			const workflowName = "mock-workflow"

			mockSerialTemplate := mock_template.NewMockSerialTemplate(mockctl)
			mockSerialTemplate.EXPECT().TemplateType().Return(template.Serial).AnyTimes()
			mockSerialTemplate.EXPECT().SerialChildrenList().Return(test.childrenTemplates).AnyTimes()

			mockWorkflowSpec := mock_workflow.NewMockWorkflowSpec(mockctl)
			mockWorkflowSpec.EXPECT().Name().Return(workflowName).AnyTimes()

			mockTreeNode := mock_node.NewMockNodeTreeNode(mockctl)
			nodeTreeChildren := mock_node.NewMockNodeTreeChildren(mockctl)
			mockTreeNode.EXPECT().Name().Return(serialNodeName).AnyTimes()
			mockTreeNode.EXPECT().Children().Return(nodeTreeChildren).AnyTimes()
			mockTreeNode.EXPECT().TemplateName().Return(serialTemplateName).AnyTimes()
			nodeTreeChildren.EXPECT().Length().Return(test.succeedChildren).AnyTimes()

			mockNode := mock_node.NewMockNode(mockctl)
			mockNode.EXPECT().Name().Return(serialNodeName).AnyTimes()
			mockNode.EXPECT().TemplateName().Return(serialTemplateName).AnyTimes()

			mockWorkflowSpec.EXPECT().FetchTemplateByName(gomock.Eq(serialTemplateName)).Return(mockSerialTemplate, nil).AnyTimes()
			for _, childTemplate := range test.childrenTemplates {
				mockChildTemplate := mock_template.NewMockSerialTemplate(mockctl)
				mockChildTemplate.EXPECT().Name().Return(childTemplate).AnyTimes()
				mockWorkflowSpec.EXPECT().FetchTemplateByName(gomock.Eq(childTemplate)).Return(mockChildTemplate, nil).AnyTimes()
			}

			scheduler := NewSerialScheduler(mockWorkflowSpec, mockNode, mockTreeNode)
			nextTemplates, parentNodeName, err := scheduler.ScheduleNext(context.TODO())
			if test.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, test.expectedError))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, serialNodeName, parentNodeName)
				var names []string
				for _, item := range nextTemplates {
					names = append(names, item.Name())
				}
				assert.Equal(t, test.expectedScheduledTemplates, names)
			}
		})
	}
}
