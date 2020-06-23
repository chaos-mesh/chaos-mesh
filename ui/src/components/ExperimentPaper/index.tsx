import { Box, Button, Collapse, IconButton, Paper, Typography, useMediaQuery, useTheme } from '@material-ui/core'
import React, { useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import DeleteOutlineIcon from '@material-ui/icons/DeleteOutline'
import { Experiment } from 'api/experiments.type'
import ExperimentEventsPreview from 'components/ExperimentEventsPreview'
import KeyboardArrowDownIcon from '@material-ui/icons/KeyboardArrowDown'
import KeyboardArrowUpIcon from '@material-ui/icons/KeyboardArrowUp'
import { Link } from 'react-router-dom'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline'
import day from 'lib/dayjs'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    marginRight: {
      '& > *': {
        marginRight: theme.spacing(3),
        '&:last-child': {
          marginRight: 0,
        },
      },
    },
  })
)

interface ExperimentPaperProps {
  experiment: Experiment
  handleSelect: (info: {
    namespace: string
    name: string
    kind: string
    title: string
    description: string
    action: string
  }) => void
  handleDialogOpen: (open: boolean) => void
}

const ExperimentPaper: React.FC<ExperimentPaperProps> = ({ experiment: e, handleSelect, handleDialogOpen }) => {
  const theme = useTheme()
  const isTabletScreen = useMediaQuery(theme.breakpoints.down('sm'))
  const classes = useStyles()

  const [open, setOpen] = useState(false)

  const handleToggle = () => setOpen(!open)

  const handleDelete = (e: Experiment) => () => {
    handleDialogOpen(true)
    handleSelect({
      namespace: e.Namespace,
      name: e.Name,
      kind: e.Kind,
      title: `Delete ${e.Name}?`,
      description: "Once you delete this experiment, it can't be recovered.",
      action: 'delete',
    })
  }

  const handlePause = (e: Experiment) => () => {
    handleDialogOpen(true)
    handleSelect({
      namespace: e.Namespace,
      name: e.Name,
      kind: e.Kind,
      title: `Pause ${e.Name}?`,
      description: 'You can restart the experiment in the same position.',
      action: 'pause',
    })
  }

  const handleStart = (e: Experiment) => () => {
    handleDialogOpen(true)
    handleSelect({
      namespace: e.Namespace,
      name: e.Name,
      kind: e.Kind,
      title: `Start ${e.Name}?`,
      description: 'The operation will take effect immediately.',
      action: 'start',
    })
  }

  const Actions = () => (
    <Box display="flex" alignItems="center" className={classes.marginRight}>
      <Typography variant="body1">Created {day(e.created).fromNow()}</Typography>
      {e.status.toLowerCase() === 'paused' ? (
        <IconButton
          color="primary"
          aria-label="Pause experiment"
          component="span"
          size="small"
          onClick={handleStart(e)}
        >
          <PlayCircleOutlineIcon />
        </IconButton>
      ) : (
        <IconButton
          color="primary"
          aria-label="Pause experiment"
          component="span"
          size="small"
          onClick={handlePause(e)}
        >
          <PauseCircleOutlineIcon />
        </IconButton>
      )}
      <IconButton
        color="primary"
        aria-label="Delete experiment"
        component="span"
        size="small"
        onClick={handleDelete(e)}
      >
        <DeleteOutlineIcon />
      </IconButton>

      <Button
        component={Link}
        to={`/experiments/${e.Name}?namespace=${e.Namespace}&kind=${e.Kind}`}
        variant="outlined"
        color="primary"
        size="small"
      >
        Detail
      </Button>
    </Box>
  )

  return (
    <Paper>
      <Box display="flex" justifyContent="space-between" alignItems="center" p={3}>
        <Box display="flex" className={classes.marginRight}>
          <ExperimentEventsPreview events={e.events} />
          <Typography variant="body1">
            {e.Name}
            {isTabletScreen && (
              <Typography variant="subtitle1" color="textSecondary">
                {e.Kind}
              </Typography>
            )}
          </Typography>
          {!isTabletScreen && (
            <Typography variant="body1" color="textSecondary">
              {e.Kind}
            </Typography>
          )}
        </Box>
        {isTabletScreen ? (
          <IconButton aria-label="Expand row" size="small" onClick={handleToggle}>
            {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
        ) : (
          <Actions />
        )}
      </Box>
      {isTabletScreen && (
        <Collapse in={open} timeout="auto">
          <Box display="flex" justifyContent="end" p={3}>
            <Actions />
          </Box>
        </Collapse>
      )}
    </Paper>
  )
}

export default ExperimentPaper
