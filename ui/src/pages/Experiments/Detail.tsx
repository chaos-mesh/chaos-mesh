import React from 'react'

import { Button } from '@material-ui/core'
import DeleteOutlinedIcon from '@material-ui/icons/DeleteOutlined'
import StopIcon from '@material-ui/icons/Stop'
import SettingsIcon from '@material-ui/icons/Settings'

import PageBar from '../../components/PageBar'
import ToolBar from '../../components/ToolBar'
import Container from '../../components/Container'

import { makeStyles, Theme, createStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    button: {
      marginRight: theme.spacing(2),
    },
  })
)

export default function ExperimentDetail() {
  const classes = useStyles()

  return (
    <>
      <PageBar
        breadcrumbs={[
          { name: 'Experiments', path: '/experiments' },
          { name: 'tikv-failure' },
        ]}
      />
      <ToolBar>
        <Button
          className={classes.button}
          variant="outlined"
          startIcon={<StopIcon />}
        >
          Stop
        </Button>
        <Button
          className={classes.button}
          variant="outlined"
          startIcon={<SettingsIcon />}
        >
          Config
        </Button>
        <Button
          className={classes.button}
          variant="outlined"
          color="secondary"
          startIcon={<DeleteOutlinedIcon />}
        >
          Delete
        </Button>
      </ToolBar>

      <Container>Experiment Detail</Container>
    </>
  )
}
