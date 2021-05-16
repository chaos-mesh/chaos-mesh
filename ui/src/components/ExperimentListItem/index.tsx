import { Box, Collapse, IconButton, Typography, useMediaQuery, useTheme } from '@material-ui/core'
import React, { useState } from 'react'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import { Archive } from 'api/archives.type'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import DateTime from 'lib/luxon'
import DeleteOutlinedIcon from '@material-ui/icons/DeleteOutlined'
import ErrorOutlineIcon from '@material-ui/icons/ErrorOutline'
import { Experiment } from 'api/experiments.type'
import ExperimentStatus from 'components/ExperimentStatus'
import { IntlShape } from 'react-intl'
import KeyboardArrowDownIcon from '@material-ui/icons/KeyboardArrowDown'
import KeyboardArrowUpIcon from '@material-ui/icons/KeyboardArrowUp'
import Paper from 'components-mui/Paper'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline'
import { RootState } from 'store'
import Space from 'components-mui/Space'
import T from 'components/T'
import { truncate } from 'lib/utils'
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
  })
)

interface ExperimentListItemProps {
  experiment: Experiment | Archive
  isArchive?: boolean
  onSelect: (info: { uuid: uuid; title: string; description: string; action: string }) => void
  intl: IntlShape
}

const ExperimentListItem: React.FC<ExperimentListItemProps> = ({
  experiment: e,
  isArchive = false,
  onSelect,
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

    switch (action) {
      case 'archive':
        onSelect({
          title: `${intl.formatMessage({ id: 'archives.single' })} ${e.name}`,
          description: intl.formatMessage({ id: 'experiments.deleteDesc' }),
          action,
          uuid: e.uid,
        })

        return
      case 'pause':
        onSelect({
          title: `${intl.formatMessage({ id: 'common.pause' })} ${e.name}`,
          description: intl.formatMessage({ id: 'experiments.pauseDesc' }),
          action,
          uuid: e.uid,
        })

        return
      case 'start':
        onSelect({
          title: `${intl.formatMessage({ id: 'common.start' })} ${e.name}`,
          description: intl.formatMessage({ id: 'experiments.startDesc' }),
          action,
          uuid: e.uid,
        })

        return
      case 'delete':
        onSelect({
          title: `${intl.formatMessage({ id: 'common.delete' })} ${e.name}`,
          description: intl.formatMessage({ id: 'archives.deleteDesc' }),
          action,
          uuid: e.uid,
        })

        return
      default:
        return
    }
  }

  const handleJumpTo = () => history.push(isArchive ? `/archives/${e.uid}` : `/experiments/${e.uid}`)

  const Actions = () => (
    <Space display="flex" justifyContent="end" alignItems="center">
      <Typography variant="body2">
        {T('table.created')}{' '}
        {DateTime.fromISO(isArchive ? (e as Archive).start_time : (e as Experiment).created, {
          locale: lang,
        }).toRelative()}
      </Typography>
      {isArchive ? (
        <IconButton
          color="primary"
          title={intl.formatMessage({ id: 'common.delete' })}
          aria-label={intl.formatMessage({ id: 'common.delete' })}
          component="span"
          size="small"
          onClick={handleAction('delete')}
        >
          <DeleteOutlinedIcon />
        </IconButton>
      ) : (
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
            onClick={handleAction('archive')}
          >
            <ArchiveOutlinedIcon />
          </IconButton>
        </>
      )}
    </Space>
  )

  return (
    <Paper padding={0} className={classes.root} onClick={handleJumpTo}>
      <Box display="flex" justifyContent="space-between" alignItems="center" p={3}>
        <Space display="flex" alignItems="center">
          {!isArchive &&
            ((e as Experiment).status === 'Failed' ? (
              <ErrorOutlineIcon color="error" />
            ) : (
              <ExperimentStatus status={(e as Experiment).status} />
            ))}
          <Typography variant="body1" component="div">
            {e.name}
            {isTabletScreen && (
              <Typography variant="body2" color="textSecondary" title={e.uid}>
                {truncate(e.uid)}
              </Typography>
            )}
          </Typography>
          {!isTabletScreen && (
            <Typography variant="body2" color="textSecondary" title={e.uid}>
              {truncate(e.uid)}
            </Typography>
          )}
        </Space>
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
