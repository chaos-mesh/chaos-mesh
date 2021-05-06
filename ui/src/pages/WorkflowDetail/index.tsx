import { Box, Button, CircularProgress, Grow } from '@material-ui/core'
import { useEffect, useRef, useState } from 'react'

import { WorkflowDetail as APIWorkflowDetail } from 'api/workflows.type'
import DeleteOutlinedIcon from '@material-ui/icons/DeleteOutlined'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import T from 'components/T'
import api from 'api'
import { constructWorkflowTopology } from 'lib/cytoscape'
import { makeStyles } from '@material-ui/core/styles'
import { useParams } from 'react-router-dom'

const useStyles = makeStyles((theme) => ({
  root: {
    height: `calc(100vh - 56px - ${theme.spacing(18)})`,
  },
  topology: {
    flex: 1,
  },
}))

const WorkflowDetail = () => {
  const classes = useStyles()
  const { namespace, name } = useParams<any>()

  const [loading, setLoading] = useState(false)
  const [detail, setDetail] = useState<APIWorkflowDetail>()
  const topologyRef = useRef<any>(null)

  const fetchWorkflowDetail = (ns: string, name: string) =>
    api.workflows
      .detail(ns, name)
      .then(({ data }) => setDetail(data))
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

      const { updateElements } = constructWorkflowTopology(topologyRef.current!, detail)

      topologyRef.current = updateElements
    }
  }, [detail])

  return (
    <Grow in={true} style={{ transformOrigin: '0 0 0' }}>
      <Space display="flex" flexDirection="column" className={classes.root} vertical spacing={6}>
        <Space>
          <Button variant="outlined" size="small" startIcon={<DeleteOutlinedIcon />} onClick={() => {}}>
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
          />
          <div ref={topologyRef} style={{ flex: 1 }} />
        </Paper>
      </Space>
    </Grow>
  )
}

export default WorkflowDetail
