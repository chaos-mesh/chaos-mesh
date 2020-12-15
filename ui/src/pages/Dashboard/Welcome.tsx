import { Box, Button, Grid, Typography } from '@material-ui/core'
import React, { useState } from 'react'
import Tour, { ReactourStep } from 'reactour'
import { makeStyles, useTheme } from '@material-ui/core/styles'

import AddIcon from '@material-ui/icons/Add'
import ArrowBackOutlinedIcon from '@material-ui/icons/ArrowBackOutlined'
import ArrowForwardOutlinedIcon from '@material-ui/icons/ArrowForwardOutlined'
import { Link } from 'react-router-dom'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import T from 'components/T'

const useStyles = makeStyles((theme) => ({
  container: {
    position: 'relative',
    height: `calc(94px + ${theme.spacing(5.25)})`,
    margin: theme.spacing(3),
    marginTop: 0,
  },
}))

const Welcome = () => {
  const classes = useStyles()
  const theme = useTheme()

  const steps: ReactourStep[] = [
    {
      selector: '.sidebar-dashboard',
      content: T('dashboard.tutorial.step1'),
    },
    {
      selector: '.sidebar-experiments',
      content: T('dashboard.tutorial.step2'),
    },
    {
      selector: '.sidebar-events',
      content: T('dashboard.tutorial.step3'),
    },
    {
      selector: '.sidebar-archives',
      content: T('dashboard.tutorial.step4'),
    },
    {
      selector: '.dashboard-new-experiment',
      content: T('dashboard.tutorial.step5'),
    },
    {
      selector: '.nav-new-experiment',
      content: T('dashboard.tutorial.step6'),
    },
    {
      selector: '.nav-search',
      content: T('dashboard.tutorial.step7'),
    },
    {
      selector: '.nav-namespace',
      content: T('dashboard.tutorial.step8'),
    },
    {
      selector: '.dashboard-tutorial',
      content: T('dashboard.tutorial.step9'),
    },
  ].map((d) => ({
    ...d,
    style: {
      background: theme.palette.background.default,
    },
  }))

  const [isTourOpen, setIsTourOpen] = useState(false)

  return (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Paper>
          <PaperTop title={T('dashboard.welcome')} />
          <Box className={classes.container}>
            <Typography>{T('dashboard.welcomeDesc')}</Typography>

            <Box position="absolute" bottom={0} width="100%">
              <Button
                className="dashboard-tutorial"
                variant="contained"
                color="primary"
                fullWidth
                onClick={() => setIsTourOpen(true)}
              >
                {T('common.tutorial')}
              </Button>
            </Box>

            <Tour
              steps={steps}
              isOpen={isTourOpen}
              onRequestClose={() => setIsTourOpen(false)}
              accentColor={theme.palette.primary.main}
              rounded={theme.shape.borderRadius}
              prevButton={<ArrowBackOutlinedIcon />}
              nextButton={<ArrowForwardOutlinedIcon />}
              showCloseButton={false}
            />
          </Box>
        </Paper>
      </Grid>
      <Grid item xs={12}>
        <Paper>
          <PaperTop title={T('dashboard.veteran')} />
          <Box className={classes.container}>
            <Typography>{T('dashboard.veteranDesc')}</Typography>

            <Box position="absolute" bottom={0} width="100%">
              <Button
                className="dashboard-new-experiment"
                component={Link}
                to="/newExperiment"
                variant="contained"
                color="primary"
                fullWidth
                startIcon={<AddIcon />}
              >
                {T('newE.title')}
              </Button>
            </Box>
          </Box>
        </Paper>
      </Grid>
    </Grid>
  )
}

export default Welcome
