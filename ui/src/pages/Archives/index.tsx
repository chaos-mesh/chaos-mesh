import { Box, Button, Checkbox, Typography } from '@material-ui/core'
import { Confirm, setAlert, setConfirm } from 'slices/globalStatus'
import { FixedSizeList as RWList, ListChildComponentProps as RWListChildComponentProps } from 'react-window'
import { useCallback, useEffect, useState } from 'react'

import { Archive } from 'api/archives.type'
import CloseIcon from '@material-ui/icons/Close'
import DeleteOutlinedIcon from '@material-ui/icons/DeleteOutlined'
import FilterListIcon from '@material-ui/icons/FilterList'
import Loading from 'components-mui/Loading'
import NotFound from 'components-mui/NotFound'
import ObjectListItem from 'components/ObjectListItem'
import PlaylistAddCheckIcon from '@material-ui/icons/PlaylistAddCheck'
import Space from 'components-mui/Space'
import T from 'components/T'
import Tab from '@material-ui/core/Tab'
import TabContext from '@material-ui/lab/TabContext'
import TabList from '@material-ui/lab/TabList'
import _groupBy from 'lodash.groupby'
import api from 'api'
import { styled } from '@material-ui/styles'
import { transByKind } from 'lib/byKind'
import { useHistory } from 'react-router-dom'
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

type PanelType = 'workflow' | 'schedule' | 'experiment'

export default function Archives() {
  const history = useHistory()
  const intl = useIntl()
  const query = useQuery()
  let kind = query.get('kind') || 'experiment'

  const dispatch = useStoreDispatch()

  const [panel, setPanel] = useState<PanelType>(kind as PanelType)
  const [loading, setLoading] = useState(true)
  const [archives, setArchives] = useState<Archive[]>([])
  const [batch, setBatch] = useState<Record<uuid, boolean>>({})
  const batchLength = Object.keys(batch).length
  const isBatchEmpty = batchLength === 0

  const fetchArchives = useCallback(() => {
    let request
    switch (kind) {
      case 'workflow':
        request = api.workflows.archives
        break
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
        switch (kind) {
          case 'workflow':
            actionFunc = api.workflows.delArchive
            break
          case 'schedule':
            actionFunc = api.schedules.delArchive
            break
          case 'experiment':
          default:
            actionFunc = api.archives.del
            break
        }
        arg = uuid

        break
      case 'deleteMulti':
        action = 'delete'
        switch (kind) {
          case 'workflow':
            actionFunc = api.workflows.delArchives
            break
          case 'schedule':
            actionFunc = api.schedules.delArchives
            break
          case 'experiment':
          default:
            actionFunc = api.archives.delMulti
            break
        }
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
              message: T(`confirm.success.${action}`, intl),
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
      title: T('archives.deleteMulti', intl),
      description: T('archives.deleteDesc', intl),
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
    history.push(`/archives?kind=${newValue}`)
    setPanel(newValue)
  }

  return (
    <TabContext value={panel}>
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <TabList onChange={onTabChange}>
          <Tab label={T('workflows.title')} value="workflow" />
          <Tab label={T('schedules.title')} value="schedule" />
          <Tab label={T('experiments.title')} value="experiment" />
        </TabList>
      </Box>

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
    </TabContext>
  )
}
