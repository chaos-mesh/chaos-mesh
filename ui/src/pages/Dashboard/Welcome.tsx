import { Button, Grid } from '@material-ui/core'
import Tour, { ReactourStep } from 'reactour'

import AccountTreeOutlinedIcon from '@material-ui/icons/AccountTreeOutlined'
import ArrowBackOutlinedIcon from '@material-ui/icons/ArrowBackOutlined'
import ArrowForwardOutlinedIcon from '@material-ui/icons/ArrowForwardOutlined'
import ExperimentIcon from 'components-mui/Icons/Experiment'
import { Link } from 'react-router-dom'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import ScheduleIcon from '@material-ui/icons/Schedule'
import Space from 'components-mui/Space'
import T from 'components/T'
import { makeStyles } from '@material-ui/styles'
import { useState } from 'react'
import { useTheme } from '@material-ui/core/styles'

const useStyles = makeStyles((theme) => ({
  space: {
    width: '75%',
    [theme.breakpoints.down('md')]: {
      width: 'unset',
    },
  },
}))

const Welcome = () => {
  const classes = useStyles()
  const theme = useTheme()

  const steps: ReactourStep[] = [
    {
      selector: '.tutorial-dashboard',
      content: T('dashboard.tutorial.steps.dashboard'),
    },
    {
      selector: '.tutorial-workflows',
      content: T('dashboard.tutorial.steps.workflows'),
    },
    {
      selector: '.tutorial-schedules',
      content: T('dashboard.tutorial.steps.schedules'),
    },
    {
      selector: '.tutorial-experiments',
      content: T('dashboard.tutorial.steps.experiments'),
    },
    {
      selector: '.tutorial-events',
      content: T('dashboard.tutorial.steps.events'),
    },
    {
      selector: '.tutorial-archives',
      content: T('dashboard.tutorial.steps.archives'),
    },
    {
      selector: '.tutorial-newW',
      content: T('dashboard.tutorial.steps.newW'),
    },
    {
      selector: '.tutorial-newS',
      content: T('dashboard.tutorial.steps.newS'),
    },
    {
      selector: '.tutorial-newE',
      content: T('dashboard.tutorial.steps.newE'),
    },
    {
      selector: '.tutorial-search',
      content: T('dashboard.tutorial.steps.search'),
    },
    {
      selector: '.tutorial-namespace',
      content: T('dashboard.tutorial.steps.namespace'),
    },
    {
      selector: '.tutorial-predefined',
      content: T('dashboard.tutorial.steps.predefined'),
    },
    {
      selector: '.tutorial-end',
      content: T('dashboard.tutorial.steps.end'),
    },
  ].map((d) => ({
    ...d,
    style: {
      background: theme.palette.background.default,
    },
  }))

  const [isTourOpen, setIsTourOpen] = useState(false)

  return (
    <Grid container spacing={6}>
      <Grid item xs={6}>
        <Paper style={{ height: '100%' }}>
          <Space className={classes.space}>
            <PaperTop title={T('dashboard.tutorial.title')} subtitle={T('dashboard.tutorial.desc')} />
            <Button
              className="tutorial-end"
              variant="contained"
              color="primary"
              fullWidth
              onClick={() => setIsTourOpen(true)}
            >
              {T('common.tutorial')}
            </Button>
            <Tour
              steps={steps}
              isOpen={isTourOpen}
              onRequestClose={() => setIsTourOpen(false)}
              accentColor={theme.palette.primary.main}
              rounded={theme.shape.borderRadius as number}
              prevButton={<ArrowBackOutlinedIcon color="action" />}
              nextButton={<ArrowForwardOutlinedIcon color="action" />}
              showCloseButton={false}
            />
            <PaperTop title={T('dashboard.newbie')} subtitle={T('dashboard.newbieDesc')} />
            <Button
              className="tutorial-newE"
              component={Link}
              to="/experiments/new"
              variant="contained"
              color="primary"
              fullWidth
              startIcon={<ExperimentIcon />}
            >
              {T('newE.title')}
            </Button>
          </Space>
        </Paper>
      </Grid>
      <Grid item xs={6}>
        <Paper style={{ height: '100%' }}>
          <Space className={classes.space}>
            <PaperTop title={T('dashboard.startAWorkflow')} subtitle={T('dashboard.startAWorkflowDesc')} />
            <Button
              className="tutorial-newW"
              component={Link}
              to="/workflows/new"
              variant="contained"
              color="primary"
              fullWidth
              startIcon={<AccountTreeOutlinedIcon />}
            >
              {T('newW.title')}
            </Button>
            <PaperTop title={T('dashboard.startASchedule')} subtitle={T('dashboard.startAScheduleDesc')} />
            <Button
              className="tutorial-newS"
              component={Link}
              to="/schedules/new"
              variant="contained"
              color="primary"
              fullWidth
              startIcon={<ScheduleIcon />}
            >
              {T('newS.title')}
            </Button>
          </Space>
        </Paper>
      </Grid>
    </Grid>
  )
}

export default Welcome
