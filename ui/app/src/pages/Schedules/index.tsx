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
import { Box, Button, Checkbox, styled } from '@mui/material'
import { Confirm, setAlert, setConfirm } from 'slices/globalStatus'
import { FixedSizeList as RWList, ListChildComponentProps as RWListChildComponentProps } from 'react-window'
import { useEffect, useState } from 'react'

import AddIcon from '@mui/icons-material/Add'
import CloseIcon from '@mui/icons-material/Close'
import DeleteOutlinedIcon from '@mui/icons-material/DeleteOutlined'
import FilterListIcon from '@mui/icons-material/FilterList'
import Loading from '@ui/mui-extends/esm/Loading'
import NotFound from 'components/NotFound'
import ObjectListItem from 'components/ObjectListItem'
import PlaylistAddCheckIcon from '@mui/icons-material/PlaylistAddCheck'
import { Schedule } from 'api/schedules.type'
import Space from '@ui/mui-extends/esm/Space'
import T from 'components/T'
import { Typography } from '@mui/material'
import _groupBy from 'lodash.groupby'
import api from 'api'
import { transByKind } from 'lib/byKind'
import { useIntl } from 'react-intl'
import { useNavigate } from 'react-router-dom'
import { useStoreDispatch } from 'store'

const StyledCheckBox = styled(Checkbox)({
  position: 'relative',
  left: -11,
  paddingRight: 0,
  '&:hover': {
    background: 'none !important',
  },
})

const Schedules = () => {
  const navigate = useNavigate()
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(true)
  const [schedules, setSchedules] = useState<Schedule[]>([])
  const [batch, setBatch] = useState<Record<uuid, boolean>>({})
  const batchLength = Object.keys(batch).length
  const isBatchEmpty = batchLength === 0

  const fetchSchedules = () => {
    api.schedules
      .schedules()
      .then(({ data }) => setSchedules(data))
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(fetchSchedules, [])

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
        actionFunc = api.schedules.del
        arg = uuid

        break
      case 'archiveMulti':
        action = 'archive'
        actionFunc = api.schedules.delMulti
        arg = Object.keys(batch)
        setBatch({})

        break
      case 'pause':
        actionFunc = api.schedules.pause
        arg = uuid

        break
      case 'start':
        actionFunc = api.schedules.start
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

          setTimeout(fetchSchedules, 300)
        })
        .catch(console.error)
    }
  }

  const handleBatchSelect = () => setBatch(isBatchEmpty ? { [schedules[0].uid]: true } : {})

  const handleBatchSelectAll = () =>
    setBatch(
      batchLength <= schedules.length
        ? schedules.reduce<Record<uuid, boolean>>((acc, d) => {
            acc[d.uid] = true

            return acc
          }, {})
        : {}
    )

  const handleBatchDelete = () =>
    handleSelect({
      title: T('schedules.deleteMulti', intl),
      description: T('schedules.deleteDesc', intl),
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
        <ObjectListItem type="schedule" data={data[index]} onSelect={onSelect} />
      </Box>
    </Box>
  )

  return (
    <>
      <Space direction="row" mb={6}>
        <Button variant="outlined" startIcon={<AddIcon />} onClick={() => navigate('/schedules/new')}>
          {T('newS.title')}
        </Button>
        <Button
          variant="outlined"
          startIcon={isBatchEmpty ? <FilterListIcon /> : <CloseIcon />}
          onClick={handleBatchSelect}
          disabled={schedules.length === 0}
        >
          {T(`common.${isBatchEmpty ? 'batchOperation' : 'cancel'}`)}
        </Button>
        {!isBatchEmpty && (
          <>
            <Button variant="outlined" startIcon={<PlaylistAddCheckIcon />} onClick={handleBatchSelectAll}>
              {T('common.selectAll')}
            </Button>
            <Button variant="outlined" color="secondary" startIcon={<DeleteOutlinedIcon />} onClick={handleBatchDelete}>
              {T('archives.single')}
            </Button>
          </>
        )}
      </Space>

      {schedules.length > 0 &&
        Object.entries(_groupBy(schedules, 'kind')).map(([type, schedulesByType]) => (
          <Box key={type} mb={6}>
            <Typography variant="overline">{transByKind(type as any)}</Typography>
            <RWList
              width="100%"
              height={schedulesByType.length > 3 ? 300 : schedulesByType.length * 70}
              itemCount={schedulesByType.length}
              itemSize={70}
              itemData={schedulesByType}
            >
              {Row}
            </RWList>
          </Box>
        ))}

      {!loading && schedules.length === 0 && (
        <NotFound illustrated textAlign="center">
          <Typography>{T('schedules.notFound')}</Typography>
        </NotFound>
      )}

      {loading && <Loading />}
    </>
  )
}

export default Schedules
