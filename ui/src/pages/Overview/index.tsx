import { Grid, Grow, Paper } from '@material-ui/core'
import React, { useEffect, useRef, useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import { Experiment } from 'api/experiments.type'
import Loading from 'components/Loading'
import PaperTop from 'components/PaperTop'
import StatusPanel from 'components/StatusPanel'
import api from 'api'
import genChaosChart from 'lib/d3/chaosBarChart'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    chaosChart: {
      height: 300,
      margin: theme.spacing(3),
    },
  })
)

export default function Overview() {
  const classes = useStyles()

  const chaosChartRef = useRef<HTMLDivElement>(null)
  const [loading, setLoading] = useState(false)
  const [experiments, setExperiments] = useState<Experiment[] | null>(null)

  const fetchExperiments = () => {
    setLoading(true)

    api.experiments
      .experiments()
      .then(({ data }) => setExperiments(data))
      .catch(console.log)
      .finally(() => setLoading(false))
  }

  useEffect(fetchExperiments, [])

  useEffect(() => {
    if (experiments) {
      const chart = chaosChartRef.current!

      genChaosChart({
        root: chart,
        chaos: Object.entries(
          experiments.reduce((acc, e) => {
            if (acc[e.kind]) {
              acc[e.kind] += 1
            } else {
              acc[e.kind] = 1
            }

            return acc
          }, {} as { [key: string]: number })
        ).map(([k, v]) => ({ kind: k, sum: v })),
      })
    }
  }, [experiments])

  return (
    <>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <Paper variant="outlined">
              <PaperTop title="Total Experiments" />
              <div ref={chaosChartRef} className={classes.chaosChart} />
            </Paper>
          </Grid>

          <Grid item xs={12} md={6}>
            <StatusPanel />
          </Grid>
        </Grid>
      </Grow>

      {loading && <Loading />}
    </>
  )
}
