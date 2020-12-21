import { Box, Grid, Grow, Paper } from '@material-ui/core'
import { Event, EventPod } from 'api/events.type'
import React, { useEffect, useMemo, useState } from 'react'

import AffectedPods from 'components/AffectedPods'
import { ArchiveDetail } from 'api/archives.type'
import ArchiveDuration from 'components/ArchiveDuration'
import ArchiveNumberOf from 'components/ArchiveNumberOf'
import EventsTable from 'components/EventsTable'
import ExperimentConfiguration from 'components/ExperimentConfiguration'
import Loading from 'components-mui/Loading'
import PaperTop from 'components-mui/PaperTop'
import T from 'components/T'
import api from 'api'
import { useParams } from 'react-router-dom'

const ArchiveReport: React.FC = () => {
  const { uuid } = useParams<{ uuid: string }>()

  const [loading, setLoading] = useState(true)
  const [detail, setDetail] = useState<ArchiveDetail>()
  const [report, setReport] = useState<{ events: Event[] }>({ events: [] })

  const events = report.events
  const affectedPods = useMemo(
    () =>
      [
        ...new Set(
          events
            .reduce<EventPod[]>((acc, e) => acc.concat(e.pods!), [])
            .map((d) => ({
              pod_ip: d.pod_ip,
              pod_name: d.pod_name,
              namespace: d.namespace,
              action: d.action,
              message: d.message,
            }))
            .map((d) => JSON.stringify(d))
        ),
      ].map((d) => JSON.parse(d)),
    [events]
  )

  const fetchDetail = () => api.archives.detail(uuid).then(({ data }) => setDetail(data))
  const fetchReport = () => api.archives.report(uuid).then(({ data }) => setReport(data))

  useEffect(() => {
    Promise.all([fetchDetail(), fetchReport()])
      .then((_) => setLoading(false))
      .catch(console.error)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <Grid container spacing={6}>
          {detail && (
            <>
              <Grid item xs={6} sm={6} md={3}>
                <ArchiveNumberOf title={T('archives.numberOfRuns')} num={events.length} />
              </Grid>
              <Grid item xs={6} sm={6} md={3}>
                <ArchiveNumberOf title={T('archives.numberOfAffectedPods')} num={affectedPods.length} />
              </Grid>

              <Grid item xs={12} md={6}>
                <ArchiveDuration start={detail.start_time} end={detail.finish_time} />
              </Grid>

              <Grid item xs={12}>
                <Paper variant="outlined">
                  <PaperTop title={T('common.configuration')} />
                  <Box p={3}>
                    <ExperimentConfiguration experimentDetail={detail} />
                  </Box>
                </Paper>
              </Grid>

              <Grid item xs={12}>
                <Paper variant="outlined">
                  <PaperTop title={T('newE.scope.affectedPods')} />
                  <AffectedPods pods={affectedPods} />
                </Paper>
              </Grid>

              <Grid item xs={12}>
                {events.length > 0 && <EventsTable events={events} detailed />}
              </Grid>
            </>
          )}
        </Grid>
      </Grow>

      {loading && <Loading />}
    </>
  )
}

export default ArchiveReport
