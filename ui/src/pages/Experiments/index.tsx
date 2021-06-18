import { Box, Button, Checkbox, Typography } from '@material-ui/core'
import { Confirm, setAlert, setConfirm } from 'slices/globalStatus'
import { FixedSizeList as RWList, ListChildComponentProps as RWListChildComponentProps } from 'react-window'

import AddIcon from '@material-ui/icons/Add'
import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import CloseIcon from '@material-ui/icons/Close'
import { Experiment } from 'api/experiments.type'
import FilterListIcon from '@material-ui/icons/FilterList'
import Loading from 'components-mui/Loading'
import NotFound from 'components-mui/NotFound'
import ObjectListItem from 'components/ObjectListItem'
import PlaylistAddCheckIcon from '@material-ui/icons/PlaylistAddCheck'
import Space from 'components-mui/Space'
import T from 'components/T'
import _groupBy from 'lodash.groupby'
import api from 'api'
import { styled } from '@material-ui/styles'
import { transByKind } from 'lib/byKind'
import { useHistory } from 'react-router-dom'
import { useIntervalFetch } from 'lib/hooks'
import { useIntl } from 'react-intl'
import { useState } from 'react'
import { useStoreDispatch } from 'store'

const StyledCheckBox = styled(Checkbox)({
  position: 'relative',
  left: -11,
  paddingRight: 0,
  '&:hover': {
    background: 'none !important',
  },
})

export default function Experiments() {
  const intl = useIntl()
  const history = useHistory()

  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(true)
  const [experiments, setExperiments] = useState<Experiment[]>([])
  const [batch, setBatch] = useState<Record<uuid, boolean>>({})
  const batchLength = Object.keys(batch).length
  const isBatchEmpty = batchLength === 0

  const fetchExperiments = (intervalID?: number) => {
    api.experiments
      .experiments()
      .then(({ data }) => {
        setExperiments(data)

        if (data.every((d) => d.status === 'finished')) {
          clearInterval(intervalID)
        }
      })
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useIntervalFetch(fetchExperiments)

  const handleSelect = (selected: Confirm) => dispatch(setConfirm(selected))
  const onSelect = (selected: Confirm) =>
    dispatch(
      setConfirm({
        title: selected.title,
        description: selected.description,
        handle: handleAction(selected.action, selected.uuid),
      })
    )

  const handleAction = (action: string, uuid?: uuid) => () => {
    let actionFunc: any
    let arg: any

    switch (action) {
      case 'archive':
        actionFunc = api.experiments.del
        arg = uuid

        break
      case 'archiveMulti':
        action = 'archive'
        actionFunc = api.experiments.delMulti
        arg = Object.keys(batch)
        setBatch({})

        break
      case 'pause':
        actionFunc = api.experiments.pause
        arg = uuid

        break
      case 'start':
        actionFunc = api.experiments.start
        arg = uuid

        break
    }

    if (actionFunc) {
      actionFunc(arg)
        .then(() => {
          dispatch(
            setAlert({
              type: 'success',
              message: T(`confirm.success.${action}`, intl),
            })
          )

          setTimeout(fetchExperiments, 300)
        })
        .catch(console.error)
    }
  }

  const handleBatchSelect = () => setBatch(isBatchEmpty ? { [experiments[0].uid]: true } : {})

  const handleBatchSelectAll = () =>
    setBatch(
      batchLength <= experiments.length
        ? experiments.reduce<Record<uuid, boolean>>((acc, d) => {
            acc[d.uid] = true

            return acc
          }, {})
        : {}
    )

  const handleBatchDelete = () =>
    handleSelect({
      title: T('experiments.deleteMulti', intl),
      description: T('experiments.deleteDesc', intl),
      handle: handleAction('archiveMulti'),
    })

  const onCheckboxChange = (uuid: uuid) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setBatch({
      ...batch,
      [uuid]: e.target.checked,
    })
  }

  const Row = ({ data, index, style }: RWListChildComponentProps) => (
    <Box display="flex" alignItems="center" mb={3} style={style}>
      {!isBatchEmpty && (
        <StyledCheckBox
          color="primary"
          checked={batch[data[index].uid] === true}
          onChange={onCheckboxChange(data[index].uid)}
          disableRipple
        />
      )}
      <Box flex={1}>
        <ObjectListItem data={data[index]} onSelect={onSelect} />
      </Box>
    </Box>
  )

  return (
    <>
      <Space direction="row" mb={6}>
        <Button variant="outlined" startIcon={<AddIcon />} onClick={() => history.push('/experiments/new')}>
          {T('newE.title')}
        </Button>
        <Button
          variant="outlined"
          startIcon={isBatchEmpty ? <FilterListIcon /> : <CloseIcon />}
          onClick={handleBatchSelect}
          disabled={experiments.length === 0}
        >
          {T(`common.${isBatchEmpty ? 'batchOperation' : 'cancel'}`)}
        </Button>
        {!isBatchEmpty && (
          <>
            <Button variant="outlined" startIcon={<PlaylistAddCheckIcon />} onClick={handleBatchSelectAll}>
              {T('common.selectAll')}
            </Button>
            <Button
              variant="outlined"
              color="secondary"
              startIcon={<ArchiveOutlinedIcon />}
              onClick={handleBatchDelete}
            >
              {T('archives.single')}
            </Button>
          </>
        )}
      </Space>

      {experiments.length > 0 &&
        Object.entries(_groupBy(experiments, 'kind')).map(([kind, experimentsByKind]) => (
          <Box key={kind} mb={6}>
            <Typography variant="overline">{transByKind(kind as any)}</Typography>
            <RWList
              width="100%"
              height={experimentsByKind.length > 3 ? 300 : experimentsByKind.length * 70}
              itemCount={experimentsByKind.length}
              itemSize={70}
              itemData={experimentsByKind}
            >
              {Row}
            </RWList>
          </Box>
        ))}

      {!loading && experiments.length === 0 && (
        <NotFound illustrated textAlign="center">
          <Typography>{T('experiments.notFound')}</Typography>
        </NotFound>
      )}

      {loading && <Loading />}
    </>
  )
}
