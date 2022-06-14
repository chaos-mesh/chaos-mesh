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
import { Drawer, IconButton, ListItemIcon, ListItemText, MenuItem } from '@mui/material'
import _ from 'lodash'
import { SyntheticEvent, useCallback, useMemo, useRef, useState } from 'react'
import type { DropTargetMonitor, XYCoord } from 'react-dnd'
import { useDrop } from 'react-dnd'
import type { Node, ReactFlowInstance, XYPosition } from 'react-flow-renderer'
import ReactFlow, { Background, Controls, MiniMap, addEdge, useEdgesState, useNodesState } from 'react-flow-renderer'
import { useIntl } from 'react-intl'
import { v4 as uuidv4 } from 'uuid'

import Menu from '@ui/mui-extends/esm/Menu'
import Paper from '@ui/mui-extends/esm/Paper'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch, useStoreSelector } from 'store'

import { setConfirm } from 'slices/globalStatus'
import { importNodes, insertWorkflowNode, removeWorkflowNode, updateWorkflowNode } from 'slices/workflows'

import AutoForm, { Belong } from 'components/AutoForm'
import i18n, { T } from 'components/T'

import { concatKindAction } from 'lib/utils'

import AdjustableEdge from './AdjustableEdge'
import { ElementDragData } from './Elements/types'
import FlowNode from './FlowNode'
import GroupNode, { ResizableHandleClassName } from './GroupNode'
import { dndAccept } from './data'
import { SpecialTemplateType, workflowToFlow } from './utils/convert'

export type DropItem = ElementDragData
type Identifier = DropItem

interface ControlProps {
  id: uuid
  onDelete: (id: uuid) => void
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
    <Space direction="row" lineHeight={1}>
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
  const nodeTypes = useMemo(() => ({ flowNode: FlowNode, groupNode: GroupNode }), [])
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

      if (store.nodes[id]) {
        dispatch(removeWorkflowNode(id))
      }
    }

    cleanup()
  }

  const addNode = (item: DropItem, monitor?: DropTargetMonitor, xyCoord?: XYCoord, parent?: uuid) => {
    const whiteboardRect = document.getElementById('workflows-whiteboard')!.getBoundingClientRect()
    let position: XYPosition

    if (xyCoord) {
      position = xyCoord
    } else {
      const sourceRect = monitor?.getSourceClientOffset()

      position = { x: sourceRect!.x - whiteboardRect.x, y: sourceRect!.y - whiteboardRect.y }
    }

    const id = uuidv4()
    const node: Node = {
      id,
      position,
      data: {},
      ...(parent && {
        parentNode: parent,
        extent: 'parent',
      }),
    }

    if (item.kind === SpecialTemplateType.Serial || item.kind === SpecialTemplateType.Parallel) {
      node.type = 'groupNode'
      node.data = {
        id,
        name: _.truncate(`${item.kind}-${id}`),
        type: item.kind,
        childrenNum: 1,
        actions: {
          initNode,
        },
      }
      node.zIndex = -1 // Make edges visible on the top of the group node.
    } else {
      node.type = 'flowNode'
      node.data = {
        kind: item.kind,
        children: concatKindAction(item.kind, item.act),
      }
    }

    setNodes((oldNodes) => [...oldNodes, node])
    dispatch(insertWorkflowNode({ id, experiment: null })) // Insert only id to distinguish from other nodes.

    return id
  }

  const editNode = (e: React.MouseEvent, { id }: Node) => {
    // Prevent editing nodes when resizing.
    //
    // See `GroupNode.tsx` for more details.
    if (
      (e.target as HTMLDivElement).className === 're-resizable' ||
      (e.target as HTMLDivElement).className === ResizableHandleClassName
    ) {
      return
    }

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
              ...(node.type === 'flowNode' && {
                sx: {
                  alignItems: 'start',
                  '& .MuiButton-startIcon': {
                    mt: 0.5,
                  },
                },
                endIcon: (
                  <Menu
                    IconButtonProps={{ size: 'small', sx: { position: 'absolute', right: 8, mt: -0.5 } }}
                    IconProps={{ fontSize: 'inherit' }}
                  >
                    <MenuItem onClick={onNodeDelete(node.id)}>
                      <ListItemIcon sx={{ fontSize: 18 }}>
                        <DeleteIcon fontSize="inherit" />
                      </ListItemIcon>
                      <ListItemText primaryTypographyProps={{ variant: 'button', color: 'secondary' }}>
                        <T id="common.delete" />
                      </ListItemText>
                    </MenuItem>
                    {/* TODO: Copy single Node to reuse */}
                    {/* <MenuItem>
                      <ListItemIcon sx={{ fontSize: 18 }}>
                        <ContentCopyIcon fontSize="inherit" />
                      </ListItemIcon>
                      <ListItemText primaryTypographyProps={{ variant: 'button', color: 'secondary' }}>
                        Copy
                      </ListItemText>
                    </MenuItem> */}
                  </Menu>
                ),
                children: values.name,
              }),
              // Serial or Parallel.
              ...(node.type === 'groupNode' && {
                name: values.name,
              }),
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

  const onNodeDelete = (id: uuid) => (e: SyntheticEvent) => {
    e.stopPropagation()

    dispatch(
      setConfirm({
        title: `Delete node ${id}`,
        description: <T id="common.deleteDesc" />,
        handle: () => deleteNode(id),
      })
    )
  }

  const initNode = (item: DropItem, monitor?: DropTargetMonitor, xyCoord?: XYCoord, parent?: uuid) => {
    // If `xyCoord` is `undefined`, the item isn't operated by dragging and dropping.
    if (!monitor?.isOver({ shallow: true }) && !xyCoord) {
      return
    }

    // Add new node into flow.
    const id = addNode(item, monitor, xyCoord, parent)

    // Start generating form.
    setIdentifier({ ...item, id })

    // Open form builder after adding new node.
    setOpenDrawer(true)
  }

  const [, drop] = useDrop(() => ({
    accept: dndAccept,
    drop: initNode,
  }))

  const deleteEdge = (id: uuid) => {
    setEdges((oldEdges) => oldEdges.filter((edge) => edge.data.id !== id))
  }

  const importWorkflow = (workflow: string) => {
    const { store, nodes, edges } = workflowToFlow(workflow)

    dispatch(importNodes(store))
    setNodes(nodes)
    setEdges(
      edges.map((edge) => ({
        ...edge,
        data: { ...edge.data, tooltipProps: { title: <EdgeControl id={edge.data.id} onDelete={deleteEdge} /> } },
      }))
    )
  }

  const updateNodeDraggable = (id: uuid, draggable: boolean) => {
    setNodes((nodes) =>
      nodes.map((node) => {
        if (node.id === id) {
          return {
            ...node,
            draggable,
          }
        }

        return node
      })
    )
  }

  const onNodeMouseMove = (e: React.MouseEvent, { id }: Node) => {
    // Resume dragging nodes after resizing.
    //
    // See `GroupNode.tsx` for more details.
    if ((e.target as HTMLDivElement).className === ResizableHandleClassName) {
      updateNodeDraggable(id, false)
    } else {
      updateNodeDraggable(id, true)
    }
  }

  return (
    <>
      <ReactFlow
        ref={drop}
        id="workflows-whiteboard"
        onInit={(flow) => {
          if (flowRef) {
            ;(flow as any).initNode = initNode
            ;(flow as any).importWorkflow = importWorkflow

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
        onNodeMouseMove={onNodeMouseMove}
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
