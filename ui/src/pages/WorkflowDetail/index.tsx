import { Box, Button, Grid } from '@material-ui/core'
import { useEffect, useRef, useState } from 'react'

import { WorkflowDetail as APIWorkflowDetail } from 'api/workflows.type'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import ReplayIcon from '@material-ui/icons/Replay'
import T from 'components/T'
import api from 'api'
import { constructWorkflowTopology } from 'lib/cytoscape'
import { makeStyles } from '@material-ui/core/styles'
import { useParams } from 'react-router-dom'

const useStyles = makeStyles(() => ({
  topology: {
    height: 300,
  },
}))

const WorkflowDetail = () => {
  const classes = useStyles()
  const { namespace, name } = useParams<any>()

  const [detail, setDetail] = useState<APIWorkflowDetail>()
  const topologyRef = useRef<HTMLDivElement>(null)

  const fetchWorkflowDetail = (ns: string, name: string) =>
    api.workflows
      .detail(ns, name)
      .then(({ data }) => setDetail(data))
      .catch(console.error)

  useEffect(() => {
    fetchWorkflowDetail(namespace, name)
  }, [namespace, name])

  useEffect(() => {
    if (detail) {
      constructWorkflowTopology(topologyRef.current!, detail)
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
            <PaperTop title={T('workflow.topology')} />
            <div className={classes.topology} ref={topologyRef}></div>
          </Paper>
        </Grid>
      </Grid>
    </>
  )
}

export default WorkflowDetail
