import { Grid, Grow } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import { ArchiveSingle } from 'api/archives.type'
import Loading from 'components-mui/Loading'
import ObjectConfiguration from 'components/ObjectConfiguration'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import T from 'components/T'
import api from 'api'
import { useParams } from 'react-router-dom'

const Single = () => {
  const { uuid } = useParams<{ uuid: string }>()

  const [loading, setLoading] = useState(true)
  const [detail, setDetail] = useState<ArchiveSingle>()

  const fetchDetail = () => api.archives.single(uuid).then(({ data }) => setDetail(data))

  useEffect(() => {
    Promise.all([fetchDetail()])
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
              <Grid item xs={6} sm={6} md={3}></Grid>
              <Grid item xs={6} sm={6} md={3}></Grid>

              <Grid item xs={12} md={6}></Grid>

              <Grid item xs={12}>
                <Paper>
                  <PaperTop title={T('common.configuration')} />
                  <ObjectConfiguration config={detail} />
                </Paper>
              </Grid>

              <Grid item xs={12}>
                {/* {events.length > 0 && <EventsTable events={events} />} */}
              </Grid>
            </>
          )}
        </Grid>
      </Grow>

      {loading && <Loading />}
    </>
  )
}

export default Single
