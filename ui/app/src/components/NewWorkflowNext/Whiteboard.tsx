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
import ContentCopyIcon from '@mui/icons-material/ContentCopy'
import DeleteIcon from '@mui/icons-material/Delete'
import { Drawer, IconButton, ListItemIcon, ListItemText, MenuItem } from '@mui/material'
import _ from 'lodash'
import React, { useCallback, useMemo, useRef, useState } from 'react'
import type { DropTargetMonitor, XYCoord } from 'react-dnd'
import { useDrop } from 'react-dnd'
import { MarkerType, Node, ReactFlowInstance, XYPosition } from 'react-flow-renderer'
import ReactFlow, { Background, Controls, MiniMap, addEdge, useEdgesState, useNodesState } from 'react-flow-renderer'
import { useIntl } from 'react-intl'
import { v4 as uuidv4 } from 'uuid'

import Menu from '@ui/mui-extends/esm/Menu'
import Paper from '@ui/mui-extends/esm/Paper'

import { useStoreDispatch, useStoreSelector } from 'store'

import { setConfirm } from 'slices/globalStatus'
import { importNodes, removeWorkflowNode, updateWorkflowNode } from 'slices/workflows'

import AutoForm, { Belong } from 'components/AutoForm'
import i18n, { T } from 'components/T'

import { concatKindAction } from 'lib/utils'

import AdjustableEdge from './AdjustableEdge'
import { ElementDragData } from './Elements/types'
import FlowNode from './FlowNode'
import GroupNode, { ResizableHandleClassName } from './GroupNode'
import { dndAccept } from './data'
import { SpecialTemplateType, workflowToFlow } from './utils/convert'

const commonMarkerEnd = {
  type: MarkerType.ArrowClosed,
  width: 18,
  height: 18,
}

export type DropItem = ElementDragData
type Identifier = DropItem

interface ControlProps {
  id: uuid
  onDelete: (id: uuid) => void
}

interface NodeControlProps extends ControlProps {
  type: 'flowNode' | 'groupNode'
  onCopy: (id: uuid) => void
}

