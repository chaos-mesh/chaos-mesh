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
import { memo } from 'react'
import { Position } from 'react-flow-renderer'
import type { NodeProps } from 'react-flow-renderer'

import BareNode from './BareNode'
import type { BareNodeProps } from './BareNode'
import StyledHandle from './StyleHandle'

export type FlowNodeProps = NodeProps<BareNodeProps & { finished: true; name: string }>

function FlowNode({ data, isConnectable }: FlowNodeProps) {
  const { finished, ...rest } = data // Exclude `finished` from the data.

  return (
    <>
      {isConnectable && <StyledHandle type="target" position={Position.Left} />}
      <BareNode {...rest} />
      {isConnectable && <StyledHandle type="source" position={Position.Right} />}
    </>
  )
}

export default memo(FlowNode)
