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
import Loading from '@/mui-extends/Loading'
import Paper from '@/mui-extends/Paper'
import PaperTop from '@/mui-extends/PaperTop'
import Space from '@/mui-extends/Space'
import {
  useDeleteSchedulesUid,
  useGetEvents,
  useGetSchedulesUid,
  usePutSchedulesPauseUid,
  usePutSchedulesStartUid,
} from '@/openapi'
import { useComponentActions } from '@/zustand/component'
import ArchiveOutlinedIcon from '@mui/icons-material/ArchiveOutlined'
import PlayCircleOutlineIcon from '@mui/icons-material/PlayCircleOutline'
import { Box, Button, Grid, Grow } from '@mui/material'
import yaml from 'js-yaml'
import { lazy } from 'react'
import { useIntl } from 'react-intl'
import { useNavigate, useParams } from 'react-router'

import EventsTimeline from '@/components/EventsTimeline'
import ObjectConfiguration from '@/components/ObjectConfiguration'
import i18n from '@/components/T'

const YAMLEditor = lazy(() => import('@/components/YAMLEditor'))

const Single = () => {
  const navigate = useNavigate()
  const { uuid } = useParams()

  const intl = useIntl()

  const { setConfirm, setAlert } = useComponentActions()

  const { data: schedule, isLoading: isLoading1, refetch } = useGetSchedulesUid(uuid!)
  const { data: events, isLoading: isLoading2 } = useGetEvents({ object_id: uuid, limit: 999 })
  const loading = isLoading1 || isLoading2
  const { mutateAsync: deleteSchedules } = useDeleteSchedulesUid()
  const { mutateAsync: pauseSchedules } = usePutSchedulesPauseUid()
  const { mutateAsync: startSchedules } = usePutSchedulesStartUid()

  const handleSelect = (action: string) => () => {
    switch (action) {
      case 'archive':
        setConfirm({
          title: `${i18n('archives.single', intl)} ${schedule!.name}`,
          description: i18n('experiments.deleteDesc', intl),
          handle: handleAction('archive'),
        })

        break
      case 'pause':
        setConfirm({
          title: `${i18n('common.pause', intl)} ${schedule!.name}`,
          description: i18n('experiments.pauseDesc', intl),
          handle: handleAction('pause'),
        })

        break
      case 'start':
        setConfirm({
          title: `${i18n('common.start', intl)} ${schedule!.name}`,
          description: i18n('experiments.startDesc', intl),
          handle: handleAction('start'),
        })

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
          setAlert({
            type: 'success',
            message: i18n(`confirm.success.${action}`, intl),
          })

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
          {schedule && <title>{`Schedule ${schedule.name}`}</title>}
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
              ) : null}
            </Space>

            <Paper>{schedule && <ObjectConfiguration config={schedule} inSchedule />}</Paper>

            <Grid container>
              <Grid item xs={12} lg={6} sx={{ pr: 3 }}>
                <EventsTimeline events={events} paperProps={{ sx: { height: 600 } }} />
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
