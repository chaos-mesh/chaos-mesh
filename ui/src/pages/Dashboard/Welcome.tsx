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
      selector: '.nav-search',
      content: T('dashboard.tutorial.step6'),
    },
    {
      selector: '.nav-namespace',
      content: T('dashboard.tutorial.step7'),
    },
    {
      selector: '.predefined-upload',
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
    <Grid container spacing={6}>
      <Grid item xs={6}>
        <Paper style={{ height: '100%' }}>
          <Space className={classes.space}>
            <PaperTop title={T('dashboard.tutorial.title')} subtitle={T('dashboard.tutorial.desc')} />
            <Button
              className="dashboard-tutorial"
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
              className="dashboard-new-experiment"
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
