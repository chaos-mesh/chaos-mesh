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
import loadable from '@loadable/component'
import ArchiveOutlinedIcon from '@mui/icons-material/ArchiveOutlined'
import PauseCircleOutlineIcon from '@mui/icons-material/PauseCircleOutline'
import PlayCircleOutlineIcon from '@mui/icons-material/PlayCircleOutline'
import { Box, Button, Grid, Grow } from '@mui/material'
import yaml from 'js-yaml'
import {
  useDeleteSchedulesUid,
  useGetEvents,
  useGetSchedulesUid,
  usePutSchedulesPauseUid,
  usePutSchedulesStartUid,
} from 'openapi'
import { useIntl } from 'react-intl'
import { useNavigate, useParams } from 'react-router-dom'

import Loading from '@ui/mui-extends/esm/Loading'
import Paper from '@ui/mui-extends/esm/Paper'
import PaperTop from '@ui/mui-extends/esm/PaperTop'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch } from 'store'

import { setAlert, setConfirm } from 'slices/globalStatus'

import EventsTimeline from 'components/EventsTimeline'
import Helmet from 'components/Helmet'
import ObjectConfiguration from 'components/ObjectConfiguration'
import i18n from 'components/T'

const YAMLEditor = loadable(() => import('components/YAMLEditor'))

const Single = () => {
  const navigate = useNavigate()
  const { uuid } = useParams()

  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const { data: schedule, isLoading: isLoading1, refetch } = useGetSchedulesUid(uuid!)
  const { data: events, isLoading: isLoading2 } = useGetEvents({ object_id: uuid, limit: 999 })
  const loading = isLoading1 || isLoading2
  const { mutateAsync: deleteSchedules } = useDeleteSchedulesUid()
  const { mutateAsync: pauseSchedules } = usePutSchedulesPauseUid()
  const { mutateAsync: startSchedules } = usePutSchedulesStartUid()

  const handleSelect = (action: string) => () => {
    switch (action) {
      case 'archive':
        dispatch(
          setConfirm({
            title: `${i18n('archives.single', intl)} ${schedule!.name}`,
            description: i18n('experiments.deleteDesc', intl),
            handle: handleAction('archive'),
          })
        )

        break
      case 'pause':
        dispatch(
          setConfirm({
            title: `${i18n('common.pause', intl)} ${schedule!.name}`,
            description: i18n('experiments.pauseDesc', intl),
            handle: handleAction('pause'),
          })
        )

        break
      case 'start':
        dispatch(
          setConfirm({
            title: `${i18n('common.start', intl)} ${schedule!.name}`,
            description: i18n('experiments.startDesc', intl),
            handle: handleAction('start'),
          })
        )

        break
    }
  }

  const handleAction = (action: string) => () => {
    let actionFunc

    switch (action) {
      case 'archive':
        actionFunc = deleteSchedules

        break
      case 'pause':
        actionFunc = pauseSchedules

        break
      case 'start':
        actionFunc = startSchedules

        break
      default:
        break
    }

    if (actionFunc) {
      actionFunc({ uid: uuid! })
        .then(() => {
          dispatch(
            setAlert({
              type: 'success',
              message: i18n(`confirm.success.${action}`, intl),
            })
          )

          if (action === 'archive') {
            navigate('/schedules')
          }

          if (action === 'pause' || action === 'start') {
            refetch()
          }
        })
        .catch(console.error)
    }
  }

  return (
    <>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <div>
          {schedule && <Helmet title={`Schedule ${schedule.name}`} />}
          <Space spacing={6}>
            <Space direction="row">
              <Button
                variant="outlined"
                size="small"
                startIcon={<ArchiveOutlinedIcon />}
                onClick={handleSelect('archive')}
              >
                {i18n('archives.single')}
              </Button>
              {schedule?.status === 'paused' ? (
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<PlayCircleOutlineIcon />}
                  onClick={handleSelect('start')}
                >
                  {i18n('common.start')}
                </Button>
              ) : schedule?.status !== 'finished' ? (
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<PauseCircleOutlineIcon />}
                  onClick={handleSelect('pause')}
                >
                  {i18n('common.pause')}
                </Button>
              ) : null}
            </Space>

            <Paper>{schedule && <ObjectConfiguration config={schedule} inSchedule />}</Paper>

            <Grid container>
              <Grid item xs={12} lg={6} sx={{ pr: 3 }}>
                <Paper sx={{ display: 'flex', flexDirection: 'column', height: 600 }}>
                  <PaperTop title={i18n('events.title')} boxProps={{ mb: 3 }} />
                  <Box flex={1} overflow="scroll">
                    {events && <EventsTimeline events={events} />}
                  </Box>
                </Paper>
              </Grid>
              <Grid item xs={12} lg={6} sx={{ pl: 3 }}>
                <Paper sx={{ height: 600, p: 0 }}>
                  {schedule && (
                    <Space display="flex" flexDirection="column" height="100%">
                      <PaperTop title={i18n('common.definition')} boxProps={{ p: 4.5, pb: 0 }} />
                      <Box flex={1}>
                        <YAMLEditor name={schedule.name} data={yaml.dump(schedule.kube_object)} download />
                      </Box>
                    </Space>
                  )}
                </Paper>
              </Grid>
            </Grid>
          </Space>
        </div>
      </Grow>

      {loading && <Loading />}
    </>
  )
}

export default Single
