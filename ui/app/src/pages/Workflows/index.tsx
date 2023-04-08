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
import RemoveCircleOutlineIcon from '@mui/icons-material/RemoveCircleOutline'
import ReplayIcon from '@mui/icons-material/Replay'
import { Box, Button, Grow, Typography } from '@mui/material'
import type { ButtonProps } from '@mui/material'
import type { GridColDef, GridRenderCellParams, GridRowParams } from '@mui/x-data-grid'
import { GridActionsCellItem } from '@mui/x-data-grid'
import _ from 'lodash'
import { getWorkflowsUid, useDeleteWorkflowsUid, useGetWorkflows, usePostWorkflows } from 'openapi'
import { CoreWorkflowMeta } from 'openapi/index.schemas'
import React, { useState } from 'react'
import { useIntl } from 'react-intl'
import { useNavigate } from 'react-router-dom'
import { v4 as uuidv4 } from 'uuid'

import Loading from '@ui/mui-extends/esm/Loading'
import PaperTop from '@ui/mui-extends/esm/PaperTop'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch, useStoreSelector } from 'store'

import { setAlert, setConfirm } from 'slices/globalStatus'

import DataTable from 'components/DataTable'
import NotFound from 'components/NotFound'
import StatusLabel from 'components/StatusLabel'
import i18n, { T } from 'components/T'

import { comparator, format, toRelative } from 'lib/luxon'

function transformWorkflows(data: CoreWorkflowMeta[]) {
  return data
    .map((d) => ({
      ...d,
      time: d.status === 'finished' ? 'Ended at: ' + format(d.end_time!) : 'Created at: ' + toRelative(d.created_at!),
    }))
    .sort((a, b) => comparator(b.created_at!, a.created_at!))
}

const Workflows = () => {
  const navigate = useNavigate()
  const intl = useIntl()

  const [loading, setLoading] = useState(true)

  const { useNextWorkflowInterface } = useStoreSelector((state) => state.settings)
  const dispatch = useStoreDispatch()

  const { data: workflows, refetch } = useGetWorkflows(undefined, {
    query: {
      select: transformWorkflows,
      onSettled() {
        setLoading(false)
      },
    },
  })
  const { mutateAsync: deleteWorkflows } = useDeleteWorkflowsUid()
  const { mutateAsync: createWorkflows } = usePostWorkflows()

  const NewWorkflow = (props: ButtonProps) => (
    <Button
      variant="contained"
      size="small"
      startIcon={<AddIcon />}
      onClick={() => navigate(useNextWorkflowInterface ? '/workflows/new/next' : '/workflows/new')}
      {...props}
    >
      <T id="newW.title" />
    </Button>
  )

  const handleAction = (action: string, uuid: uuid) => () => {
    let actionFunc

    switch (action) {
      case 'archive':
        actionFunc = deleteWorkflows

        break
      default:
        break
    }

    if (actionFunc) {
      actionFunc({ uid: uuid })
        .then(() => {
          dispatch(
            setAlert({
              type: 'success',
              message: <T id={`confirm.success.${action}`} />,
            })
          )

          refetch()
        })
        .catch(console.error)
    }
  }

  const handleDelete =
    ({ uid, name }: CoreWorkflowMeta) =>
    (e: React.SyntheticEvent) => {
      e.stopPropagation()

      dispatch(
        setConfirm({
          title: `${i18n('archives.single', intl)} ${name}`,
          description: <T id="workflows.deleteDesc" />,
          handle: handleAction('archive', uid!),
        })
      )
    }

  const handleReRun = (uid: uuid) => async (e: React.SyntheticEvent) => {
    e.stopPropagation()

    const { name, kube_object } = await getWorkflowsUid(uid)

    dispatch(
      setConfirm({
        title: `Re-run ${name}`,
        description: 'This will re-create a new workflow with the same configuration.',
        handle: () => {
          createWorkflows({
            data: {
              apiVersion: 'chaos-mesh.org/v1alpha1',
              kind: 'Workflow',
              metadata: {
                ...kube_object!.metadata,
                name: `${name}-${uuidv4()}`,
              },
              spec: kube_object!.spec,
            } as any,
          })
            .then(() => {
              dispatch(
                setAlert({
                  type: 'success',
                  message: <T id="confirm.success.create" />,
                })
              )

              refetch()
            })
            .catch(console.error)
        },
      })
    )
  }

  const jumpToSingleWorkflow = ({ row }: GridRowParams) => navigate(`/workflows/${row.uid}`)

  const columns: GridColDef[] = [
    {
      field: 'status',
      headerName: 'Status',
      width: 150,
      renderCell: ({ value }: GridRenderCellParams) => <StatusLabel status={value} />,
    },
    {
      field: 'name',
      headerName: 'Name',
      renderCell: ({ value }: GridRenderCellParams) => <Typography variant="body2">{_.truncate(value)}</Typography>,
    },
    { field: 'time', headerName: 'Time' },
    {
      field: 'actions',
      type: 'actions',
      headerName: 'Operations',
      align: 'left',
      width: 150,
      getActions: ({ row }: GridRowParams) => [
        <GridActionsCellItem icon={<RemoveCircleOutlineIcon />} label="Archive" onClick={handleDelete(row)} />,
        ...(row.status === 'finished'
          ? [<GridActionsCellItem icon={<ReplayIcon />} label="Re-run" onClick={handleReRun(row.uid)} />]
          : []),
      ],
    },
  ]

  return (
    <>
      <Grow in={!loading} style={{ transformOrigin: '0 0 0' }}>
        <div style={{ height: '100%' }}>
          {workflows && workflows.length > 0 ? (
            <Space>
              <Box display="flex" justifyContent="space-between" alignItems="center">
                <PaperTop title="All Workflows" subtitle="Manage your workflows." h1 divider>
                  <NewWorkflow />
                </PaperTop>
              </Box>
              <DataTable columns={columns} rows={workflows} onRowClick={jumpToSingleWorkflow} />
            </Space>
          ) : (
            <NotFound illustrated>
              <Typography fontWeight="medium">
                <T id="workflows.notFound" />
              </Typography>
              <NewWorkflow sx={{ mt: 3 }} />
            </NotFound>
          )}
        </div>
      </Grow>

      {loading && <Loading />}
    </>
  )
}

export default Workflows
