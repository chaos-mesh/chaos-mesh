import { Box, Grid, Grow, Typography } from '@material-ui/core'
import React, { useEffect, useRef, useState } from 'react'

import { Event } from 'api/events.type'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import { RootState } from 'store'
import T from 'components/T'
import TotalExperiments from './TotalExperiments'
import Welcome from './Welcome'
import api from 'api'
import genChaosStatePieChart from 'lib/d3/chaosStatePieChart'
import genEventsChart from 'lib/d3/eventsChart'
import { getStateofExperiments } from 'slices/experiments'
import { makeStyles } from '@material-ui/core/styles'
import { useIntl } from 'react-intl'
import { useSelector } from 'react-redux'
import { useStoreDispatch } from 'store'

const useStyles = makeStyles((theme) => ({
  container: {
    height: 300,
    margin: theme.spacing(3),
  },
  notFound: {
    position: 'absolute',
    top: '50%',
    left: '50%',
    transform: 'translate3d(-50%, -50%, 0)',
  },
}))

export default function Dashboard() {
  const classes = useStyles()

  const intl = useIntl()

  const { theme } = useSelector((state: RootState) => state.settings)
  const { stateOfExperiments } = useSelector((state: RootState) => state.experiments)
  const dispatch = useStoreDispatch()

  const chartRef = useRef<HTMLDivElement>(null)
  const chaosStatePieChartRef = useRef<any>(null)

  const [events, setEvents] = useState<Event[] | null>(null)

  const fetchEvents = () => {
    api.events
      .events()
      .then(({ data }) => setEvents(data))
      .catch(console.error)
  }

  useEffect(fetchEvents, [])

  useEffect(() => {
    if (events && events.length) {
      const chart = chartRef.current!

      genEventsChart({
        root: chart,
        events,
        intl,
        theme,
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [events])

  useEffect(() => {
    dispatch(getStateofExperiments())

    const id = setInterval(() => dispatch(getStateofExperiments()), 15000)

    return () => clearInterval(id)
  }, [dispatch])

  useEffect(() => {
    if (typeof chaosStatePieChartRef.current === 'function') {
      chaosStatePieChartRef.current(stateOfExperiments)

      return
    }

    const update = genChaosStatePieChart({
      root: chaosStatePieChartRef.current,
      chaosStatus: stateOfExperiments,
      intl,
      theme,
    })
    chaosStatePieChartRef.current = update
  }, [stateOfExperiments, intl, theme])

  return (
    <>
      <Grow in={true} style={{ transformOrigin: '0 0 0' }}>
        <Grid container spacing={3}>
          <Grid item xs={12} md={3}>
            <Welcome />
          </Grid>

          <Grid item xs={12} md={6}>
            <Paper>
              <PaperTop title={T('dashboard.totalExperiments')} />
              <Box height={300} m={3} overflow="scroll">
                <TotalExperiments />
              </Box>
            </Paper>
          </Grid>

          <Grid item xs={12} md={3}>
            <Paper style={{ position: 'relative' }}>
              <PaperTop title={T('dashboard.totalState')} />
              <div ref={chaosStatePieChartRef} className={classes.container} />
              {Object.values(stateOfExperiments).filter((d) => d !== 0).length === 0 && (
                <Typography className={classes.notFound} align="center">
                  {T('experiments.noExperimentsFound')}
                </Typography>
              )}
            </Paper>
          </Grid>

          <Grid item xs={12} md={9}>
            <Paper>
              <PaperTop title={T('dashboard.predefined')} />
              <Box height={150} m={3}></Box>
            </Paper>
          </Grid>

          <Grid item xs={12} md={9}>
            <Paper>
              <PaperTop title={T('common.timeline')} />
              <div ref={chartRef} className={classes.container} />
            </Paper>
          </Grid>
        </Grid>
      </Grow>
    </>
  )
}
