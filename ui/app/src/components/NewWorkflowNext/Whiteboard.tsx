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

import AutoForm, { Belong } from 'components/AutoForm'
import { Box, Drawer } from '@mui/material'
import { ElementDragData, ElementTypes } from './Elements/types'
import type { Node, ReactFlowInstance } from 'react-flow-renderer'
import ReactFlow, { Background, Controls, MiniMap, addEdge, useEdgesState, useNodesState } from 'react-flow-renderer'
import { insertWorkflowNode, removeWorkflowNode, updateWorkflowNode } from 'slices/workflows'
import { useCallback, useMemo, useRef, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'

import DeleteIcon from '@mui/icons-material/Delete'
import type { DropTargetMonitor } from 'react-dnd'
import FlowNode from './FlowNode'
import Paper from '@ui/mui-extends/esm/Paper'
import Space from '@ui/mui-extends/esm/Space'
import { T } from 'components/T'
import { concatKindAction } from 'lib/utils'
import { setConfirm } from 'slices/globalStatus'
import { useDrop } from 'react-dnd'
import { v4 as uuidv4 } from 'uuid'

type DropItem = ElementDragData
type Identifier = DropItem

interface NodeControlProps {
  id: uuid
  onDelete: (id: uuid) => void
}

const NodeControl = ({ id, onDelete }: NodeControlProps) => {
  const dispatch = useStoreDispatch()

  const onClick = () => {
    dispatch(
      setConfirm({
        title: `Delete Node ${id}`,
        description: <T id="common.deleteDesc" />,
        handle: () => onDelete(id),
      })
    )
  }

  return (
    <Space direction="row" lineHeight={1}>
      <Box className="nodrag" title="Delete" onClick={onClick} sx={{ cursor: 'pointer' }}>
        <DeleteIcon fontSize="small" />
      </Box>
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

        return addEdge({ ...connection, type: 'smoothstep' }, eds)
      }),
    [setEdges]
  )
  const nodeTypes = useMemo(() => ({ flowNode: FlowNode }), [])

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

  const addNode = (item: DropItem, monitor: DropTargetMonitor) => {
    const whiteboardRect = document.getElementById('workflows-whiteboard')!.getBoundingClientRect()
    const sourceRect = monitor.getSourceClientOffset()

    const id = uuidv4()

    setNodes((oldNodes) => [
      ...oldNodes,
      {
        id,
        type: 'flowNode',
        position: { x: sourceRect!.x - whiteboardRect.x, y: sourceRect!.y - whiteboardRect.y },
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

  const updateNode = (values: Record<string, any>) => {
    setNodes((oldNodes) =>
      oldNodes.map((node) => {
        if (node.id === values.id) {
          return {
            ...node,
            data: { ...node.data, children: `${values.name} (${concatKindAction(values.kind, values.action)})` },
          }
        }

        return node
      })
    )
    dispatch(updateWorkflowNode(values))

    cleanup()
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

  const deleteNode = (id: uuid) => {
    setNodes((oldNodes) => oldNodes.filter((node) => node.id !== id))
    dispatch(removeWorkflowNode(id))
  }

  const [, drop] = useDrop(() => ({
    accept: [ElementTypes.Kubernetes, ElementTypes.PhysicalNodes],
    drop: (item: DropItem, monitor) => {
      // Add new node into flow.
      const id = addNode(item, monitor)
      // Start generating form.
      setIdentifier({ ...item, id })

      // Open form builder after adding new node.
      setOpenDrawer(true)
    },
  }))

  return (
    <>
      <ReactFlow
        ref={drop}
        id="workflows-whiteboard"
        onInit={(flow) => {
          if (flowRef) {
            flowRef.current = flow
          }
        }}
        nodes={nodes}
        onNodesChange={onNodesChange}
        onNodeClick={editNode}
        onNodeMouseEnter={() => {}}
        edges={edges}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        nodeTypes={nodeTypes}
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
