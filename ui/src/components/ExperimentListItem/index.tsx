import { Box, Collapse, IconButton, Paper, Typography, useMediaQuery, useTheme } from '@material-ui/core'
import React, { useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import { Archive } from 'api/archives.type'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import ErrorOutlineIcon from '@material-ui/icons/ErrorOutline'
import { Experiment } from 'api/experiments.type'
import ExperimentEventsPreview from 'components/ExperimentEventsPreview'
import { IntlShape } from 'react-intl'
import KeyboardArrowDownIcon from '@material-ui/icons/KeyboardArrowDown'
import KeyboardArrowUpIcon from '@material-ui/icons/KeyboardArrowUp'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline'
import { RootState } from 'store'
import T from 'components/T'
import day from 'lib/dayjs'
import { useHistory } from 'react-router-dom'
import { useSelector } from 'react-redux'

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

interface ExperimentListItemProps {
  experiment: Experiment | Archive
  isArchive?: boolean
  handleSelect: (info: { uuid: uuid; title: string; description: string; action: string }) => void
  handleDialogOpen: (open: boolean) => void
  intl: IntlShape
}

const ExperimentListItem: React.FC<ExperimentListItemProps> = ({
  experiment: e,
  isArchive = false,
  handleSelect,
  handleDialogOpen,
  intl,
}) => {
  const theme = useTheme()
  const isTabletScreen = useMediaQuery(theme.breakpoints.down('sm'))
  const classes = useStyles()

  const history = useHistory()

  const { lang } = useSelector((state: RootState) => state.settings)

  const [open, setOpen] = useState(false)

  const handleToggle = (e: any) => {
    e.stopPropagation()
    setOpen(!open)
  }

  const handleAction = (action: string) => (event: React.MouseEvent<HTMLSpanElement>) => {
    event.stopPropagation()

    handleDialogOpen(true)
    switch (action) {
      case 'delete':
        handleSelect({
          uuid: (e as Experiment).uid,
          title: `${intl.formatMessage({ id: 'archives.single' })} ${e.name}?`,
          description: intl.formatMessage({ id: 'experiments.deleteDesc' }),
          action: 'delete',
        })

        return
      case 'pause':
        handleSelect({
          uuid: (e as Experiment).uid,
          title: `${intl.formatMessage({ id: 'common.pause' })} ${e.name}?`,
          description: intl.formatMessage({ id: 'experiments.pauseDesc' }),
          action: 'pause',
        })

        return
      case 'start':
        handleSelect({
          uuid: (e as Experiment).uid,
          title: `${intl.formatMessage({ id: 'common.start' })} ${e.name}?`,
          description: intl.formatMessage({ id: 'experiments.startDesc' }),
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
      <Typography variant="body2">
        {T('experiments.createdAt')}{' '}
        {day(isArchive ? (e as Archive).start_time : (e as Experiment).created)
          .locale(lang)
          .fromNow()}
      </Typography>
      {!isArchive && (
        <>
          {(e as Experiment).status === 'Paused' ? (
            <IconButton
              color="primary"
              title={intl.formatMessage({ id: 'common.start' })}
              aria-label={intl.formatMessage({ id: 'common.start' })}
              component="span"
              size="small"
              onClick={handleAction('start')}
            >
              <PlayCircleOutlineIcon />
            </IconButton>
          ) : (
            <IconButton
              color="primary"
              title={intl.formatMessage({ id: 'common.pause' })}
              aria-label={intl.formatMessage({ id: 'common.pause' })}
              component="span"
              size="small"
              onClick={handleAction('pause')}
            >
              <PauseCircleOutlineIcon />
            </IconButton>
          )}
          <IconButton
            color="primary"
            title={intl.formatMessage({ id: 'archives.single' })}
            aria-label={intl.formatMessage({ id: 'archives.single' })}
            component="span"
            size="small"
            onClick={handleAction('delete')}
          >
            <ArchiveOutlinedIcon />
          </IconButton>
        </>
      )}
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

export default ExperimentListItem
