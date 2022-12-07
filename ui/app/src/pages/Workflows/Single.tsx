/*
 * Copyright 2021 Chaos Mesh Authors.
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
import loadable from '@loadable/component'
import ArchiveOutlinedIcon from '@mui/icons-material/ArchiveOutlined'
import { Box, Button, Grid, Grow, Modal, useTheme } from '@mui/material'
import { makeStyles } from '@mui/styles'
import { EventHandler } from 'cytoscape'
import yaml from 'js-yaml'
import { useDeleteWorkflowsUid, useGetEventsWorkflowUid, useGetWorkflowsUid } from 'openapi'
import { CoreWorkflowDetail } from 'openapi/index.schemas'
import { useEffect, useRef, useState } from 'react'
import { useIntl } from 'react-intl'
import { useNavigate, useParams } from 'react-router-dom'

import Paper from '@ui/mui-extends/esm/Paper'
import PaperTop from '@ui/mui-extends/esm/PaperTop'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch } from 'store'

import { Confirm, setAlert, setConfirm } from 'slices/globalStatus'

import EventsTimeline from 'components/EventsTimeline'
import Helmet from 'components/Helmet'
import NodeConfiguration from 'components/ObjectConfiguration/Node'
import i18n from 'components/T'

import { constructWorkflowTopology } from 'lib/cytoscape'

const YAMLEditor = loadable(() => import('components/YAMLEditor'))

const useStyles = makeStyles((theme) => ({
  root: {},
  configPaper: {
    position: 'absolute',
    top: '50%',
    left: '50%',
    width: '75vw',
    height: '90vh',
    padding: 0,
    transform: 'translate(-50%, -50%)',
    [theme.breakpoints.down('lg')]: {
      width: '90vw',
    },
  },
}))

function transformWorkflow(data: CoreWorkflowDetail) {
  // TODO: remove noise in API
  data.kube_object!.metadata!.annotations &&
    delete data.kube_object!.metadata!.annotations['kubectl.kubernetes.io/last-applied-configuration']

  return data
}

const Single = () => {
  const classes = useStyles()
  const intl = useIntl()
  const navigate = useNavigate()
  const theme = useTheme()
  const { uuid } = useParams()

  const dispatch = useStoreDispatch()

  const [data, setData] = useState<any>()
  const [selected, setSelected] = useState<'workflow' | 'node'>('workflow')
  const [configOpen, setConfigOpen] = useState(false)
  const topologyRef = useRef<any>(null)

  const { data: workflow } = useGetWorkflowsUid(uuid!, {
    query: {
      select: transformWorkflow,
    },
  })
  const modalTitle = selected === 'workflow' ? workflow?.name : selected === 'node' ? data.name : ''
  const { data: events } = useGetEventsWorkflowUid(uuid!, { limit: 999 })
  const { mutateAsync: deleteWorkflows } = useDeleteWorkflowsUid()

  useEffect(() => {
    if (workflow) {
      const topology = topologyRef.current!

      if (typeof topology === 'function') {
        topology(workflow)

        return
      }

      const { updateElements } = constructWorkflowTopology(
        topologyRef.current!,
        workflow as any,
        theme,
        handleNodeClick
      )

      topologyRef.current = updateElements
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [workflow])

  const onModalOpen = () => setConfigOpen(true)
  const onModalClose = () => setConfigOpen(false)

  const handleSelect = (selected: Confirm) => () => dispatch(setConfirm(selected))

  const handleAction = (action: string) => () => {
    let actionFunc

    switch (action) {
      case 'archive':
        actionFunc = deleteWorkflows

        break
      default:
        break
    }

    if (actionFunc) {
      actionFunc({ uid: uuid! })
        .then(() => {
          dispatch(
            setAlert({
              type: 'success',
              message: i18n(`confirm.success.${action}`, intl),
            })
          )

          if (action === 'archive') {
            navigate('/workflows')
          }
        })
        .catch(console.error)
    }
  }

  const handleNodeClick: EventHandler = (e) => {
    const node = e.target
    const { template: nodeTemplate } = node.data()
    const template = (workflow?.kube_object!.spec as any).templates.find((t: any) => t.name === nodeTemplate)

    setData(template)
    setSelected('node')

    onModalOpen()
  }

  return (
    <>
      <Grow in={true} style={{ transformOrigin: '0 0 0' }}>
        <div style={{ height: '100%' }}>
          {workflow && <Helmet title={`Workflow ${workflow.name}`} />}
          <Space spacing={6} className={classes.root}>
            <Space direction="row">
              <Button
                variant="outlined"
                size="small"
                startIcon={<ArchiveOutlinedIcon />}
                onClick={handleSelect({
                  title: `${i18n('archives.single', intl)} ${workflow?.name}`,
                  description: i18n('workflows.deleteDesc', intl),
                  handle: handleAction('archive'),
                })}
              >
                {i18n('archives.single')}
              </Button>
            </Space>
            <Paper sx={{ display: 'flex', flexDirection: 'column', height: 450 }}>
              <PaperTop title={i18n('workflow.topology')} />
              <div ref={topologyRef} style={{ flex: 1 }} />
            </Paper>

            <Grid container>
              <Grid item xs={12} lg={6} sx={{ pr: 3 }}>
                <Paper sx={{ display: 'flex', flexDirection: 'column', height: 600 }}>
                  <PaperTop title={i18n('events.title')} boxProps={{ mb: 3 }} />
                  <Box flex={1} overflow="scroll">
                    {events && <EventsTimeline events={events} />}
                  </Box>
                </Paper>
              </Grid>
              <Grid item xs={12} lg={6} sx={{ pl: 3 }}>
                <Paper sx={{ height: 600, p: 0 }}>
                  {workflow && (
                    <Space display="flex" flexDirection="column" height="100%">
                      <PaperTop title={i18n('common.definition')} boxProps={{ p: 4.5, pb: 0 }} />
                      <Box flex={1}>
                        <YAMLEditor
                          name={workflow.name}
                          data={yaml.dump({
                            apiVersion: 'chaos-mesh.org/v1alpha1',
                            kind: 'Workflow',
                            ...workflow.kube_object,
                          })}
                          download
                        />
                      </Box>
                    </Space>
                  )}
                </Paper>
              </Grid>
            </Grid>
          </Space>
        </div>
      </Grow>

      <Modal open={configOpen} onClose={onModalClose}>
        <div>
          <Paper
            className={classes.configPaper}
            sx={{ width: selected === 'workflow' ? '50vw' : selected === 'node' ? '70vw' : '50vw' }}
          >
            {workflow && configOpen && (
              <Space display="flex" flexDirection="column" height="100%">
                <PaperTop title={modalTitle} boxProps={{ p: 4.5, pb: 0 }} />
                <Box display="flex" flex={1}>
                  {selected === 'node' && (
                    <Box width="50%">
                      <NodeConfiguration template={data} />
                    </Box>
                  )}
                  <YAMLEditor name={modalTitle} data={yaml.dump(data)} />
                </Box>
              </Space>
            )}
          </Paper>
        </div>
      </Modal>
    </>
  )
}

export default Single
