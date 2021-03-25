import { Box, Button, Checkbox, Grid, Typography } from '@material-ui/core'
import { useEffect, useState } from 'react'

import { Archive } from 'api/archives.type'
import ConfirmDialog from 'components-mui/ConfirmDialog'
import DeleteOutlinedIcon from '@material-ui/icons/DeleteOutlined'
import ExperimentListItem from 'components/ExperimentListItem'
import FilterListIcon from '@material-ui/icons/FilterList'
import Loading from 'components-mui/Loading'
import NotFound from 'components-mui/NotFound'
import PlaylistAddCheckIcon from '@material-ui/icons/PlaylistAddCheck'
import Space from 'components-mui/Space'
import T from 'components/T'
import _groupBy from 'lodash.groupby'
import api from 'api'
import { setAlert } from 'slices/globalStatus'
import { transByKind } from 'lib/byKind'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

export default function Archives() {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(false)
  const [archives, setArchives] = useState<Archive[]>([])
  const [dialogOpen, setDialogOpen] = useState(false)
  const [selected, setSelected] = useState({
    uuid: '',
    title: '',
    description: '',
    action: 'archive',
  })
  const [batch, setBatch] = useState<Record<uuid, boolean>>({})
  const batchLength = Object.keys(batch).length
  const isBatchEmpty = batchLength === 0

  const fetchArchives = () => {
    setLoading(true)

    api.archives
      .archives()
      .then(({ data }) => setArchives(data))
      .catch(console.error)
      .finally(() => setLoading(false))
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

  const handleBatchSelect = () => setBatch(isBatchEmpty ? { [archives[0].uid]: true } : {})

  const handleBatchSelectAll = () =>
    setBatch(
      batchLength < archives.length
        ? archives.reduce<Record<uuid, boolean>>((acc, d) => {
            acc[d.uid] = true

            return acc
          }, {})
        : {}
    )

  const handleBatchDelete = () => {}

  const onCheckboxChange = (uuid: uuid) => (e: React.ChangeEvent<HTMLInputElement>) => {
    const newBatch = batch

    newBatch[uuid] = e.target.checked

    setBatch(newBatch)
  }

  return (
    <>
      <Box mb={6} textAlign="right">
        <Space>
          {!isBatchEmpty && (
            <>
              <Button
                variant="outlined"
                color="secondary"
                startIcon={<DeleteOutlinedIcon />}
                onClick={handleBatchDelete}
              >
                {T('common.delete')}
              </Button>
              <Button variant="outlined" startIcon={<PlaylistAddCheckIcon />} onClick={handleBatchSelectAll}>
                {T('common.selectAll')}
              </Button>
            </>
          )}
          <Button
            variant="outlined"
            startIcon={<FilterListIcon />}
            onClick={handleBatchSelect}
            disabled={archives.length === 0}
          >
            {T(`common.${isBatchEmpty ? 'batchOperation' : 'cancel'}`)}
          </Button>
        </Space>
      </Box>

      {archives &&
        archives.length > 0 &&
        Object.entries(_groupBy(archives, 'kind')).map(([kind, archivesByKind]) => (
          <Box key={kind} mb={6}>
            <Box mb={3} ml={1}>
              <Typography variant="overline">{transByKind(kind as any)}</Typography>
            </Box>
            <Grid container spacing={6}>
              {archivesByKind.length > 0 &&
                archivesByKind.map((e) => (
                  <Grid key={e.uid} item xs={12}>
                    <Box display="flex">
                      <Box flex={1}>
                        <ExperimentListItem
                          experiment={e}
                          isArchive
                          handleSelect={setSelected}
                          handleDialogOpen={setDialogOpen}
                          intl={intl}
                        />
                      </Box>
                      {!isBatchEmpty && (
                        <Checkbox
                          color="primary"
                          checked={batch[e.uid] === true}
                          onChange={onCheckboxChange(e.uid)}
                          style={{ width: 56 }}
                        />
                      )}
                    </Box>
                  </Grid>
                ))}
            </Grid>
          </Box>
        ))}

      {!loading && archives.length === 0 && (
        <NotFound illustrated textAlign="center">
          <Typography>{T('archives.noArchivesFound')}</Typography>
        </NotFound>
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
