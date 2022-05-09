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
import DeleteIcon from '@mui/icons-material/Delete'
import { Box, Drawer, IconButton } from '@mui/material'
import { useCallback, useMemo, useRef, useState } from 'react'
import type { DropTargetMonitor, XYCoord } from 'react-dnd'
import { useDrop } from 'react-dnd'
import type { Node, ReactFlowInstance, XYPosition } from 'react-flow-renderer'
import ReactFlow, { Background, Controls, MiniMap, addEdge, useEdgesState, useNodesState } from 'react-flow-renderer'
import { useIntl } from 'react-intl'
import { v4 as uuidv4 } from 'uuid'

import Paper from '@ui/mui-extends/esm/Paper'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch, useStoreSelector } from 'store'

import { setConfirm } from 'slices/globalStatus'
import { insertWorkflowNode, removeWorkflowNode, updateWorkflowNode } from 'slices/workflows'

import AutoForm, { Belong } from 'components/AutoForm'
import i18n, { T } from 'components/T'

import { concatKindAction } from 'lib/utils'

import AdjustableEdge from './AdjustableEdge'
import { ElementDragData, ElementTypes } from './Elements/types'
import FlowNode from './FlowNode'

type DropItem = ElementDragData
type Identifier = DropItem

interface ControlProps {
  id: uuid
  onDelete: (id: uuid) => void
}

const NodeControl = ({ id, onDelete }: ControlProps) => {
  const intl = useIntl()
  const dispatch = useStoreDispatch()

  const onNodeDelete = () => {
    dispatch(
      setConfirm({
        title: `Delete node ${id}`,
        description: <T id="common.deleteDesc" />,
        handle: () => onDelete(id),
      })
    )
  }

  return (
    <Space className="nodrag" direction="row" lineHeight={1}>
      {/* TODO: Copy single Node to reuse */}
      {/* <IconButton size="small" >
        <FileCopyIcon fontSize="inherit" />
      </IconButton>
      <Divider orientation="vertical" variant="middle" flexItem /> */}
      <IconButton size="small" onClick={onNodeDelete} title={i18n('common.delete', intl)} aria-label="delete">
        <DeleteIcon fontSize="inherit" />
      </IconButton>
    </Space>
  )
}

const EdgeControl = ({ id, onDelete }: ControlProps) => {
  const intl = useIntl()
  const dispatch = useStoreDispatch()

  const onEdgeDelete = () => {
    dispatch(
      setConfirm({
        title: `Delete edge ${id}`,
        description: <T id="common.deleteDesc" />,
        handle: () => onDelete(id),
      })
    )
  }

  return (
    <Space className="nodrag" direction="row" lineHeight={1}>
      <IconButton size="small" onClick={onEdgeDelete} title={i18n('common.delete', intl)} aria-label="delete">
        <DeleteIcon fontSize="inherit" />
      </IconButton>
    </Space>
  )
}

interface WhiteboardProps {
  flowRef: React.MutableRefObject<ReactFlowInstance | undefined>
}

