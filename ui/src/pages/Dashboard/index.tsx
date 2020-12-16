import { Box, Grid, Grow, Typography } from '@material-ui/core'
import React, { useEffect, useRef } from 'react'

import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Predefined from './Predefined'
import { RootState } from 'store'
import T from 'components/T'
import Timeline from 'components/Timeline'
import TotalExperiments from './TotalExperiments'
import Welcome from './Welcome'
import genChaosStatePieChart from 'lib/d3/chaosStatePieChart'
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
  totalExperiments: {
    display: 'flex',
    alignItems: 'center',
    [theme.breakpoints.down('sm')]: {
      alignItems: 'unset',
    },
  },
}))

export default function Dashboard() {
  const classes = useStyles()

  const intl = useIntl()

  const { theme } = useSelector((state: RootState) => state.settings)
  const { stateOfExperiments } = useSelector((state: RootState) => state.experiments)
  const dispatch = useStoreDispatch()

  const chaosStatePieChartRef = useRef<any>(null)

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
          <Grid item md={12} lg={3}>
            <Welcome />
          </Grid>

          <Grid item md={12} lg={6}>
            <Paper>
              <PaperTop title={T('dashboard.totalExperiments')} />

              <Box className={classes.totalExperiments} height={300} m={3} overflow="scroll">
                <TotalExperiments />
              </Box>
            </Paper>
          </Grid>

          <Grid item xs={12} md={12} lg={3}>
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

          <Grid item xs={12} md={12} lg={9}>
            <Paper>
              <PaperTop title={T('dashboard.predefined')} />

              <Box height={150} mx={3}>
                <Typography>{T('dashboard.predefinedDesc')}</Typography>
                <Box pt={6}>
                  <Predefined />
                </Box>
              </Box>
            </Paper>
          </Grid>

          <Grid item xs={12} md={12} lg={9}>
            <Paper>
              <PaperTop title={T('common.timeline')} />

              <Timeline className={classes.container} />
            </Paper>
          </Grid>
        </Grid>
      </Grow>
    </>
  )
}
