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

import { Handle, Position } from 'react-flow-renderer'
import { Tooltip, styled, tooltipClasses } from '@mui/material'

import BareNode from './BareNode'
import type { BareNodeProps } from './BareNode'
import type { Node } from 'react-flow-renderer'
import type { TooltipProps } from '@mui/material'

const StyledHandle = styled(Handle)(({ theme }) => ({
  width: '8px !important',
  height: '8px !important',
  background: `${theme.palette.background.default} !important`,
  borderColor: `${theme.palette.outline.main} !important`,
  zIndex: 1,
}))

const NodeTooltip = styled(({ className, ...props }: TooltipProps) => (
  <Tooltip {...props} classes={{ popper: className }} />
))(({ theme }) => ({
  [`& .${tooltipClasses.arrow}`]: {
    color: theme.palette.surfaceVariant.main,
  },
  [`& .${tooltipClasses.tooltip}`]: {
    backgroundColor: theme.palette.surfaceVariant.main,
    color: theme.palette.onSurfaceVariant.main,
  },
}))

export type FlowNodeProps = Node<BareNodeProps & { origin?: boolean; tooltipProps: TooltipProps }>

export default function FlowNode({ data }: FlowNodeProps) {
  const { origin, tooltipProps, ...rest } = data

  return (
    <>
      {!origin && <StyledHandle type="target" position={Position.Left} />}
      <NodeTooltip arrow placement="top" {...tooltipProps}>
        <BareNode {...rest} />
      </NodeTooltip>
      <StyledHandle type="source" position={Position.Right} />
    </>
  )
}
