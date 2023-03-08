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
import CloseIcon from '@mui/icons-material/Close'
import DeleteOutlinedIcon from '@mui/icons-material/DeleteOutlined'
import FilterListIcon from '@mui/icons-material/FilterList'
import PlaylistAddCheckIcon from '@mui/icons-material/PlaylistAddCheck'
import TabContext from '@mui/lab/TabContext'
import TabList from '@mui/lab/TabList'
import { Box, Button, Checkbox, Typography, styled } from '@mui/material'
import Tab from '@mui/material/Tab'
import _ from 'lodash'
import {
  useDeleteArchives,
  useDeleteArchivesSchedules,
  useDeleteArchivesSchedulesUid,
  useDeleteArchivesUid,
  useDeleteArchivesWorkflows,
  useDeleteArchivesWorkflowsUid,
  useGetArchives,
  useGetArchivesSchedules,
  useGetArchivesWorkflows,
} from 'openapi'
import {
  DeleteArchivesWorkflowsParams,
  DeleteExperimentsParams,
  DeleteSchedulesParams,
  TypesArchive,
} from 'openapi/index.schemas'
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
import { useQuery } from 'lib/hooks'

const StyledCheckBox = styled(Checkbox)({
  position: 'relative',
  left: -11,
  paddingRight: 0,
  '&:hover': {
    background: 'none !important',
  },
})

type PanelType = 'workflow' | 'schedule' | 'experiment'

export default function Archives() {
  const navigate = useNavigate()
  const intl = useIntl()
  const query = useQuery()
  let kind = query.get('kind') || 'experiment'

  const dispatch = useStoreDispatch()

  const [panel, setPanel] = useState<PanelType>(kind as PanelType)
  const [archives, setArchives] = useState<TypesArchive[]>([])
  const [batch, setBatch] = useState<Record<uuid, boolean>>({})
  const batchLength = Object.keys(batch).length
  const isBatchEmpty = batchLength === 0

  const { isLoading: loading1, refetch: refetchWorkflows } = useGetArchivesWorkflows(undefined, {
    query: { enabled: kind === 'workflow', onSuccess: setArchives },
  })
  const { isLoading: loading2, refetch: refetchSchedules } = useGetArchivesSchedules(undefined, {
    query: { enabled: kind === 'schedule', onSuccess: setArchives },
  })
  const { isLoading: loading3, refetch: refetchExperiments } = useGetArchives(undefined, {
    query: { enabled: kind === 'experiment', onSuccess: setArchives },
  })
  const loading = kind === 'workflow' ? loading1 : kind === 'schedule' ? loading2 : loading3
  function refetchByKind() {
    switch (kind) {
      case 'workflow':
        refetchWorkflows()

        break
      case 'schedule':
        refetchSchedules()

        break
      default:
        refetchExperiments()
    }
  }
  const { mutateAsync: deleteWorkflows } = useDeleteArchivesWorkflows()
  const { mutateAsync: deleteSchedules } = useDeleteArchivesSchedules()
  const { mutateAsync: deleteExperiments } = useDeleteArchives()
  const { mutateAsync: deleteWorkflowsByUUID } = useDeleteArchivesWorkflowsUid()
  const { mutateAsync: deleteSchedulesByUUID } = useDeleteArchivesSchedulesUid()
  const { mutateAsync: deleteExperimentsByUUID } = useDeleteArchivesUid()

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
    let arg:
      | { uid: string }
      | { params: DeleteArchivesWorkflowsParams | DeleteExperimentsParams | DeleteSchedulesParams }
      | undefined

    switch (action) {
      case 'delete':
        switch (kind) {
          case 'workflow':
            actionFunc = deleteWorkflowsByUUID
            break
          case 'schedule':
            actionFunc = deleteSchedulesByUUID
            break
          case 'experiment':
          default:
            actionFunc = deleteExperimentsByUUID
            break
        }
        arg = { uid: uuid! }

        break
      case 'deleteMulti':
        action = 'delete'
        switch (kind) {
          case 'workflow':
            actionFunc = deleteWorkflows
            break
          case 'schedule':
            actionFunc = deleteSchedules
            break
          case 'experiment':
          default:
            actionFunc = deleteExperiments
            break
        }
        arg = {
          params: {
            uids: Object.keys(batch)
              .filter((d) => batch[d] === true)
              .join(','),
          },
        }
        setBatch({})

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

          refetchByKind()
        })
        .catch(console.error)
    }
  }

  const handleBatchSelect = () => setBatch(isBatchEmpty ? { [archives[0].uid!]: true } : {})

  const handleBatchSelectAll = () =>
    setBatch(
      batchLength <= archives.length
        ? archives.reduce<Record<uuid, boolean>>((acc, d) => {
            acc[d.uid!] = true

            return acc
          }, {})
        : {}
    )

  const handleBatchDelete = () =>
    handleSelect({
      title: i18n('archives.deleteMulti', intl),
      description: i18n('archives.deleteDesc', intl),
      handle: handleAction('deleteMulti'),
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
        <ObjectListItem type="archive" archive={kind as any} data={data[index]} onSelect={onSelect} />
      </Box>
    </Box>
  )

  const onTabChange = (_: any, newValue: PanelType) => {
    navigate(`/archives?kind=${newValue}`)
    setPanel(newValue)
  }

  return (
    <TabContext value={panel}>
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <TabList onChange={onTabChange}>
          <Tab label={i18n('workflows.title')} value="workflow" />
          <Tab label={i18n('schedules.title')} value="schedule" />
          <Tab label={i18n('experiments.title')} value="experiment" />
        </TabList>
      </Box>

      <Space direction="row" my={6}>
        <Button
          variant="outlined"
          startIcon={isBatchEmpty ? <FilterListIcon /> : <CloseIcon />}
          onClick={handleBatchSelect}
          disabled={archives.length === 0}
        >
          {i18n(`common.${isBatchEmpty ? 'batchOperation' : 'cancel'}`)}
        </Button>
        {!isBatchEmpty && (
          <>
            <Button variant="outlined" startIcon={<PlaylistAddCheckIcon />} onClick={handleBatchSelectAll}>
              {i18n('common.selectAll')}
            </Button>
            <Button variant="outlined" color="secondary" startIcon={<DeleteOutlinedIcon />} onClick={handleBatchDelete}>
              {i18n('common.delete')}
            </Button>
          </>
        )}
      </Space>

      {archives.length > 0 &&
        Object.entries(_.groupBy(archives, 'kind')).map(([kind, archivesByKind]) => (
          <Box key={kind} mb={6}>
            <Typography variant="overline">{transByKind(kind as any)}</Typography>
            <RWList
              width="100%"
              height={archivesByKind.length > 3 ? 300 : archivesByKind.length * 70}
              itemCount={archivesByKind.length}
              itemSize={70}
              itemData={archivesByKind}
            >
              {Row}
            </RWList>
          </Box>
        ))}

      {!loading && archives.length === 0 && (
        <NotFound illustrated textAlign="center">
          <Typography>{i18n('archives.notFound')}</Typography>
        </NotFound>
      )}

      {loading && <Loading />}
    </TabContext>
  )
}
