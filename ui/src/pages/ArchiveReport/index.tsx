import { Typography, Grow } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import Loading from 'components/Loading'
import api from 'api'
import { useParams } from 'react-router-dom'
import { ExperimentDetail } from 'api/experiments.type'

const ArchiveReport: React.FC = () => {
  const { uuid } = useParams()

  const [loading, setLoading] = useState(true)
  const [detail, setDetail] = useState<ExperimentDetail>()
  const [report, setReport] = useState(null)

  const fetchDetail = () => api.archives.detail(uuid).then(({ data }) => setDetail(data))
  const fetchReport = () => api.archives.report(uuid).then(({ data }) => setReport(data))

  useEffect(() => {
    Promise.all([fetchDetail(), fetchReport()])
      .then((_) => setLoading(false))
      .catch(console.log)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <>
          {detail && (
            <>
              <Typography variant="h6">{detail.name}</Typography>
            </>
          )}
        </>
      </Grow>

      {loading && <Loading />}
    </>
  )
}

export default ArchiveReport
