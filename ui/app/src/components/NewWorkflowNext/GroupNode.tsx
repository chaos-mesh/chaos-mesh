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
import { Box, Typography } from '@mui/material'
import { Resizable } from 're-resizable'
import { DropTargetMonitor, XYCoord, useDrop } from 'react-dnd'
import type { NodeProps } from 'react-flow-renderer'
import { Position } from 'react-flow-renderer'

import Paper from '@ui/mui-extends/esm/Paper'
import Space from '@ui/mui-extends/esm/Space'

import { iconByKind } from 'lib/byKind'

import StyledHandle from './StyleHandle'
import { DropItem } from './Whiteboard'
import { dndAccept } from './data'
import { SpecialTemplateType } from './utils/convert'

export const ResizableHandleClassName = 're-resizable-handle'
const handleClasses = {
  top: ResizableHandleClassName,
  right: ResizableHandleClassName,
  bottom: ResizableHandleClassName,
  left: ResizableHandleClassName,
  topRight: ResizableHandleClassName,
  bottomRight: ResizableHandleClassName,
  bottomLeft: ResizableHandleClassName,
  topLeft: ResizableHandleClassName,
}

interface GroupNodeProps {
  id: uuid
  name: React.ReactNode
  type: SpecialTemplateType.Serial | SpecialTemplateType.Parallel
  childrenNum: number
  width: number
  height: number
  actions: {
    initNode: (item: DropItem, monitor?: DropTargetMonitor, xyCoord?: XYCoord, parent?: uuid) => void
  }
}

export default function GroupNode({ data, isConnectable }: NodeProps<GroupNodeProps>) {
  const {
    id,
    name,
    type,
    childrenNum,
    width = type === SpecialTemplateType.Serial ? 200 * childrenNum + 30 * (childrenNum + 1) : 200 + 30 * childrenNum,
    height = type === SpecialTemplateType.Parallel ? 30 * childrenNum + 15 * (childrenNum + 1) : 30 + 15 * childrenNum,
    actions,
  } = data
  const groupNodeID = `group-node-${id}`

  const [{ isOverCurrent }, drop] = useDrop(() => ({
    accept: dndAccept,
    drop: (item: DropItem, monitor: DropTargetMonitor) => {
      const groupNodeRect = document.getElementById(groupNodeID)!.getBoundingClientRect()
      const sourceRect = monitor.getSourceClientOffset()
      const position = { x: sourceRect!.x - groupNodeRect.x, y: sourceRect!.y - groupNodeRect.y }

      actions.initNode(item, monitor, position, id)
    },
    collect: (monitor) => ({
      isOverCurrent: monitor.isOver({ shallow: true }),
    }),
  }))

  return (
    <Box id={groupNodeID} ref={drop}>
      <Space direction="row" spacing={1} alignItems="center" sx={{ color: 'secondary.main', fontSize: 18 }}>
        {iconByKind(type, 'inherit')}
        <Typography component="div" fontWeight="medium">
          {name}
        </Typography>
      </Space>
      <Resizable
        className="re-resizable"
        handleClasses={handleClasses}
        defaultSize={{ width, height }}
        onResize={(e) => e.stopImmediatePropagation()}
      >
        <StyledHandle type="target" position={Position.Left} isConnectable={isConnectable} />
        <Paper
          sx={{
            '&:hover': { cursor: 'pointer' },
            ...(isOverCurrent && { bgcolor: 'background.default', borderColor: 'secondary.main' }),
          }}
        />
        <StyledHandle type="source" position={Position.Right} isConnectable={isConnectable} />
      </Resizable>
    </Box>
  )
}
