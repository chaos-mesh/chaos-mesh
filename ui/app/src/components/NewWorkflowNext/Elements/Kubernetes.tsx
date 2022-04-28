/*
 * Copyright 2022 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { Typography } from '@mui/material'
import _actions from 'formik/actions'
import _ from 'lodash'

import Space from '@ui/mui-extends/esm/Space'

import DraggableBareNode from './DraggableBareNode'
import { ElementTypes, ElementsProps } from './types'

const actions: Record<string, string[]> = _.omit(_actions, 'PhysicalMachineChaos')

export default function Kubernetes({ onElementClick }: ElementsProps) {
  return (
    <Space>
      {Object.entries(actions).map(([kind, list]) => (
        <Space key={kind}>
          <Typography variant="body2" fontWeight="medium">
            {kind}
          </Typography>
          {list.length > 0 ? (
            list
              .filter((action) => action !== 'netem') // TODO: support NetworkChaos/netem
              .map((action) => (
                // TODO: refactor ExperimentKind
                <DraggableBareNode
                  key={action}
                  elementType={ElementTypes.Kubernetes}
                  kind={kind}
                  act={action}
                  onNodeClick={onElementClick}
                >
                  {action}
                </DraggableBareNode>
              ))
          ) : (
            <DraggableBareNode elementType={ElementTypes.Kubernetes} kind={kind} onNodeClick={onElementClick}>
              {kind}
            </DraggableBareNode>
          )}
        </Space>
      ))}
    </Space>
  )
}
