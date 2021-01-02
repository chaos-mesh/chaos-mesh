import { Box, BoxProps, LinearProgress, Typography } from '@material-ui/core'
import { Link, LinkProps } from 'react-router-dom'

import { Event } from 'api/events.type'
import NotFound from 'components-mui/NotFound'
import React from 'react'
import T from 'components/T'
import day from 'lib/dayjs'
import { iconByKind } from 'lib/byKind'
import { makeStyles } from '@material-ui/core/styles'
import { useStoreSelector } from 'store'

const LinkBox: React.ComponentType<BoxProps & LinkProps> = Box as any

interface RecentProps {
  events: Event[]
}

const useStyles = makeStyles((theme) => ({
  event: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    height: 72,
    margin: theme.spacing(3),
    color: 'inherit',
    borderRadius: theme.shape.borderRadius,
    textDecoration: 'none',
    '&:hover': {
      background: theme.palette.divider,
      cursor: 'pointer',
    },
  },
}))

const Recent: React.FC<RecentProps> = ({ events }) => {
  const classes = useStyles()

  const { lang } = useStoreSelector((state) => state.settings)

  return (
    <Box>
      {events.reverse().map((d) => (
        <LinkBox key={d.id} className={classes.event} component={Link} to={`/events?event_id=${d.id}`}>
          <Box display="flex" justifyContent="center" flex={1}>
            {iconByKind(d.kind as any, 'small')}
          </Box>
          <Box display="flex" flexDirection="column" justifyContent="center" flex={2} px={1.5}>
            <Typography gutterBottom>{d.experiment}</Typography>
            <LinearProgress
              variant={d.finish_time ? 'determinate' : 'indeterminate'}
              value={d.finish_time ? 100 : undefined}
              style={{ width: '100%' }}
            />
          </Box>
          <Box display="flex" justifyContent="center" flex={2}>
            {day(d.start_time).locale(lang).fromNow()}
          </Box>
        </LinkBox>
      ))}
      {events.length === 0 && <NotFound>{T('events.noEventsFound')}</NotFound>}
    </Box>
  )
}

export default Recent