export default function Whiteboard({ flowRef }: WhiteboardProps) {
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const onConnect = useCallback(
    (connection) =>
      setEdges((eds) => {
        // A node can have only one incomer.
        //
        // Unless the `source` of the `connection` has the same parent as the existing incomer.
        // const hasConnected = eds.find((e) => e.target === connection.target)

        // if (hasConnected) {
        //   const hasConnectedParent = eds.find((e) => e.target === hasConnected.source)

        //   if (eds.some((e) => e.source === hasConnectedParent?.source && e.target === connection.source)) {
        //     return addEdge({ ...connection, type: 'smoothstep' }, eds)
        //   }

        //   return eds
        // }

        const id = uuidv4()

        return addEdge(
          {
            ...connection,
            type: 'adjustableEdge',
            data: {
              id,
              tooltipProps: {
                title: <EdgeControl id={id} onDelete={deleteEdge} />,
              },
            },
          },
          eds
        )
      }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [setEdges]
  )
  const nodeTypes = useMemo(() => ({ flowNode: FlowNode }), [])
  const edgeTypes = useMemo(() => ({ adjustableEdge: AdjustableEdge }), [])

  const store = useStoreSelector((state) => state.workflows)
  const dispatch = useStoreDispatch()

  const [openDrawer, setOpenDrawer] = useState(false)
  const [identifier, setIdentifier] = useState<Identifier | null>(null)
  const formInitialValues = useRef()
  const cleanup = () => {
    setOpenDrawer(false)
    setIdentifier(null)
    formInitialValues.current = undefined
  }
  const closeDrawer = () => {
    const id = identifier!.id!
    const lastWorkflowNode = store.nodes[id]

    // Remove empty node.
    if (!lastWorkflowNode) {
      setNodes(nodes.filter((node) => node.id !== id))
      dispatch(removeWorkflowNode(id))
    }

    cleanup()
  }

  const addNode = (item: DropItem, monitor: DropTargetMonitor, xyCoord?: XYCoord) => {
    const whiteboardRect = document.getElementById('workflows-whiteboard')!.getBoundingClientRect()
    let position: XYPosition

    if (xyCoord) {
      position = xyCoord
    } else {
      const sourceRect = monitor.getSourceClientOffset()

      position = { x: sourceRect!.x - whiteboardRect.x, y: sourceRect!.y - whiteboardRect.y }
    }

    const id = uuidv4()

    setNodes((oldNodes) => [
      ...oldNodes,
      {
        id,
        type: 'flowNode',
        position,
        data: {
          origin: oldNodes.length === 0,
          tooltipProps: {
            title: <NodeControl id={id} onDelete={deleteNode} />,
          },
          kind: item.kind,
          children: concatKindAction(item.kind, item.act),
        },
      },
    ])
    dispatch(insertWorkflowNode({ id, experiment: null })) // Insert only id to distinguish from other nodes.

    return id
  }

  const editNode = (_: any, { id }: Node) => {
    const workflowNode = store.nodes[id]

    formInitialValues.current = workflowNode

    setIdentifier({
      id,
      kind: workflowNode.kind,
      act: workflowNode.action,
    })
    setOpenDrawer(true)
  }

  const updateNode = (values: Record<string, any>) => {
    setNodes((oldNodes) =>
      oldNodes.map((node) => {
        if (node.id === values.id) {
          return {
            ...node,
            data: {
              ...node.data,
              disableRipple: true,
              sx: {
                alignItems: 'start',
                '& .MuiButton-startIcon': {
                  mt: 0.5,
                },
              },
              children: (
                <Box>
                  <Box>
                    {values.name} ({concatKindAction(values.kind, values.action)})
                  </Box>
                  <Box>deadline: {values.deadline}</Box>
                </Box>
              ),
            },
          }
        }

        return node
      })
    )
    dispatch(updateWorkflowNode(values))

    cleanup()
  }

  const deleteNode = (id: uuid) => {
    setNodes((oldNodes) => oldNodes.filter((node) => node.id !== id))
    dispatch(removeWorkflowNode(id))
  }

  const initNode = (item: DropItem, monitor: DropTargetMonitor, xyCoord?: XYCoord) => {
    // Add new node into flow.
    const id = addNode(item, monitor, xyCoord)
    // Start generating form.
    setIdentifier({ ...item, id })

    // Open form builder after adding new node.
    setOpenDrawer(true)
  }

  const [, drop] = useDrop(() => ({
    accept: [ElementTypes.Kubernetes, ElementTypes.PhysicalNodes, ElementTypes.Suspend],
    drop: initNode,
  }))

  const deleteEdge = (id: uuid) => {
    setEdges((oldEdges) => oldEdges.filter((edge) => edge.data.id !== id))
  }

  return (
    <>
      <ReactFlow
        ref={drop}
        id="workflows-whiteboard"
        onInit={(flow) => {
          if (flowRef) {
            ;(flow as any).initNode = initNode

            flowRef.current = flow
          }
        }}
        nodes={nodes}
        onNodesChange={onNodesChange}
        onNodeClick={editNode}
        edges={edges}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
      >
        <Background />
        <Controls />
        <MiniMap style={{ top: 0, right: 0 }} />
      </ReactFlow>
      <Drawer anchor="right" open={openDrawer} onClose={closeDrawer}>
        <Paper sx={{ width: 768, pr: 16, overflowY: 'auto' }}>
          {identifier && (
            <AutoForm
              {...identifier}
              belong={Belong.Workflow}
              formikProps={{ initialValues: formInitialValues.current, onSubmit: updateNode }}
            />
          )}
        </Paper>
      </Drawer>
    </>
  )
}
