import { Box, Button, Grid, Grow } from '@material-ui/core'
import { useCallback, useEffect, useState } from 'react'

import { ArchiveSingle } from 'api/archives.type'
import CloudDownloadOutlinedIcon from '@material-ui/icons/CloudDownloadOutlined'
import { Event } from 'api/events.type'
import EventsTimeline from 'components/EventsTimeline'
import Loading from 'components-mui/Loading'
import ObjectConfiguration from 'components/ObjectConfiguration'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import T from 'components/T'
import api from 'api'
import fileDownload from 'js-file-download'
import loadable from '@loadable/component'
import { useParams } from 'react-router-dom'
import { useQuery } from 'lib/hooks'
import yaml from 'js-yaml'

const YAMLEditor = loadable(() => import('components/YAMLEditor'))

const Single = () => {
  const { uuid } = useParams<{ uuid: string }>()
  const query = useQuery()
  let kind = query.get('kind') || 'experiment'

  const [loading, setLoading] = useState(true)
  const [single, setSingle] = useState<ArchiveSingle>()
  const [events, setEvents] = useState<Event[]>([])

  const fetchSingle = useCallback(() => {
    let request
    switch (kind) {
      case 'workflow':
        request = api.workflows.singleArchive
        break
      case 'schedule':
        request = api.schedules.singleArchive
        break
      case 'experiment':
      default:
        request = api.archives.single
        break
    }

    request(uuid)
      .then(({ data }) => {
        setSingle(data)
      })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [uuid, kind])

  useEffect(fetchSingle, [fetchSingle])

  useEffect(() => {
    const fetchEvents = () => {
      api.events
        .events({ object_id: uuid, limit: 999 })
        .then(({ data }) => setEvents(data))
        .catch(console.error)
        .finally(() => {
          setLoading(false)
        })
    }

    if (single) {
      fetchEvents()
    }
  }, [uuid, single])

  const handleDownloadExperiment = () => fileDownload(yaml.dump(single!.kube_object), `${single!.name}.yaml`)

  const YAML = () => (
    <Paper sx={{ height: 600, p: 0 }}>
      {single && (
        <Box display="flex" flexDirection="column" height="100%">
          <PaperTop title={T('common.definition')} boxProps={{ p: 4.5, pb: 3 }}>
            <Space direction="row">
              <Button
                variant="outlined"
                size="small"
                startIcon={<CloudDownloadOutlinedIcon />}
                onClick={handleDownloadExperiment}
              >
                {T('common.download')}
              </Button>
            </Space>
          </PaperTop>
          <Box flex={1}>
            <YAMLEditor data={yaml.dump(single.kube_object)} aceProps={{ readOnly: true }} />
          </Box>
        </Box>
      )}
    </Paper>
  )

  return (
    <>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <div>
          {kind !== 'workflow' ? (
            <Space spacing={6}>
              <Paper>
                <PaperTop title={T('common.configuration')} boxProps={{ mb: 3 }} />
                {single && <ObjectConfiguration config={single} inSchedule={kind === 'schedule'} />}
              </Paper>

              <Grid container>
                <Grid item xs={12} lg={6} sx={{ pr: 3 }}>
                  <Paper sx={{ display: 'flex', flexDirection: 'column', height: 600 }}>
                    <PaperTop title={T('events.title')} boxProps={{ mb: 3 }} />
                    <Box flex={1} overflow="scroll">
                      <EventsTimeline events={events} />
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
