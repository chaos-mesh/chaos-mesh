import { Grid, Grow, Paper } from '@material-ui/core'
import React, { useEffect, useRef, useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import { Experiment } from 'api/experiments.type'
import Loading from 'components-mui/Loading'
import PaperTop from 'components-mui/PaperTop'
import { RootState } from 'store'
import StatusPanel from 'components/StatusPanel'
import T from 'components/T'
import api from 'api'
import genChaosChart from 'lib/d3/chaosBarChart'
import { useSelector } from 'react-redux'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    chaosChart: {
      height: 250,
      margin: theme.spacing(3),
    },
  })
)

export default function Overview() {
  const classes = useStyles()

  const { theme } = useSelector((state: RootState) => state.settings)

  const chaosChartRef = useRef<HTMLDivElement>(null)
  const [loading, setLoading] = useState(false)
  const [experiments, setExperiments] = useState<Experiment[] | null>(null)

  const fetchExperiments = () => {
    setLoading(true)

    api.experiments
      .experiments()
      .then(({ data }) => setExperiments(data))
      .catch((error) => {
        console.error(error)

        setExperiments([])
      })
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
          }, {} as Record<string, number>)
        ).map(([k, v]) => ({ kind: k, sum: v })),
        theme,
      })
    }
  }, [experiments, theme])

  return (
    <>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <Paper variant="outlined">
              <PaperTop title={T('overview.totalExperiments')} />
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
