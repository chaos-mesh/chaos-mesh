import { Box, Button } from '@material-ui/core'
import { Confirm, setAlert, setConfirm } from 'slices/globalStatus'
import { FixedSizeList as RWList, ListChildComponentProps as RWListChildComponentProps } from 'react-window'
import { useEffect, useState } from 'react'

import AddIcon from '@material-ui/icons/Add'
import Loading from 'components-mui/Loading'
import NotFound from 'components-mui/NotFound'
import ObjectListItem from 'components/ObjectListItem'
import { Schedule } from 'api/schedules.type'
import Space from 'components-mui/Space'
import T from 'components/T'
import { Typography } from '@material-ui/core'
import _groupBy from 'lodash.groupby'
import api from 'api'
import { transByKind } from 'lib/byKind'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

const Schedules = () => {
  const history = useHistory()
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(true)
  const [schedules, setSchedules] = useState<Schedule[]>([])

  const fetchSchedules = () => {
    api.schedules
      .schedules()
      .then(({ data }) => setSchedules(data))
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(fetchSchedules, [])

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

          setTimeout(fetchSchedules, 300)
        })
        .catch(console.error)
    }
  }

  const Row = ({ data, index, style }: RWListChildComponentProps) => (
    <Box display="flex" alignItems="center" mb={3} style={style}>
      <Box flex={1}>
        <ObjectListItem type="schedule" data={data[index]} onSelect={onSelect} />
      </Box>
    </Box>
  )

  return (
    <>
      <Space mb={6}>
        <Button variant="outlined" startIcon={<AddIcon />} onClick={() => history.push('/schedules/new')}>
          {T('newS.title')}
        </Button>
      </Space>

      {schedules.length > 0 &&
        Object.entries(_groupBy(schedules, 'type')).map(([type, schedulesByType]) => (
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
