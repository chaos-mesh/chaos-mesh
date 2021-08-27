import { Box, Button, Grow, Modal, useTheme } from '@material-ui/core'
import { Confirm, setAlert, setConfirm } from 'slices/globalStatus'
import { useEffect, useRef, useState } from 'react'
import { useHistory, useParams } from 'react-router-dom'

import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import { EventHandler } from 'cytoscape'
import NodeConfiguration from 'components/ObjectConfiguration/Node'
import NoteOutlinedIcon from '@material-ui/icons/NoteOutlined'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import T from 'components/T'
import { WorkflowSingle } from 'api/workflows.type'
import api from 'api'
import { constructWorkflowTopology } from 'lib/cytoscape'
import loadable from '@loadable/component'
import { makeStyles } from '@material-ui/styles'
import { useIntervalFetch } from 'lib/hooks'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'
import yaml from 'js-yaml'

const YAMLEditor = loadable(() => import('components/YAMLEditor'))

const useStyles = makeStyles((theme) => ({
  root: {
    height: `calc(100vh - 56px - ${theme.spacing(18)})`,
  },
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

const Single = () => {
  const classes = useStyles()
  const intl = useIntl()
  const history = useHistory()
  const theme = useTheme()
  const { uuid } = useParams<{ uuid: uuid }>()

  const dispatch = useStoreDispatch()

  const [single, setSingle] = useState<WorkflowSingle>()
  const [data, setData] = useState<any>()
  const [selected, setSelected] = useState<'workflow' | 'node'>('workflow')
  const modalTitle = selected === 'workflow' ? single?.name : selected === 'node' ? data.name : ''
  const [configOpen, setConfigOpen] = useState(false)
  const topologyRef = useRef<any>(null)

  const fetchWorkflowSingle = (intervalID?: number) =>
    api.workflows
      .single(uuid)
      .then(({ data }) => {
        // TODO: remove noise in API
        data.kube_object.metadata.annotations &&
          delete data.kube_object.metadata.annotations['kubectl.kubernetes.io/last-applied-configuration']

        setSingle(data)

        // Clear interval after workflow succeed
        if (data.status === 'finished') {
          clearInterval(intervalID)
        }
      })
      .catch(console.error)

  useIntervalFetch(fetchWorkflowSingle)

  useEffect(() => {
    if (single) {
      const topology = topologyRef.current!

      if (typeof topology === 'function') {
        topology(single)

        return
      }

      const { updateElements } = constructWorkflowTopology(topologyRef.current!, single, theme, handleNodeClick)

      topologyRef.current = updateElements
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [single])

  const onModalOpen = () => setConfigOpen(true)
  const onModalClose = () => setConfigOpen(false)

  const handleSelect = (selected: Confirm) => () => dispatch(setConfirm(selected))

  const handleAction = (action: string) => () => {
    let actionFunc: any

    switch (action) {
      case 'archive':
        actionFunc = api.workflows.del

        break
      default:
        actionFunc = null
    }

    if (actionFunc) {
      actionFunc(uuid)
        .then(() => {
          dispatch(
            setAlert({
              type: 'success',
              message: T(`confirm.success.${action}`, intl),
            })
          )

          if (action === 'archive') {
            history.push('/workflows')
          }
        })
        .catch(console.error)
    }
  }

  const handleOpenConfig = () => {
    setData({
      apiVersion: 'chaos-mesh.org/v1alpha1',
      kind: 'Workflow',
      ...single?.kube_object,
    })
    setSelected('workflow')

    onModalOpen()
  }

  const handleNodeClick: EventHandler = (e) => {
    const node = e.target
    const { template: nodeTemplate } = node.data()
    const template = single?.kube_object.spec.templates.find((t: any) => t.name === nodeTemplate)

    setData(template)
    setSelected('node')

    onModalOpen()
  }

  const handleUpdateWorkflow = (data: any) => {
    if (selected === 'node') {
      const kubeObject = single?.kube_object
      kubeObject.spec.templates = kubeObject.spec.templates.map((t: any) => {
        if (t.name === (data as any).name) {
          return data
        }

        return t
      })

      data = {
        apiVersion: 'chaos-mesh.org/v1alpha1',
        kind: 'Workflow',
        ...kubeObject,
      }
    }

    api.workflows
      .update(uuid, data)
      .then(() => {
        onModalClose()

        dispatch(
          setAlert({
            type: 'success',
            message: T(`confirm.success.update`, intl),
          })
        )

        fetchWorkflowSingle()
      })
      .catch(console.error)
  }

  return (
    <>
      <Grow in={true} style={{ transformOrigin: '0 0 0' }}>
        <div>
          <Space spacing={6} className={classes.root}>
            <Space direction="row">
              <Button
                variant="outlined"
                size="small"
                startIcon={<ArchiveOutlinedIcon />}
                onClick={handleSelect({
                  title: `${T('archives.single', intl)} ${single?.name}`,
                  description: T('workflows.deleteDesc', intl),
                  handle: handleAction('archive'),
                })}
              >
                {T('archives.single')}
              </Button>
            </Space>
            <Paper sx={{ display: 'flex', flexDirection: 'column', flex: 1 }}>
              <PaperTop
                title={
                  <Space spacing={1.5} alignItems="center">
                    <Box>{T('workflow.topology')}</Box>
                  </Space>
                }
              >
                <Button
                  variant="outlined"
                  size="small"
                  color="primary"
                  startIcon={<NoteOutlinedIcon />}
                  onClick={handleOpenConfig}
                >
                  {T('common.configuration')}
                </Button>
              </PaperTop>
              <div ref={topologyRef} style={{ flex: 1 }} />
            </Paper>
          </Space>
        </div>
      </Grow>

      <Modal open={configOpen} onClose={onModalClose}>
        <div>
          <Paper
            className={classes.configPaper}
            sx={{ width: selected === 'workflow' ? '50vw' : selected === 'node' ? '70vw' : '50vw' }}
          >
            {single && configOpen && (
              <Space display="flex" flexDirection="column" height="100%">
                <PaperTop title={modalTitle} boxProps={{ p: 4.5, pb: 0 }} />
                <Box display="flex" flex={1}>
                  {selected === 'node' && (
                    <Box width="50%">
                      <NodeConfiguration template={data} />
                    </Box>
                  )}
                  <YAMLEditor name={modalTitle} data={yaml.dump(data)} onUpdate={handleUpdateWorkflow} download />
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
