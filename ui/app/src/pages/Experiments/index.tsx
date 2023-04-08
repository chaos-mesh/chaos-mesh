/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import AddIcon from '@mui/icons-material/Add'
import ArchiveOutlinedIcon from '@mui/icons-material/ArchiveOutlined'
import CloseIcon from '@mui/icons-material/Close'
import FilterListIcon from '@mui/icons-material/FilterList'
import PlaylistAddCheckIcon from '@mui/icons-material/PlaylistAddCheck'
import { Box, Button, Checkbox, Typography, styled } from '@mui/material'
import _ from 'lodash'
import {
  useDeleteExperiments,
  useDeleteExperimentsUid,
  useGetExperiments,
  usePutExperimentsPauseUid,
  usePutExperimentsStartUid,
} from 'openapi'
import { DeleteExperimentsParams } from 'openapi/index.schemas'
import { useState } from 'react'
import { useIntl } from 'react-intl'
import { useNavigate } from 'react-router-dom'
import { FixedSizeList as RWList, ListChildComponentProps as RWListChildComponentProps } from 'react-window'

import Loading from '@ui/mui-extends/esm/Loading'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch } from 'store'

import { Confirm, setAlert, setConfirm } from 'slices/globalStatus'

import NotFound from 'components/NotFound'
import ObjectListItem from 'components/ObjectListItem'
import i18n from 'components/T'

import { transByKind } from 'lib/byKind'

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
  const navigate = useNavigate()

  const dispatch = useStoreDispatch()

  const [batch, setBatch] = useState<Record<uuid, boolean>>({})
  const batchLength = Object.keys(batch).length
  const isBatchEmpty = batchLength === 0

  const { data: experiments, isLoading: loading, refetch } = useGetExperiments()
  const { mutateAsync: deleteExperimentsByUUID } = useDeleteExperimentsUid()
  const { mutateAsync: deleteExperiments } = useDeleteExperiments()
  const { mutateAsync: pauseExperiments } = usePutExperimentsPauseUid()
  const { mutateAsync: startExperiments } = usePutExperimentsStartUid()

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
    let actionFunc
    let arg: { uid: string } | { params: DeleteExperimentsParams } | undefined

    switch (action) {
      case 'archive':
        actionFunc = deleteExperimentsByUUID
        arg = { uid: uuid! }

        break
      case 'archiveMulti':
        action = 'archive'
        actionFunc = deleteExperiments
        arg = { params: { uids: Object.keys(batch).join(',') } }
        setBatch({})

        break
      case 'pause':
        actionFunc = pauseExperiments
        arg = { uid: uuid! }

        break
      case 'start':
        actionFunc = startExperiments
        arg = { uid: uuid! }

        break
    }

    if (actionFunc) {
      actionFunc(arg as any)
        .then(() => {
          dispatch(
            setAlert({
              type: 'success',
              message: i18n(`confirm.success.${action}`, intl),
            })
          )

          refetch()
        })
        .catch(console.error)
    }
  }

  const handleBatchSelect = () => setBatch(isBatchEmpty ? { [experiments![0].uid!]: true } : {})

  const handleBatchSelectAll = () =>
    setBatch(
      batchLength <= experiments!.length
        ? experiments!.reduce<Record<uuid, boolean>>((acc, d) => {
            acc[d.uid!] = true

            return acc
          }, {})
        : {}
    )

  const handleBatchDelete = () =>
    handleSelect({
      title: i18n('experiments.deleteMulti', intl),
      description: i18n('experiments.deleteDesc', intl),
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
        <Button variant="outlined" startIcon={<AddIcon />} onClick={() => navigate('/experiments/new')}>
          {i18n('newE.title')}
        </Button>
        <Button
          variant="outlined"
          startIcon={isBatchEmpty ? <FilterListIcon /> : <CloseIcon />}
          onClick={handleBatchSelect}
          disabled={experiments?.length === 0}
        >
          {i18n(`common.${isBatchEmpty ? 'batchOperation' : 'cancel'}`)}
        </Button>
        {!isBatchEmpty && (
          <>
            <Button variant="outlined" startIcon={<PlaylistAddCheckIcon />} onClick={handleBatchSelectAll}>
              {i18n('common.selectAll')}
            </Button>
            <Button
              variant="outlined"
              color="secondary"
              startIcon={<ArchiveOutlinedIcon />}
              onClick={handleBatchDelete}
            >
              {i18n('archives.single')}
            </Button>
          </>
        )}
      </Space>

      {experiments &&
        experiments.length > 0 &&
        Object.entries(_.groupBy(experiments, 'kind')).map(([kind, experimentsByKind]) => (
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

      {!loading && experiments?.length === 0 && (
        <NotFound illustrated textAlign="center">
          <Typography>{i18n('experiments.notFound')}</Typography>
        </NotFound>
      )}

      {loading && <Loading />}
    </>
  )
}
