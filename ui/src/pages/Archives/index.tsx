import { Box, Grid, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import { Archive } from 'api/archives.type'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import ConfirmDialog from 'components/ConfirmDialog'
import ExperimentPaper from 'components/ExperimentPaper'
import Loading from 'components/Loading'
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
      <Grid container spacing={3}>
        {archives &&
          archives.length > 0 &&
          archives.map((a) => (
            <Grid key={a.name} item xs={12}>
              <ExperimentPaper experiment={a} isArchive handleSelect={setSelected} handleDialogOpen={setDialogOpen} />
            </Grid>
          ))}
      </Grid>

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
