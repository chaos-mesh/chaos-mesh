import { Box, Breadcrumbs, Button, Checkbox, LinkProps, Link as MUILink, Typography } from '@material-ui/core'
import { Confirm, setAlert, setConfirm } from 'slices/globalStatus'
import { FixedSizeList as RWList, ListChildComponentProps as RWListChildComponentProps } from 'react-window'
import { useCallback, useEffect, useState } from 'react'

import { Archive } from 'api/archives.type'
import CloseIcon from '@material-ui/icons/Close'
import DeleteOutlinedIcon from '@material-ui/icons/DeleteOutlined'
import FilterListIcon from '@material-ui/icons/FilterList'
import { Link } from 'react-router-dom'
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
import { useIntl } from 'react-intl'
import { useQuery } from 'lib/hooks'
import { useStoreDispatch } from 'store'

const StyledCheckBox = styled(Checkbox)({
  position: 'relative',
  left: -11,
  paddingRight: 0,
  '&:hover': {
    background: 'none !important',
  },
})

export default function Archives() {
  const intl = useIntl()
  const query = useQuery()
  let kind = query.get('kind') || 'experiment'

  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(true)
  const [archives, setArchives] = useState<Archive[]>([])
  const [batch, setBatch] = useState<Record<uuid, boolean>>({})
  const batchLength = Object.keys(batch).length
  const isBatchEmpty = batchLength === 0

  const fetchArchives = useCallback(() => {
    let request
    switch (kind) {
      case 'schedule':
        request = api.schedules.archives
        break
      case 'experiment':
      default:
        request = api.archives.archives
        break
    }

    request()
      .then(({ data }) => {
        setArchives(data)
      })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [kind])

  useEffect(fetchArchives, [fetchArchives])

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
      case 'delete':
        actionFunc = api.archives.del
        arg = uuid

        break
      case 'deleteMulti':
        action = 'archive'
        actionFunc = api.archives.delMulti
        arg = Object.keys(batch)
        setBatch({})

        break
    }

    if (actionFunc) {
      actionFunc(arg)
        .then(() => {
          dispatch(
            setAlert({
              type: 'success',
              message: intl.formatMessage({ id: `confirm.${action}Successfully` }),
            })
          )

          fetchArchives()
        })
        .catch(console.error)
    }
  }

  const handleBatchSelect = () => setBatch(isBatchEmpty ? { [archives[0].uid]: true } : {})

  const handleBatchSelectAll = () =>
    setBatch(
      batchLength <= archives.length
        ? archives.reduce<Record<uuid, boolean>>((acc, d) => {
            acc[d.uid] = true

            return acc
          }, {})
        : {}
    )

  const handleBatchDelete = () =>
    handleSelect({
      title: `${intl.formatMessage({ id: 'archives.deleteMulti' })}`,
      description: intl.formatMessage({ id: 'archives.deleteDesc' }),
      handle: handleAction('deleteMulti'),
    })

  const onCheckboxChange = (uuid: uuid) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setBatch({
      ...batch,
      [uuid]: e.target.checked,
    })
  }

  const ActiveLink = ({ href, children }: LinkProps) => (
    <MUILink
      component={Link}
      color={kind === href ? 'primary' : kind === undefined && href === 'experiment' ? 'primary' : 'inherit'}
      to={`/archives?kind=${href}`}
    >
      {children}
    </MUILink>
  )

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

  return (
    <>
      <Breadcrumbs aria-label="breadcrumb">
        <ActiveLink href="workflow">{T('workflows.title')}</ActiveLink>
        <ActiveLink href="schedule">{T('schedules.title')}</ActiveLink>
        <ActiveLink href="experiment">{T('experiments.title')}</ActiveLink>
      </Breadcrumbs>

      <Space direction="row" my={6}>
        <Button
          variant="outlined"
          startIcon={isBatchEmpty ? <FilterListIcon /> : <CloseIcon />}
          onClick={handleBatchSelect}
          disabled={archives.length === 0}
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

      {archives.length > 0 &&
        Object.entries(_groupBy(archives, 'kind')).map(([kind, archivesByKind]) => (
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
          <Typography>{T('archives.notFound')}</Typography>
        </NotFound>
      )}

      {loading && <Loading />}
    </>
  )
}
