import { Box, Button, Collapse, IconButton, Paper, Typography, useMediaQuery, useTheme } from '@material-ui/core'
import React, { useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import { Archive } from 'api/archives.type'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import ErrorOutlineIcon from '@material-ui/icons/ErrorOutline'
import { Experiment } from 'api/experiments.type'
import ExperimentEventsPreview from 'components/ExperimentEventsPreview'
import KeyboardArrowDownIcon from '@material-ui/icons/KeyboardArrowDown'
import KeyboardArrowUpIcon from '@material-ui/icons/KeyboardArrowUp'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline'
import day from 'lib/dayjs'
import { useHistory } from 'react-router-dom'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      '&:hover': {
        backgroundColor: theme.palette.action.hover,
        cursor: 'pointer',
      },
    },
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
  handleSelect: (info: { uuid: uuid; title: string; description: string; action: string }) => void
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

  const history = useHistory()

  const [open, setOpen] = useState(false)

  const handleToggle = () => setOpen(!open)

  const handleAction = (action: string) => (event: React.MouseEvent<HTMLSpanElement>) => {
    event.stopPropagation()

    handleDialogOpen(true)
    switch (action) {
      case 'delete':
        handleSelect({
          uuid: (e as Experiment).uid,
          title: `Archive ${e.name}?`,
          description: 'You can still find this experiment in the archives.',
          action: 'delete',
        })

        return
      case 'pause':
        handleSelect({
          uuid: (e as Experiment).uid,
          title: `Pause ${e.name}?`,
          description: 'You can restart the experiment in the same position.',
          action: 'pause',
        })

        return
      case 'start':
        handleSelect({
          uuid: (e as Experiment).uid,
          title: `Start ${e.name}?`,
          description: 'The operation will take effect immediately.',
          action: 'start',
        })

        return
      default:
        return
    }
  }

  const handleJumpTo = () => history.push(isArchive ? `/archives/${e.uid}` : `/experiments/${(e as Experiment).uid}`)

  const Actions = () => (
    <Box display="flex" justifyContent="flex-end" alignItems="center" className={classes.marginRight}>
      {!isArchive && (
        <>
          <Typography variant="body2">Created {day((e as Experiment).created).fromNow()}</Typography>
          {(e as Experiment).status === 'Paused' ? (
            <IconButton
              color="primary"
              title="Start experiment"
              aria-label="Start experiment"
              component="span"
              size="small"
              onClick={handleAction('start')}
            >
              <PlayCircleOutlineIcon />
            </IconButton>
          ) : (
            <IconButton
              color="primary"
              title="Pause experiment"
              aria-label="Pause experiment"
              component="span"
              size="small"
              onClick={handleAction('pause')}
            >
              <PauseCircleOutlineIcon />
            </IconButton>
          )}
          <IconButton
            color="primary"
            title="Archive experiment"
            aria-label="Archive experiment"
            component="span"
            size="small"
            onClick={handleAction('delete')}
          >
            <ArchiveOutlinedIcon />
          </IconButton>
        </>
      )}
      <Button variant="outlined" color="primary" size="small">
        {isArchive ? 'Report' : 'Detail'}
      </Button>
    </Box>
  )

  return (
    <Paper variant="outlined" className={classes.root} onClick={handleJumpTo}>
      <Box display="flex" justifyContent="space-between" alignItems="center" p={3}>
        <Box display="flex" alignItems="center" className={classes.marginRight}>
          {!isArchive &&
            ((e as Experiment).status === 'Failed' ? (
              <ErrorOutlineIcon color="error" />
            ) : (
              <ExperimentEventsPreview events={(e as Experiment).events} />
            ))}
          <Typography variant="body1" component="div">
            {e.name}
            {isTabletScreen && (
              <Typography variant="body2" color="textSecondary">
                {e.uid}
              </Typography>
            )}
          </Typography>
          {!isTabletScreen && (
            <Typography variant="body2" color="textSecondary">
              {e.uid}
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
