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

import { ElementDragData, ElementTypes } from './types'

import BareNode from '../BareNode'
import type { BareNodeProps } from '../BareNode'
import Space from '@ui/mui-extends/esm/Space'
import { Typography } from '@mui/material'
import _ from 'lodash'
import _actions from 'formik/actions'
import { useDrag } from 'react-dnd'

const actions: Record<string, string[]> = _.omit(_actions, 'PhysicalMachineChaos')

const DraggableBareNode = ({ kind, act, sx, ...rest }: BareNodeProps & ElementDragData) => {
  const [{ isDragging }, drag] = useDrag(() => ({
    type: ElementTypes.Kubernetes,
    item: {
      kind,
      act,
    },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  }))

  return <BareNode sx={{ cursor: isDragging ? 'grab' : 'pointer' }} kind={kind} {...rest} ref={drag} />
}

export default function Kubernetes() {
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
                <DraggableBareNode key={action} kind={kind} act={action}>
                  {action}
                </DraggableBareNode>
              ))
          ) : (
            <DraggableBareNode kind={kind}>{kind}</DraggableBareNode>
          )}
        </Space>
      ))}
    </Space>
  )
}
