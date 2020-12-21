import { Box, Grid, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import { Archive } from 'api/archives.type'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import ExperimentListItem from 'components/ExperimentListItem'
import Loading from 'components-mui/Loading'
import T from 'components/T'
import _groupBy from 'lodash.groupby'
import api from 'api'
import { useIntl } from 'react-intl'

export default function Archives() {
  const intl = useIntl()

  const [loading, setLoading] = useState(false)
  const [archives, setArchives] = useState<Archive[] | null>(null)

  const fetchArchives = () => {
    setLoading(true)

    api.archives
      .archives()
      .then(({ data }) => setArchives(data))
      .catch(console.error)
      .finally(() => {
        setLoading(false)
      })
  }

  useEffect(fetchArchives, [])

  return (
    <>
      {archives &&
        archives.length > 0 &&
        Object.entries(_groupBy(archives, 'kind'))
          .sort((a, b) => (a[0] > b[0] ? 1 : -1))
          .map(([kind, archivesByKind]) => (
            <Box key={kind} mb={6}>
              <Box mb={6}>
                <Typography variant="button">{kind}</Typography>
              </Box>
              <Grid container spacing={3}>
                {archivesByKind.length > 0 &&
                  archivesByKind.map((e) => (
                    <Grid key={e.uid} item xs={12}>
                      <ExperimentListItem
                        experiment={e}
                        isArchive
                        handleSelect={() => {}}
                        handleDialogOpen={() => {}}
                        intl={intl}
                      />
                    </Grid>
                  ))}
              </Grid>
            </Box>
          ))}

      {!loading && archives && archives.length === 0 && (
        <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" height="100%">
          <Box mb={3}>
            <ArchiveOutlinedIcon fontSize="large" />
          </Box>
          <Typography variant="h6" align="center">
            {T('archives.no_archives_found')}
          </Typography>
        </Box>
      )}

      {loading && <Loading />}
    </>
  )
}
