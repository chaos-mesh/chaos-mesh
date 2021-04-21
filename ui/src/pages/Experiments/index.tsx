import { Box, Button, Checkbox, Typography } from '@material-ui/core'
import ConfirmDialog, { ConfirmDialogHandles } from 'components-mui/ConfirmDialog'
import { FixedSizeList as RWList, ListChildComponentProps as RWListChildComponentProps } from 'react-window'
import { useEffect, useRef, useState } from 'react'

import AddIcon from '@material-ui/icons/Add'
import CloseIcon from '@material-ui/icons/Close'
import DeleteOutlinedIcon from '@material-ui/icons/DeleteOutlined'
import { Experiment } from 'api/experiments.type'
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
import { styled } from '@material-ui/core/styles'
import { transByKind } from 'lib/byKind'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

const StyledCheckBox = styled(Checkbox)({
  position: 'relative',
  left: -11,
  paddingRight: 0,
  '&:hover': {
    background: 'none !important',
  },
})

const initialSelected = {
  uuid: '',
  title: '',
  description: '',
  action: '',
}

export default function Experiments() {
  const intl = useIntl()
  const history = useHistory()

  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(true)
  const [experiments, setExperiments] = useState<Experiment[]>([])
  const [selected, setSelected] = useState(initialSelected)
  const [batch, setBatch] = useState<Record<uuid, boolean>>({})
  const batchLength = Object.keys(batch).length
  const isBatchEmpty = batchLength === 0
  const confirmRef = useRef<ConfirmDialogHandles>(null)

  const fetchExperiments = () => {
    api.experiments
      .experiments()
      .then(({ data }) => setExperiments(data))
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(fetchExperiments, [])

  const handleSelect = (selected: typeof initialSelected) => {
    setSelected(selected)

    confirmRef.current!.setOpen(true)
  }

  const handleAction = (action: string) => () => {
    const { uuid } = selected

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

    confirmRef.current!.setOpen(false)

    if (actionFunc) {
      actionFunc(arg)
        .then(() => {
          dispatch(
            setAlert({
              type: 'success',
              message: intl.formatMessage({ id: `common.${action}Successfully` }),
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
      uuid: '',
      title: `${intl.formatMessage({ id: 'experiments.deleteMulti' })}`,
      description: intl.formatMessage({ id: 'experiments.deleteDesc' }),
      action: 'archiveMulti',
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
        <ExperimentListItem experiment={data[index]} onSelect={handleSelect} intl={intl} />
      </Box>
    </Box>
  )

  return (
    <>
      <Space mb={6}>
        <Button variant="outlined" startIcon={<AddIcon />} onClick={() => history.push('/newExperiment')}>
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
            <Button variant="outlined" color="secondary" startIcon={<DeleteOutlinedIcon />} onClick={handleBatchDelete}>
              {T('common.delete')}
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
          <Typography>{T('experiments.noExperimentsFound')}</Typography>
        </NotFound>
      )}

      {loading && <Loading />}

      <ConfirmDialog
        ref={confirmRef}
        title={selected.title}
        description={selected.description}
        onConfirm={handleAction(selected.action)}
      />
    </>
  )
}
