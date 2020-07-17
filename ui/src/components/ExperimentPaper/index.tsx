import { Box, Button, Collapse, IconButton, Paper, Typography, useMediaQuery, useTheme } from '@material-ui/core'
import React, { useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import { Archive } from 'api/archives.type'
import DeleteOutlineIcon from '@material-ui/icons/DeleteOutline'
import ErrorOutlineIcon from '@material-ui/icons/ErrorOutline'
import { Experiment } from 'api/experiments.type'
import ExperimentEventsPreview from 'components/ExperimentEventsPreview'
import KeyboardArrowDownIcon from '@material-ui/icons/KeyboardArrowDown'
import KeyboardArrowUpIcon from '@material-ui/icons/KeyboardArrowUp'
import { Link } from 'react-router-dom'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline'
import RefreshIcon from '@material-ui/icons/Refresh'
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
  experiment: Experiment | Archive
  isArchive?: boolean
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

const ExperimentPaper: React.FC<ExperimentPaperProps> = ({
  experiment: e,
  isArchive = false,
  handleSelect,
  handleDialogOpen,
}) => {
  const theme = useTheme()
  const isTabletScreen = useMediaQuery(theme.breakpoints.down('sm'))
  const classes = useStyles()

  const [open, setOpen] = useState(false)

  const handleToggle = () => setOpen(!open)

  const handleDelete = () => {
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

  const handlePause = () => {
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

  const handleStart = () => {
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
    <Box display="flex" justifyContent="flex-end" alignItems="center" className={classes.marginRight}>
      {!isArchive && (
        <>
          <Typography variant="body1">Created {day((e as Experiment).created).fromNow()}</Typography>
          {(e as Experiment).status.toLowerCase() === 'paused' ? (
            <IconButton
              color="primary"
              aria-label="Start experiment"
              component="span"
              size="small"
              onClick={handleStart}
            >
              <PlayCircleOutlineIcon />
            </IconButton>
          ) : (
            <IconButton
              color="primary"
              aria-label="Pause experiment"
              component="span"
              size="small"
              onClick={handlePause}
            >
              <PauseCircleOutlineIcon />
            </IconButton>
          )}
          <IconButton
            color="primary"
            aria-label="Delete experiment"
            component="span"
            size="small"
            onClick={handleDelete}
          >
            <DeleteOutlineIcon />
          </IconButton>
        </>
      )}
      {isArchive && (
        <IconButton color="primary" aria-label="Recreate experiment" component="span" size="small">
          <RefreshIcon />
        </IconButton>
      )}
      <Button
        component={Link}
        to={
          isArchive
            ? `/`
            : `/experiments/${e.Name}?namespace=${e.Namespace}&kind=${
                e.Kind
              }&status=${(e as Experiment).status.toLowerCase()}`
        }
        variant="outlined"
        color="primary"
        size="small"
        disabled={!isArchive && (e as Experiment).status.toLowerCase() === 'failed'}
      >
        Detail
      </Button>
    </Box>
  )

  return (
    <Paper variant="outlined">
      <Box display="flex" justifyContent="space-between" alignItems="center" p={3}>
        <Box display="flex" alignItems="center" className={classes.marginRight}>
          {!isArchive && (
            <>
              {(e as Experiment).status.toLowerCase() === 'failed' && <ErrorOutlineIcon color="error" />}
              <ExperimentEventsPreview events={(e as Experiment).events} />
            </>
          )}
          <Typography variant="body1" component="div">
            {e.Name}
            {isTabletScreen && (
              <Typography variant="body2" color="textSecondary">
                {e.Kind}
              </Typography>
            )}
          </Typography>
          {!isTabletScreen && (
            <Typography variant="body2" color="textSecondary">
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
          <Box p={3}>
            <Actions />
          </Box>
        </Collapse>
      )}
    </Paper>
  )
}

export default ExperimentPaper
