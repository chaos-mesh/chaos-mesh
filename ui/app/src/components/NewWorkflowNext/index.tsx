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
import TabPanelUnstyled from '@mui/base/TabPanelUnstyled'
import TabUnstyled from '@mui/base/TabUnstyled'
import TabsListUnstyled from '@mui/base/TabsListUnstyled'
import TabsUnstyled from '@mui/base/TabsUnstyled'
import { Badge, Box, Button, Grow, Typography } from '@mui/material'
import { styled } from '@mui/material/styles'
import _ from 'lodash'
import { useEffect, useRef, useState } from 'react'
import type { ReactFlowInstance } from 'react-flow-renderer'

import Paper from '@ui/mui-extends/esm/Paper'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch, useStoreSelector } from 'store'

import { loadRecentlyUsedExperiments } from 'slices/workflows'

import YAML from 'components/YAML'

import FunctionalNodesElements from './Elements/FunctionalNodes'
import KubernetesElements from './Elements/Kubernetes'
import PhysicalNodesElements from './Elements/PhysicalNodes'
import SubmitWorkflow from './SubmitWorkflow'
import Whiteboard from './Whiteboard'
import { flowToWorkflow } from './utils/convert'

const Tabs = styled(TabsUnstyled)`
  display: flex;
  flex-direction: column;
`
const TabsList = styled(TabsListUnstyled)`
  display: flex;
  height: 36px;
`
const Tab = styled(TabUnstyled)(
  ({ theme }) => `
  flex: 1;
  padding: 8px 12px;
  background-color: transparent;
  color: ${theme.palette.onSurfaceVariant.main};
  font-family: "Roboto";
  font-weight: 500;
  border: 1px solid ${theme.palette.outline.main};
  transition: all 0.3s ease;
  cursor: pointer;

  &:hover,
  &.Mui-selected {
    background-color: ${theme.palette.secondaryContainer.main};
    color: ${theme.palette.onSecondaryContainer.main};
  }

  &:first-child {
    border-top-left-radius: 4px;
    border-bottom-left-radius: 4px;
  }

  &:not(:first-child) {
    margin-left: -1px;
  }

  &:last-child {
    border-top-right-radius: 4px;
    border-bottom-right-radius: 4px;
  }
  `
)
const TabPanel = styled(TabPanelUnstyled)`
  flex-grow: 1;
  flex-basis: 0;
  overflow-y: auto;
`

export default function NewWorkflow() {
  const [openSubmitDialog, setOpenSubmitDialog] = useState(false)
  const [workflow, setWorkflow] = useState('')

  const { nodes, recentUse } = useStoreSelector((state) => state.workflows)
  const dispatch = useStoreDispatch()

  useEffect(() => {
    dispatch(loadRecentlyUsedExperiments())
  }, [dispatch])

  const flowRef = useRef<ReactFlowInstance>()

  const handleClickElement = (kind: string, act?: string) => {
    ;(flowRef.current as any).initNode({ kind, act }, undefined, { x: 100, y: 100 }) // TODO: calculate the appropriate coordinates automatically
  }

  const handleImportWorkflow = (workflow: string) => {
    ;(flowRef.current as any).importWorkflow(workflow)
  }

  const onFinishWorkflow = () => {
    const nds = flowRef.current?.getNodes()!
    const eds = flowRef.current?.getEdges()!

    const workflow = flowToWorkflow(nds, eds, nodes)

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
            <Space direction="row">
              <YAML callback={handleImportWorkflow}>Import Workflow</YAML>
              {!_.isEmpty(nodes) && (
                <Button variant="contained" size="small" onClick={onFinishWorkflow}>
                  Submit Workflow
                </Button>
              )}
            </Space>
          </Box>
          <Paper sx={{ display: 'flex', flex: 1 }}>
            <Space sx={{ width: 300, pr: 4, borderRight: (theme) => `1px solid ${theme.palette.divider}` }}>
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
                <Typography fontWeight="medium">Chaos Nodes</Typography>
                <Typography variant="body2" color="secondary" fontSize={12}>
                  Drag or click items below into the board to create a Chaos node.
                </Typography>
              </Box>
              <Tabs sx={{ flex: 1 }} defaultValue={0}>
                <TabsList sx={{ mb: 3 }}>
                  <Tab>Kubernetes</Tab>
                  <Tab>Hosts</Tab>
                </TabsList>
                <TabPanel value={0}>
                  <KubernetesElements onElementClick={handleClickElement} />
                </TabPanel>
                <TabPanel value={1}>
                  <PhysicalNodesElements onElementClick={handleClickElement} />
                </TabPanel>
              </Tabs>
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
