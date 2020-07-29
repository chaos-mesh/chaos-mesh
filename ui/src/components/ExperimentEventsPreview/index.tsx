import { Box, CircularProgress } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import CheckCircleOutlineIcon from '@material-ui/icons/CheckCircleOutline'
import { Event } from 'api/events.type'
import React from 'react'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      position: 'relative',
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      marginLeft: theme.spacing(3),
      cursor: 'pointer',
      '&:first-child': {
        marginLeft: 0,
      },
    },
    success: {
      color: theme.palette.success.main,
    },
    bottom: {
      color: theme.palette.grey[theme.palette.type === 'light' ? 200 : 700],
    },
    top: {
      position: 'absolute',
      left: 0,
    },
  })
)

interface ExperimentEventsPreviewProps {
  events: Event[] | undefined
}

const ExperimentEventsPreview: React.FC<ExperimentEventsPreviewProps> = ({ events }) => {
  const classes = useStyles()

  const props = {
    variant: 'determinate' as 'determinate',
    size: 20,
    thickness: 5,
    value: 100,
  }

  return events ? (
    <Box display="flex">
      {events.map((e) => (
        <Box key={e.id} className={classes.root}>
          {e.finish_time ? (
            <CheckCircleOutlineIcon className={classes.success} />
          ) : (
            <Box display="flex" alignItems="center">
              <CircularProgress className={classes.bottom} {...props} />
              <CircularProgress className={classes.top} {...props} variant="indeterminate" disableShrink value={50} />
            </Box>
          )}
        </Box>
      ))}
    </Box>
  ) : null
}

export default ExperimentEventsPreview
