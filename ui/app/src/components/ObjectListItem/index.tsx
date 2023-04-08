/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import ArchiveOutlinedIcon from '@mui/icons-material/ArchiveOutlined'
import DeleteOutlinedIcon from '@mui/icons-material/DeleteOutlined'
import PauseCircleOutlineIcon from '@mui/icons-material/PauseCircleOutline'
import PlayCircleOutlineIcon from '@mui/icons-material/PlayCircleOutline'
import { Box, IconButton, Typography } from '@mui/material'
import _ from 'lodash'
import { TypesArchive, TypesExperiment, TypesSchedule } from 'openapi/index.schemas'
import { useIntl } from 'react-intl'
import { useNavigate } from 'react-router-dom'

import Paper from '@ui/mui-extends/esm/Paper'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreSelector } from 'store'

import StatusLabel from 'components/StatusLabel'
import i18n from 'components/T'

import DateTime, { format } from 'lib/luxon'

interface ObjectListItemProps {
  type?: 'schedule' | 'experiment' | 'archive'
  archive?: 'workflow' | 'schedule' | 'experiment'
  data: TypesSchedule | TypesExperiment | TypesArchive
  onSelect: (info: { uuid: uuid; title: string; description: string; action: string }) => void
}

const ObjectListItem: React.FC<ObjectListItemProps> = ({ data, type = 'experiment', archive, onSelect }) => {
  const navigate = useNavigate()
  const intl = useIntl()

  const { lang } = useStoreSelector((state) => state.settings)

  const handleAction = (action: string) => (event: React.MouseEvent<HTMLSpanElement>) => {
    event.stopPropagation()

    switch (action) {
      case 'archive':
        onSelect({
          title: `${i18n('archives.single', intl)} ${data.name}`,
          description: i18n(`${type}s.deleteDesc`, intl),
          action,
          uuid: data.uid!,
        })

        return
      case 'pause':
        onSelect({
          title: `${i18n('common.pause', intl)} ${data.name}`,
          description: i18n('experiments.pauseDesc', intl),
          action,
          uuid: data.uid!,
        })

        return
      case 'start':
        onSelect({
          title: `${i18n('common.start', intl)} ${data.name}`,
          description: i18n('experiments.startDesc', intl),
          action,
          uuid: data.uid!,
        })

        return
      case 'delete':
        onSelect({
          title: `${i18n('common.delete', intl)} ${data.name}`,
          description: i18n('archives.deleteDesc', intl),
          action,
          uuid: data.uid!,
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

    navigate(path)
  }

  const Actions = () => (
    <Space direction="row" justifyContent="end" alignItems="center">
      <Typography variant="body2" title={format(data.created_at!)}>
        {i18n('table.created')}{' '}
        {DateTime.fromISO(data.created_at!, {
          locale: lang,
        }).toRelative()}
      </Typography>
      {(type === 'schedule' || type === 'experiment') &&
        ((data as any).status === 'paused' ? (
          <IconButton color="primary" title={i18n('common.start', intl)} size="small" onClick={handleAction('start')}>
            <PlayCircleOutlineIcon />
          </IconButton>
        ) : (data as any).status !== 'finished' ? (
          <IconButton color="primary" title={i18n('common.pause', intl)} size="small" onClick={handleAction('pause')}>
            <PauseCircleOutlineIcon />
          </IconButton>
        ) : null)}
      {type !== 'archive' && (
        <IconButton
          color="primary"
          title={i18n('archives.single', intl)}
          size="small"
          onClick={handleAction('archive')}
        >
          <ArchiveOutlinedIcon />
        </IconButton>
      )}
      {type === 'archive' && (
        <IconButton color="primary" title={i18n('common.delete', intl)} size="small" onClick={handleAction('delete')}>
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
          {type !== 'archive' && <StatusLabel status={(data as any).status} />}
          <Typography component="div" title={data.name}>
            {_.truncate(data.name!)}
          </Typography>
          <Typography component="div" variant="body2" color="textSecondary" title={data.uid}>
            {_.truncate(data.uid!)}
          </Typography>
        </Space>

        <Actions />
      </Box>
    </Paper>
  )
}

export default ObjectListItem