const NodeControl = ({ id, type, onDelete, onCopy }: NodeControlProps) => {
  const intl = useIntl()
  const dispatch = useStoreDispatch()

  const onNode = (type: 'copy' | 'delete', onClose: any) => (e: React.SyntheticEvent) => {
    e.stopPropagation()

    let action: typeof onDelete
    switch (type) {
      case 'copy':
        action = onCopy
        break
      case 'delete':
        action = onDelete
        break
      default:
        break
    }

    dispatch(
      setConfirm({
        title: `${i18n(`common.${type}`, intl)} node ${id}`,
        description: <T id="common.deleteDesc" />,
        handle: () => {
          action(id)

          onClose()
        },
      })
    )
  }
  return (
    <Menu
      IconButtonProps={{
        size: 'small',
        sx: type === 'flowNode' ? { position: 'absolute', top: 1, right: 4 } : undefined,
      }}
      IconProps={{ fontSize: 'inherit' }}
    >
      {({ onClose }: any) => [
        <MenuItem key={'copy-' + id} onClick={onNode('copy', onClose)}>
          <ListItemIcon sx={{ fontSize: 18 }}>
            <ContentCopyIcon fontSize="inherit" />
          </ListItemIcon>
          <ListItemText primaryTypographyProps={{ variant: 'button', color: 'secondary' }}>
            <T id="common.copy" />
          </ListItemText>
        </MenuItem>,
        <MenuItem key={'delete-' + id} onClick={onNode('delete', onClose)}>
          <ListItemIcon sx={{ fontSize: 18 }}>
            <DeleteIcon fontSize="inherit" />
          </ListItemIcon>
          <ListItemText primaryTypographyProps={{ variant: 'button', color: 'secondary' }}>
            <T id="common.delete" />
          </ListItemText>
        </MenuItem>,
      ]}
    </Menu>
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
    <IconButton size="small" onClick={onEdgeDelete} title={i18n('common.delete', intl)} aria-label="delete">
      <DeleteIcon fontSize="inherit" />
    </IconButton>
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
            markerEnd: commonMarkerEnd,
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
    const node = nodes.find((n) => n.id === id)!

    // Remove empty node.
    if (!node.data.finished) {
      setNodes(nodes.filter((node) => node.id !== id))
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
      data: {
        finished: false,
      },
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

    return id
  }

  const editNode = (e: React.MouseEvent, { id, data }: Node) => {
    // Prevent editing nodes when resizing.
    //
    // See `GroupNode.tsx` for more details.
    if (
      (e.target as HTMLDivElement).className === 're-resizable' ||
      (e.target as HTMLDivElement).className === ResizableHandleClassName
    ) {
      return
    }

    const workflowNode = store.nodes[data.name]

    formInitialValues.current = workflowNode

    setIdentifier({
      id,
      kind: workflowNode.kind,
      act: workflowNode.action,
    })
    setOpenDrawer(true)
  }

  const addNodeControl = (id: uuid, type: NodeControlProps['type']) => {
    switch (type) {
      case 'groupNode':
        return { nodeControl: <NodeControl id={id} type={type} onDelete={deleteNode} onCopy={copyNode} /> }
      case 'flowNode':
      default:
        return { endIcon: <NodeControl id={id} type={type} onDelete={deleteNode} onCopy={copyNode} /> }
    }
  }

  const updateNode = (values: Record<string, any>) => {
    const { id, ...rest } = values

    setNodes((oldNodes) =>
      oldNodes.map((node) => {
        if (node.id === id) {
          return {
            ...node,
            data: {
              ...node.data,
              finished: true,
              name: values.name,
              ...(node.type === 'flowNode' && {
                children: values.name,
              }),
              ...addNodeControl(id, node.type as NodeControlProps['type']),
            },
          }
        }

        return node
      })
    )
    dispatch(updateWorkflowNode(rest))

    cleanup()
  }

  const copyNode = (id: uuid) => {
    setNodes((oldNodes) => {
      function genNewChildrenNodes(id: uuid, results: Node[], parent?: uuid) {
        const node = oldNodes.find((n) => n.id === id)!
        const newID = uuidv4()

        results.push({
          ...node,
          id: newID,
          ...(!parent && { position: { x: node.position.x, y: node.position.y + node.height! + 100 } }),
          selected: false, // Reset selection.
          data: {
            ...node.data,
            ...addNodeControl(newID, node.type as NodeControlProps['type']),
          },
          ...(parent && {
            parentNode: parent,
            extent: 'parent',
          }),
        })

        // Copy all children nodes.
        if (node.type === 'groupNode') {
          oldNodes
            .filter((node) => node.parentNode === id)
            .forEach((node) => genNewChildrenNodes(node.id, results, newID))
        }
      }

      const newNodes: Node[] = []
      genNewChildrenNodes(id, newNodes)

      return [...oldNodes, ...newNodes]
    })
  }

  const deleteNode = (id: uuid) => {
    setNodes((oldNodes) => {
      function findDeletedNodes(id: uuid, results: Node[]) {
        const node = oldNodes.find((n) => n.id === id)!

        results.push(node)

        if (node.type === 'groupNode') {
          oldNodes.filter((n) => n.parentNode === id).forEach((n) => findDeletedNodes(n.id, results))
        }
      }

      const deletedNodes: Node[] = []
      findDeletedNodes(id, deletedNodes)
      const restNodes = _.differenceBy(oldNodes, deletedNodes, 'id')
      const templates = deletedNodes.map((n) => n.data.name)

      // Remove templates if they are not used by other nodes.
      templates.forEach((template) => {
        if (!restNodes.some((n) => n.data.name === template)) {
          dispatch(removeWorkflowNode(template))
        }
      })

      return restNodes
    })
  }

  const initNode = (item: DropItem, monitor?: DropTargetMonitor, xyCoord?: XYCoord, parent?: uuid) => {
    // If `xyCoord` is `undefined`, the item isn't operated by dragging and dropping.
    if (!monitor?.isOver({ shallow: true }) && !xyCoord) {
      return
    }

    // Add new node into flow.
    const id = addNode(item, monitor, xyCoord, parent)

    // Start generating form.
    setIdentifier({ id, ...item })

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
    setNodes(
      nodes.map((node) => ({
        ...node,
        data: {
          ...node.data,
          finished: true,
          ...addNodeControl(node.id, node.type as NodeControlProps['type']),
        },
      }))
    )
    setEdges(
      edges.map((edge) => ({
        ...edge,
        data: { ...edge.data, tooltipProps: { title: <EdgeControl id={edge.data.id} onDelete={deleteEdge} /> } },
        markerEnd: commonMarkerEnd,
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
