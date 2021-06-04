import { Box, IconButton, Typography } from '@material-ui/core'

import { Archive } from 'api/archives.type'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import DateTime from 'lib/luxon'
import DeleteOutlinedIcon from '@material-ui/icons/DeleteOutlined'
import { Experiment } from 'api/experiments.type'
import ExperimentStatus from 'components/ExperimentStatus'
import Paper from 'components-mui/Paper'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline'
import { Schedule } from 'api/schedules.type'
import Space from 'components-mui/Space'
import T from 'components/T'
import { makeStyles } from '@material-ui/styles'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'
import { useStoreSelector } from 'store'

const useStyles = makeStyles((theme) => ({
  root: {
    '&:hover': {
      backgroundColor: theme.palette.action.hover,
      cursor: 'pointer',
    },
  },
}))

interface ObjectListItemProps {
  type?: 'schedule' | 'experiment' | 'archive'
  archive?: 'workflow' | 'schedule' | 'experiment'
  data: Schedule | Experiment | Archive
  onSelect: (info: { uuid: uuid; title: string; description: string; action: string }) => void
}

const ObjectListItem: React.FC<ObjectListItemProps> = ({ data, type = 'experiment', archive, onSelect }) => {
  const classes = useStyles()

  const history = useHistory()
  const intl = useIntl()

  const { lang } = useStoreSelector((state) => state.settings)

  const handleAction = (action: string) => (event: React.MouseEvent<HTMLSpanElement>) => {
    event.stopPropagation()

    switch (action) {
      case 'archive':
        onSelect({
          title: `${intl.formatMessage({ id: 'archives.single' })} ${data.name}`,
          description: intl.formatMessage({ id: `${type}s.deleteDesc` }),
          action,
          uuid: data.uid,
        })

        return
      case 'pause':
        onSelect({
          title: `${intl.formatMessage({ id: 'common.pause' })} ${data.name}`,
          description: intl.formatMessage({ id: 'experiments.pauseDesc' }),
          action,
          uuid: data.uid,
        })

        return
      case 'start':
        onSelect({
          title: `${intl.formatMessage({ id: 'common.start' })} ${data.name}`,
          description: intl.formatMessage({ id: 'experiments.startDesc' }),
          action,
          uuid: data.uid,
        })

        return
      case 'delete':
        onSelect({
          title: `${intl.formatMessage({ id: 'common.delete' })} ${data.name}`,
          description: intl.formatMessage({ id: 'archives.deleteDesc' }),
          action,
          uuid: data.uid,
        })

        return
      default:
        return
    }
  }

  const handleJumpTo = () => {
    let path
    switch (type) {
      case 'schedule':
      case 'experiment':
        path = `/${type}s/${data.uid}`
        break
      case 'archive':
        path = `/archives/${data.uid}?kind=${archive!}`
        break
    }

    history.push(path)
  }

  const Actions = () => (
    <Space display="flex" justifyContent="end" alignItems="center">
      <Typography variant="body2">
        {T('table.created')}{' '}
        {DateTime.fromISO(data.created_at, {
          locale: lang,
        }).toRelative()}
      </Typography>
      {type === 'experiment' && (
        <>
          {(data as Experiment).status === 'paused' ? (
            <IconButton
              color="primary"
              title={intl.formatMessage({ id: 'common.start' })}
              aria-label={intl.formatMessage({ id: 'common.start' })}
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
              size="small"
              onClick={handleAction('pause')}
            >
              <PauseCircleOutlineIcon />
            </IconButton>
          )}
        </>
      )}
      {type !== 'archive' && (
        <IconButton
          color="primary"
          title={intl.formatMessage({ id: 'archives.single' })}
          aria-label={intl.formatMessage({ id: 'archives.single' })}
          size="small"
          onClick={handleAction('archive')}
        >
          <ArchiveOutlinedIcon />
        </IconButton>
      )}
      {type === 'archive' && (
        <IconButton
          color="primary"
          title={intl.formatMessage({ id: 'common.delete' })}
          aria-label={intl.formatMessage({ id: 'common.delete' })}
          size="small"
          onClick={handleAction('delete')}
        >
          <DeleteOutlinedIcon />
        </IconButton>
      )}
    </Space>
  )

  return (
    <Paper padding={0} className={classes.root} onClick={handleJumpTo}>
      <Box display="flex" justifyContent="space-between" alignItems="center" p={3}>
        <Space display="flex" alignItems="center">
          {type === 'experiment' && <ExperimentStatus status={(data as Experiment).status} />}
          <Typography variant="body1" component="div">
            {data.name}
          </Typography>
        </Space>

        <Actions />
      </Box>
    </Paper>
  )
}

export default ObjectListItem
