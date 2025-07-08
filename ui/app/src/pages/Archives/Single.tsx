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
import { useGetArchivesSchedulesUid, useGetArchivesUid, useGetArchivesWorkflowsUid, useGetEvents } from '@/openapi'
import { Box, Grid, Grow } from '@mui/material'
import yaml from 'js-yaml'
import { lazy } from 'react'
import { useParams } from 'react-router'

import EventsTimeline from '@/components/EventsTimeline'
import ObjectConfiguration from '@/components/ObjectConfiguration'
import i18n from '@/components/T'

import { useQuery } from '@/lib/hooks'

const YAMLEditor = lazy(() => import('@/components/YAMLEditor'))

const Single = () => {
  const { uuid } = useParams()
  const query = useQuery()
  const kind = query.get('kind') || 'experiment'
  const useGetArchives =
    kind === 'workflow'
      ? useGetArchivesWorkflowsUid
      : kind === 'schedule'
        ? useGetArchivesSchedulesUid
        : useGetArchivesUid

  const { data: archive, isLoading: loadingArchives } = useGetArchives(uuid!)
  const { data: events, isLoading: loadingEvents } = useGetEvents(
    {
      object_id: uuid,
      limit: 999,
    },
    { query: { enabled: kind !== 'workflow' } },
  )
  const loading = kind === 'workflow' ? loadingArchives : loadingArchives && loadingEvents

  const YAML = () => (
    <Paper sx={{ height: kind === 'workflow' ? (theme) => `calc(100vh - 56px - ${theme.spacing(18)})` : 600, p: 0 }}>
      {archive && (
        <Space display="flex" flexDirection="column" height="100%">
          <PaperTop title={i18n('common.definition')} boxProps={{ p: 4.5, pb: 0 }} />
          <Box flex={1}>
            <YAMLEditor
              name={archive.name}
              data={yaml.dump(archive.kube_object)}
              download
              aceProps={{ readOnly: true }}
            />
          </Box>
        </Space>
      )}
    </Paper>
  )

  return (
    <>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <div>
          {archive && <title>{`Archive ${archive.name}`}</title>}
          {kind !== 'workflow' ? (
            <Space spacing={6}>
              {archive && (
                <Paper>
                  <ObjectConfiguration config={archive} inSchedule={kind === 'schedule'} inArchive={true} />
                </Paper>
              )}

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
                  <YAML />
                </Grid>
              </Grid>
            </Space>
          ) : (
            <YAML />
          )}
        </div>
      </Grow>

      {loading && <Loading />}
    </>
  )
}

export default Single
