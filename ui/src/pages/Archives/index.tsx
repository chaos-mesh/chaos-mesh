import { Box, Grid, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import { Archive } from 'api/archives.type'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import ConfirmDialog from 'components/ConfirmDialog'
import ExperimentPaper from 'components/ExperimentPaper'
import Loading from 'components/Loading'
import _groupBy from 'lodash.groupby'
import api from 'api'

export default function Archives() {
  const [loading, setLoading] = useState(false)
  const [archives, setArchives] = useState<Archive[] | null>(null)
  const [selected, setSelected] = useState({
    uuid: '',
    title: '',
    description: '',
    action: 'recover',
  })
  const [dialogOpen, setDialogOpen] = useState(false)

  const fetchArchives = () => {
    setLoading(true)

    api.archives
      .archives()
      .then(({ data }) => setArchives(data))
      .catch(console.log)
      .finally(() => {
        setLoading(false)
      })
  }

  useEffect(fetchArchives, [])

  const handleArchive = (action: string) => () => {
    switch (action) {
      case 'recover':
        break

      default:
        break
    }
  }

  return (
    <>
      {archives &&
        archives.length > 0 &&
        Object.entries(_groupBy(archives, 'kind'))
          .sort((a, b) => (a[0] > b[0] ? 1 : -1))
          .map(([kind, archivesByKind]) => (
            <Box mb={6}>
              <Box mb={6}>
                <Typography variant="button">{kind}</Typography>
              </Box>
              <Grid container spacing={3}>
                {archivesByKind.length > 0 &&
                  archivesByKind.map((e) => (
                    <Grid key={e.name} item xs={12}>
                      <ExperimentPaper
                        experiment={e}
                        isArchive
                        handleSelect={setSelected}
                        handleDialogOpen={setDialogOpen}
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
            No archives found.
          </Typography>
        </Box>
      )}

      {loading && <Loading />}

      <ConfirmDialog
        open={dialogOpen}
        setOpen={setDialogOpen}
        title={selected.title}
        description={selected.description}
        handleConfirm={handleArchive(selected.action)}
      />
    </>
  )
}
