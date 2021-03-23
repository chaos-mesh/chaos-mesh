import { Box, Button, Grid } from '@material-ui/core'
import cytoscape, { workflowStyle } from 'lib/cytoscape'
import { useEffect, useRef } from 'react'

import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import ReplayIcon from '@material-ui/icons/Replay'
import T from 'components/T'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles(() => ({
  topology: {
    height: 300,
  },
}))

const WorkflowDetail = () => {
  const classes = useStyles()

  const topologyRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    cytoscape(topologyRef.current!, {
      elements: {
        nodes: [
          {
            data: {
              id: 'random-pod-failure',
            },
          },
          {
            data: {
              id: 'random-pod-kill',
            },
          },
          {
            data: {
              id: 'network-loss',
              position: {
                x: 100,
                y: 100,
              },
            },
          },
          {
            data: {
              id: 'network-delay',
            },
          },
          {
            data: {
              id: 'random-pod-kill-1',
            },
          },
          {
            data: {
              id: 'random-pod-kill-2',
            },
          },
        ],
        edges: [
          {
            data: {
              id: 'random-pod-failure-to-random-pod-kill',
              source: 'random-pod-failure',
              target: 'random-pod-kill',
            },
          },
          {
            data: {
              id: 'random-pod-kill-to-network-loss',
              source: 'random-pod-kill',
              target: 'network-loss',
            },
          },
          {
            data: {
              id: 'network-loss-to-random-pod-kill-1',
              source: 'network-loss',
              target: 'random-pod-kill-1',
            },
          },
          {
            data: {
              id: 'random-pod-kill-to-network-delay',
              source: 'random-pod-kill',
              target: 'network-delay',
            },
          },
          {
            data: {
              id: 'network-delay-to-random-pod-kill-2',
              source: 'network-delay',
              target: 'random-pod-kill-2',
            },
          },
        ],
      },
      style: workflowStyle,
      layout: {
        name: 'dagre',
        rankDir: 'LR',
        minLen: 9,
      } as any,
    })
  }, [])

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
