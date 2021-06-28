import { Box, IconButton, Typography } from '@material-ui/core'
import DateTime, { format } from 'lib/luxon'

import { Archive } from 'api/archives.type'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import DeleteOutlinedIcon from '@material-ui/icons/DeleteOutlined'
import { Experiment } from 'api/experiments.type'
import Paper from 'components-mui/Paper'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline'
import { Schedule } from 'api/schedules.type'
import Space from 'components-mui/Space'
import StatusLabel from 'components-mui/StatusLabel'
import T from 'components/T'
import { truncate } from 'lib/utils'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'
import { useStoreSelector } from 'store'

interface ObjectListItemProps {
  type?: 'schedule' | 'experiment' | 'archive'
  archive?: 'workflow' | 'schedule' | 'experiment'
  data: Schedule | Experiment | Archive
  onSelect: (info: { uuid: uuid; title: string; description: string; action: string }) => void
}

const ObjectListItem: React.FC<ObjectListItemProps> = ({ data, type = 'experiment', archive, onSelect }) => {
  const history = useHistory()
  const intl = useIntl()

  const { lang } = useStoreSelector((state) => state.settings)

  const handleAction = (action: string) => (event: React.MouseEvent<HTMLSpanElement>) => {
    event.stopPropagation()

    switch (action) {
      case 'archive':
        onSelect({
          title: `${T('archives.single', intl)} ${data.name}`,
          description: T(`${type}s.deleteDesc`, intl),
          action,
          uuid: data.uid,
        })

        return
      case 'pause':
        onSelect({
          title: `${T('common.pause', intl)} ${data.name}`,
          description: T('experiments.pauseDesc', intl),
          action,
          uuid: data.uid,
        })

        return
      case 'start':
        onSelect({
          title: `${T('common.start', intl)} ${data.name}`,
          description: T('experiments.startDesc', intl),
          action,
          uuid: data.uid,
        })

        return
      case 'delete':
        onSelect({
          title: `${T('common.delete', intl)} ${data.name}`,
          description: T('archives.deleteDesc', intl),
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
    <Space direction="row" justifyContent="end" alignItems="center">
      <Typography variant="body2" title={format(data.created_at)}>
        {T('table.created')}{' '}
        {DateTime.fromISO(data.created_at, {
          locale: lang,
        }).toRelative()}
      </Typography>
      {(type === 'schedule' || type === 'experiment') &&
        ((data as Experiment).status === 'paused' ? (
          <IconButton color="primary" title={T('common.start', intl)} size="small" onClick={handleAction('start')}>
            <PlayCircleOutlineIcon />
          </IconButton>
        ) : (data as Experiment).status !== 'finished' ? (
          <IconButton color="primary" title={T('common.pause', intl)} size="small" onClick={handleAction('pause')}>
            <PauseCircleOutlineIcon />
          </IconButton>
        ) : null)}
      {type !== 'archive' && (
        <IconButton color="primary" title={T('archives.single', intl)} size="small" onClick={handleAction('archive')}>
          <ArchiveOutlinedIcon />
        </IconButton>
      )}
      {type === 'archive' && (
        <IconButton color="primary" title={T('common.delete', intl)} size="small" onClick={handleAction('delete')}>
          <DeleteOutlinedIcon />
        </IconButton>
      )}
    </Space>
  )

  return (
    <Paper
      sx={{
        p: 0,
        ':hover': {
          bgcolor: 'action.hover',
          cursor: 'pointer',
        },
      }}
      onClick={handleJumpTo}
    >
      <Box display="flex" justifyContent="space-between" alignItems="center" p={3}>
        <Space direction="row" alignItems="center">
          {type !== 'archive' && <StatusLabel status={(data as Experiment).status} />}
          <Typography component="div" title={data.name}>
            {truncate(data.name)}
          </Typography>
          <Typography component="div" variant="body2" color="textSecondary" title={data.uid}>
            {truncate(data.uid)}
          </Typography>
        </Space>

        <Actions />
      </Box>
    </Paper>
  )
}

export default ObjectListItem
