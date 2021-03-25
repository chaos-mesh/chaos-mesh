import { Box, Button, CircularProgress, Grid } from '@material-ui/core'
import { useEffect, useRef, useState } from 'react'

import { WorkflowDetail as APIWorkflowDetail } from 'api/workflows.type'
import CheckCircleOutlineIcon from '@material-ui/icons/CheckCircleOutline'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import ReplayIcon from '@material-ui/icons/Replay'
import Space from 'components-mui/Space'
import T from 'components/T'
import api from 'api'
import { constructWorkflowTopology } from 'lib/cytoscape'
import { makeStyles } from '@material-ui/core/styles'
import { useParams } from 'react-router-dom'

const useStyles = makeStyles((theme) => ({
  topology: {
    height: 300,
  },
  success: {
    color: theme.palette.success.main,
  },
}))

const WorkflowDetail = () => {
  const classes = useStyles()
  const { namespace, name } = useParams<any>()

  const [detail, setDetail] = useState<APIWorkflowDetail>()
  const [loading, setLoading] = useState(false)
  const topologyRef = useRef<any>(null)

  const fetchWorkflowDetail = (ns: string, name: string) =>
    api.workflows
      .detail(ns, name)
      .then(({ data }) => setDetail(data))
      .catch(console.error)

  useEffect(() => {
    fetchWorkflowDetail(namespace, name)

    const id = setInterval(() => {
      setLoading(false)
      fetchWorkflowDetail(namespace, name)
      setTimeout(() => setLoading(true), 1000)
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
    <>
      <Box mb={6}>
        <Button variant="outlined" startIcon={<ReplayIcon />} onClick={() => {}}>
          {T('workflow.rerun')}
        </Button>
      </Box>
      <Grid container>
        <Grid item xs={12} md={8}>
          <Paper>
            <PaperTop
              title={
                <Space spacing={1.5} display="flex" alignItems="center">
                  <Box>{T('workflow.topology')}</Box>
                  {loading ? (
                    <CircularProgress size={15} />
                  ) : (
                    <CheckCircleOutlineIcon className={classes.success} style={{ width: 20, height: 20 }} />
                  )}
                </Space>
              }
            />
            <div className={classes.topology} ref={topologyRef}></div>
          </Paper>
        </Grid>
      </Grid>
    </>
  )
}

export default WorkflowDetail
