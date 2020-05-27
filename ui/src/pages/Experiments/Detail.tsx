import React from 'react'

import { Button } from '@material-ui/core'
import DeleteOutlinedIcon from '@material-ui/icons/DeleteOutlined'
import StopIcon from '@material-ui/icons/Stop'
import SettingsIcon from '@material-ui/icons/Settings'
import { makeStyles, Theme, createStyles } from '@material-ui/core/styles'

import PageBar from '../../components/PageBar'
import ToolBar from '../../components/ToolBar'
import Container from '../../components/Container'
import InfoList from '../../components/InfoList'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    button: {
      marginRight: theme.spacing(2),
    },
  })
)

const fakeExperiment = {
  'start time': '2020-05-22 10:00',
  'end time': '2020-05-22 10:00',
  message:
    'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. ',
  scope: 'Lorem ipsum dolor sit amet',
  target: 'Pod',
  action: 'Killing Pod',
  duration: '1d',
  experiment: 'tikv failure',
}

export default function ExperimentDetail() {
  const classes = useStyles()

  return (
    <>
      <PageBar />
      <ToolBar>
        <Button className={classes.button} variant="outlined" startIcon={<StopIcon />}>
          Stop
        </Button>
        <Button className={classes.button} variant="outlined" startIcon={<SettingsIcon />}>
          Config
        </Button>
        <Button className={classes.button} variant="outlined" color="secondary" startIcon={<DeleteOutlinedIcon />}>
          Delete
        </Button>
      </ToolBar>

      <Container>
        <InfoList info={fakeExperiment} />
      </Container>
    </>
  )
}
