import React, { useState } from 'react'
import Tour, { ReactourStep } from 'reactour'
import { makeStyles, useTheme } from '@material-ui/core/styles'

import AddIcon from '@material-ui/icons/Add'
import ArrowBackOutlinedIcon from '@material-ui/icons/ArrowBackOutlined'
import ArrowForwardOutlinedIcon from '@material-ui/icons/ArrowForwardOutlined'
import { Button } from '@material-ui/core'
import { Link } from 'react-router-dom'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import T from 'components/T'
import clsx from 'clsx'

const useStyles = makeStyles((theme) => ({
  button: {
    width: `calc(100% - ${theme.spacing(6)})`,
    margin: theme.spacing(3),
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
      selector: '.predefined-upload',
      content: T('dashboard.tutorial.step9'),
    },
    {
      selector: '.dashboard-tutorial',
      content: T('dashboard.tutorial.step10'),
    },
  ].map((d) => ({
    ...d,
    style: {
      background: theme.palette.background.default,
    },
  }))

  const [isTourOpen, setIsTourOpen] = useState(false)

  return (
    <Paper style={{ height: '100%' }}>
      <PaperTop title={T('dashboard.welcome')} subtitle={T('dashboard.welcomeDesc')} />
      <Button
        className={clsx(classes.button, 'dashboard-tutorial')}
        variant="contained"
        color="primary"
        onClick={() => setIsTourOpen(true)}
      >
        {T('common.tutorial')}
      </Button>
      <Tour
        steps={steps}
        isOpen={isTourOpen}
        onRequestClose={() => setIsTourOpen(false)}
        accentColor={theme.palette.primary.main}
        rounded={theme.shape.borderRadius}
        prevButton={<ArrowBackOutlinedIcon color="action" />}
        nextButton={<ArrowForwardOutlinedIcon color="action" />}
        showCloseButton={false}
      />

      <PaperTop title={T('dashboard.veteran')} subtitle={T('dashboard.veteranDesc')} />
      <Button
        className={clsx(classes.button, 'dashboard-new-experiment')}
        component={Link}
        to="/newExperiment"
        variant="contained"
        color="primary"
        startIcon={<AddIcon />}
      >
        {T('newE.title')}
      </Button>
    </Paper>
  )
}

export default Welcome
