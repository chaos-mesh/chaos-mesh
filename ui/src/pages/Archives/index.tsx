import { Box, Grid, Typography } from '@material-ui/core'
import React, { useEffect, useState } from 'react'

import { Archive } from 'api/archives.type'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import ConfirmDialog from 'components-mui/ConfirmDialog'
import ExperimentListItem from 'components/ExperimentListItem'
import Loading from 'components-mui/Loading'
import T from 'components/T'
import _groupBy from 'lodash.groupby'
import api from 'api'
import { setAlert } from 'slices/globalStatus'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

export default function Archives() {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(false)
  const [archives, setArchives] = useState<Archive[] | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [selected, setSelected] = useState({
    uuid: '',
    title: '',
    description: '',
    action: 'archive',
  })

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

  const handleExperiment = (action: string) => () => {
    let actionFunc: any

    switch (action) {
      case 'delete':
        actionFunc = api.archives.del

        break
      default:
        actionFunc = null
    }

    if (actionFunc === null) {
      return
    }

    setDialogOpen(false)

    const { uuid } = selected

    actionFunc(uuid)
      .then(() => {
        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: `common.${action}Successfully` }),
          })
        )

        fetchArchives()
      })
      .catch(console.error)
  }

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
                        handleSelect={setSelected}
                        handleDialogOpen={setDialogOpen}
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
            {T('archives.noArchivesFound')}
          </Typography>
        </Box>
      )}

      {loading && <Loading />}

      <ConfirmDialog
        open={dialogOpen}
        setOpen={setDialogOpen}
        title={selected.title}
        description={selected.description}
        onConfirm={handleExperiment(selected.action)}
      />
    </>
  )
}
