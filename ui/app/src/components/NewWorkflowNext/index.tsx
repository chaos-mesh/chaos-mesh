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
import { Badge, Box, Button, Divider, Grow, Typography } from '@mui/material'
import { useEffect, useRef, useState } from 'react'
import type { ReactFlowInstance } from 'react-flow-renderer'

import Paper from '@ui/mui-extends/esm/Paper'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch, useStoreSelector } from 'store'

import { LoadRecentlyUsedExperiments } from 'slices/workflows'

import FunctionalNodesElements from './Elements/FunctionalNodes'
import KubernetesElements from './Elements/Kubernetes'
import PhysicalNodesElements from './Elements/PhysicalNodes'
import SubmitWorkflow from './SubmitWorkflow'
import Whiteboard from './Whiteboard'
import { flowToWorkflow } from './utils/convert'

export default function NewWorkflow() {
  const [openSubmitDialog, setOpenSubmitDialog] = useState(false)
  const [workflow, setWorkflow] = useState('')

  const { nodes, recentUse } = useStoreSelector((state) => state.workflows)
  const dispatch = useStoreDispatch()

  useEffect(() => {
    dispatch(LoadRecentlyUsedExperiments())
  }, [dispatch])

  const flowRef = useRef<ReactFlowInstance>()

  const handleClickElement = (kind: string, act?: string) => {
    ;(flowRef.current as any).initNode({ kind, act }, undefined, { x: 50, y: 50 }) // TODO: calculate the appropriate coordinates automatically
  }

  const onFinishWorkflow = () => {
    const nds = flowRef.current?.getNodes()!
    const origin = nds.find((n) => n.data.origin)!
    const eds = flowRef.current?.getEdges()!

    const workflow = flowToWorkflow(nodes[origin.id], nodes, eds)

    setWorkflow(workflow)
    setOpenSubmitDialog(true)
  }

  return (
    <Grow in={true} style={{ transformOrigin: '0 0 0' }}>
      <div style={{ height: '100%' }}>
        <Space sx={{ height: '100%' }}>
          <Box display="flex" justifyContent="space-between" alignItems="center">
            <Box>
              <Badge badgeContent="Preview" color="primary">
                <Typography variant="h5" component="h1" fontWeight="bold">
                  New Workflow
                </Typography>
              </Badge>
              <Typography variant="body2">Use flowchart to create a new workflow.</Typography>
            </Box>
            {Object.keys(nodes).length > 0 && (
              <Button variant="contained" size="small" onClick={onFinishWorkflow}>
                Submit Workflow
              </Button>
            )}
          </Box>
          <Paper sx={{ display: 'flex', flex: 1 }}>
            <Space sx={{ width: 256, pr: 4, borderRight: (theme) => `1px solid ${theme.palette.divider}` }}>
              <Typography variant="h6" component="div" fontWeight="bold">
                Elements
              </Typography>
              {recentUse.length > 0 && (
                <Box>
                  <Typography fontWeight="medium">Recently Used</Typography>
                  <Typography variant="body2" color="secondary" fontSize={12}>
                    Recently used experiments
                  </Typography>
                </Box>
              )}
              <Box>
                <Typography fontWeight="medium">Functional Nodes</Typography>
                <Typography variant="body2" color="secondary" fontSize={12}>
                  Drag or click items below into the board to create a functional node.
                </Typography>
              </Box>
              <FunctionalNodesElements onElementClick={handleClickElement} />
              <Box>
                <Typography fontWeight="medium">Kubernetes</Typography>
                <Typography variant="body2" color="secondary" fontSize={12}>
                  Drag or click items below into the board to create a Chaos in Kubernetes.
                </Typography>
              </Box>
              <Box sx={{ height: 450, overflowY: 'auto' }}>
                <KubernetesElements onElementClick={handleClickElement} />
              </Box>
              <Divider />
              <Box>
                <Typography fontWeight="medium">Physical Nodes</Typography>
                <Typography variant="body2" color="secondary" fontSize={12}>
                  Drag or click items below into the board to create a Chaos in Physical Nodes.
                </Typography>
              </Box>
              <Box sx={{ height: 450, overflowY: 'auto' }}>
                <PhysicalNodesElements onElementClick={handleClickElement} />
              </Box>
            </Space>
            <Space sx={{ flex: 1, px: 4 }}>
              <Typography variant="h6" component="div" fontWeight="bold">
                Pipeline Board
              </Typography>
              <Box flex={1}>
                <Whiteboard flowRef={flowRef} />
              </Box>
            </Space>
          </Paper>
        </Space>
        {workflow && <SubmitWorkflow open={openSubmitDialog} setOpen={setOpenSubmitDialog} workflow={workflow} />}
      </div>
    </Grow>
  )
}
