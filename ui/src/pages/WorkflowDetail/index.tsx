import { Box, Button, CircularProgress, Grow, Modal } from '@material-ui/core'
import { Confirm, setAlert, setConfirm } from 'slices/globalStatus'
import { useEffect, useRef, useState } from 'react'
import { useHistory, useParams } from 'react-router-dom'
import { useStoreDispatch, useStoreSelector } from 'store'

import { WorkflowDetail as APIWorkflowDetail } from 'api/workflows.type'
import { Ace } from 'ace-builds'
import DeleteOutlinedIcon from '@material-ui/icons/DeleteOutlined'
import { EventHandler } from 'cytoscape'
import NoteOutlinedIcon from '@material-ui/icons/NoteOutlined'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import T from 'components/T'
import api from 'api'
import { constructWorkflowTopology } from 'lib/cytoscape'
import loadable from '@loadable/component'
import { makeStyles } from '@material-ui/core/styles'
import { useIntl } from 'react-intl'
import yaml from 'js-yaml'

const YAMLEditor = loadable(() => import('components/YAMLEditor'))

const useStyles = makeStyles((theme) => ({
  root: {
    height: `calc(100vh - 56px - ${theme.spacing(18)})`,
  },
  topology: {
    flex: 1,
  },
  configPaper: {
    position: 'absolute',
    top: '50%',
    left: '50%',
    width: '50vw',
    height: '90vh',
    transform: 'translate(-50%, -50%)',
    [theme.breakpoints.down('sm')]: {
      width: '90vw',
    },
  },
}))

const WorkflowDetail = () => {
  const classes = useStyles()
  const intl = useIntl()
  const history = useHistory()
  const { namespace, name } = useParams<any>()

  const { theme } = useStoreSelector((state) => state.settings)
  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(false)
  const [detail, setDetail] = useState<APIWorkflowDetail>()
  const [yamlEditor, setYAMLEditor] = useState<Ace.Editor>()
  const [data, setData] = useState<any>()
  const [selected, setSelected] = useState<'workflow' | 'node'>('workflow')
  const modalTitle = selected === 'workflow' ? detail?.name : selected === 'node' ? data.name : ''
  const [configOpen, setConfigOpen] = useState(false)
  const topologyRef = useRef<any>(null)

  const fetchWorkflowDetail = (ns: string, name: string) =>
    api.workflows
      .detail(ns, name)
      .then(({ data }) => {
        // TODO: remove noise in API
        delete data.kube_object.metadata.annotations['kubectl.kubernetes.io/last-applied-configuration']

        setDetail(data)
      })
      .catch(console.error)
      .finally(() => setTimeout(() => setLoading(true), 1000))

  useEffect(() => {
    fetchWorkflowDetail(namespace, name)

    const id = setInterval(() => {
      setLoading(false)

      fetchWorkflowDetail(namespace, name)
    }, 6000)

    return () => clearInterval(id)
  }, [namespace, name])

  useEffect(() => {
    if (detail) {
      const topology = topologyRef.current!

      if (typeof topology === 'function') {
        topology(detail)

        return
      }

      const { updateElements } = constructWorkflowTopology(topologyRef.current!, detail, handleNodeClick)

      topologyRef.current = updateElements
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [detail])

  const onModalOpen = () => setConfigOpen(true)
  const onModalClose = () => setConfigOpen(false)

  const handleSelect = (selected: Confirm) => () => dispatch(setConfirm(selected))

  const handleAction = (action: string, data: { namespace: string; name: string }) => () => {
    let actionFunc: any

    switch (action) {
      case 'delete':
        actionFunc = api.workflows.del

        break
      default:
        actionFunc = null
    }

    const { namespace, name } = data

    if (actionFunc) {
      actionFunc(namespace, name)
        .then(() => {
          dispatch(
            setAlert({
              type: 'success',
              message: intl.formatMessage({ id: `confirm.${action}Successfully` }),
            })
          )

          if (action === 'delete') {
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
      ...detail?.kube_object,
    })
    setSelected('workflow')

    onModalOpen()
  }

  const handleNodeClick: EventHandler = (e) => {
    const node = e.target
    const { id } = node.data()
    const template = detail?.kube_object.spec.templates.find((t: any) => t.name === id)

    setData(template)
    setSelected('node')

    onModalOpen()
  }

  const handleUpdateWorkflow = () => {
    const data = yaml.load(yamlEditor!.getValue())

    api.workflows
      .update(namespace, name, data)
      .then(() => {
        onModalClose()

        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: `common.updateSuccessfully` }),
          })
        )

        fetchWorkflowDetail(namespace, name)
      })
      .catch(console.error)
  }

  return (
    <>
      <Grow in={true} style={{ transformOrigin: '0 0 0' }}>
        <Space display="flex" flexDirection="column" className={classes.root} vertical spacing={6}>
          <Space>
            <Button
              variant="outlined"
              size="small"
              startIcon={<DeleteOutlinedIcon />}
              onClick={handleSelect({
                title: `${intl.formatMessage({ id: 'common.delete' })} ${name}`,
                description: intl.formatMessage({ id: 'workflows.deleteDesc' }),
                handle: handleAction('delete', { namespace, name }),
              })}
            >
              {T('common.delete')}
            </Button>
          </Space>
          <Paper className={classes.topology} boxProps={{ display: 'flex', flexDirection: 'column' }}>
            <PaperTop
              title={
                <Space spacing={1.5} display="flex" alignItems="center">
                  <Box>{T('workflow.topology')}</Box>
                  {loading && <CircularProgress size={15} />}
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
      </Grow>

      <Modal open={configOpen} onClose={onModalClose}>
        <div>
          <Paper className={classes.configPaper} padding={0}>
            {detail && configOpen && (
              <Box display="flex" flexDirection="column" height="100%">
                <Box px={3} pt={3}>
                  <PaperTop title={modalTitle}>
                    <Button variant="contained" color="primary" size="small" onClick={handleUpdateWorkflow}>
                      {T('common.update')}
                    </Button>
                  </PaperTop>
                </Box>
                <Box display="flex" flex={1}>
                  <YAMLEditor theme={theme} data={yaml.dump(data)} mountEditor={setYAMLEditor} />
                </Box>
              </Box>
            )}
          </Paper>
        </div>
      </Modal>
    </>
  )
}

export default WorkflowDetail
