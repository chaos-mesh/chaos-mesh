import { Box, Divider, FormControlLabel, Radio, RadioGroup, Typography } from '@material-ui/core'
import { PreDefinedValue, getDB } from 'lib/idb'
import { useEffect, useState } from 'react'

import { Archive } from 'api/archives.type'
import { Experiment } from 'api/experiments.type'
import Paper from 'components-mui/Paper'
import RadioLabel from './RadioLabel'
import { Schedule } from 'api/schedules.type'
import SkeletonN from 'components-mui/SkeletonN'
import Space from 'components-mui/Space'
import T from 'components/T'
import api from 'api'
import { setAlert } from 'slices/globalStatus'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

interface LoadFromProps {
  callback?: (data: any) => void
  inSchedule?: boolean
  inWorkflow?: boolean
}

const LoadFrom: React.FC<LoadFromProps> = ({ callback, inSchedule, inWorkflow }) => {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(true)
  const [data, setData] = useState<{
    experiments: Experiment[]
    archives: Archive[]
    schedules: Schedule[]
  }>({
    experiments: [],
    archives: [],
    schedules: [],
  })
  const [predefined, setPredefined] = useState<PreDefinedValue[]>([])
  const [radio, setRadio] = useState('')

  useEffect(() => {
    const fetchExperiments = api.experiments.experiments()
    const fetchArchives = api.archives.archives()
    const promises: Promise<any>[] = [fetchExperiments, fetchArchives]

    if (inSchedule) {
      promises.push(api.schedules.schedules())
    }

    const fetchAll = async () => {
      const data = await Promise.all(promises)

      setData({
        experiments: data[0].data,
        archives: data[1].data,
        schedules: data[2] ? data[2].data : [],
      })

      let _predefined = await (await getDB()).getAll('predefined' as never) // never?
      if (inWorkflow) {
        _predefined = _predefined.filter((d) => d.kind !== 'Schedule')
      }
      setPredefined(_predefined)

      setLoading(false)
    }

    fetchAll()
  }, [inSchedule, inWorkflow])

  const onRadioChange = (e: any) => {
    const [type, uuid] = e.target.value.split('+')

    if (type === 'p') {
      const experiment = predefined?.filter((p) => p.name === uuid)[0].yaml

      callback && callback(experiment)

      dispatch(
        setAlert({
          type: 'success',
          message: T('confirm.success.load', intl),
        })
      )

      return
    }

    let apiRequest
    switch (type) {
      case 's':
        apiRequest = api.schedules
        break
      case 'e':
        apiRequest = api.experiments
        break
      case 'a':
        apiRequest = api.archives
        break
    }

    setRadio(e.target.value)

    if (apiRequest) {
      apiRequest
        .single(uuid)
        .then(({ data }) => {
          callback && callback(data.kube_object)

          dispatch(
            setAlert({
              type: 'success',
              message: T('confirm.success.load', intl),
            })
          )
        })
        .catch(console.error)
    }
  }

  return (
    <Paper>
      <RadioGroup value={radio} onChange={onRadioChange}>
        <Space>
          {inSchedule && (
            <>
              <Typography>{T('schedules.title')}</Typography>

              {loading ? (
                <SkeletonN n={3} />
              ) : data.schedules.length > 0 ? (
                <Box display="flex" flexWrap="wrap">
                  {data.schedules.map((d) => (
                    <FormControlLabel
                      key={d.uid}
                      value={`s+${d.uid}`}
                      control={<Radio color="primary" />}
                      label={RadioLabel(d.name, d.uid)}
                    />
                  ))}
                </Box>
              ) : (
                <Typography variant="body2" color="textSecondary">
                  {T('experiments.notFound')}
                </Typography>
              )}

              <Divider />
            </>
          )}

          <Typography>{T('experiments.title')}</Typography>

          {loading ? (
            <SkeletonN n={3} />
          ) : data.experiments.length > 0 ? (
            <Box display="flex" flexWrap="wrap">
              {data.experiments.map((d) => (
                <FormControlLabel
                  key={d.uid}
                  value={`e+${d.uid}`}
                  control={<Radio color="primary" />}
                  label={RadioLabel(d.name, d.uid)}
                />
              ))}
            </Box>
          ) : (
            <Typography variant="body2" color="textSecondary">
              {T('experiments.notFound')}
            </Typography>
          )}

          <Divider />

          <Typography>{T('archives.title')}</Typography>

          {loading ? (
            <SkeletonN n={3} />
          ) : data.archives.length > 0 ? (
            <Box display="flex" flexWrap="wrap">
              {data.archives.map((d) => (
                <FormControlLabel
                  key={d.uid}
                  value={`a+${d.uid}`}
                  control={<Radio color="primary" />}
                  label={RadioLabel(d.name, d.uid)}
                />
              ))}
            </Box>
          ) : (
            <Typography variant="body2" color="textSecondary">
              {T('archives.notFound')}
            </Typography>
          )}

          <Divider />

          <Typography>{T('dashboard.predefined')}</Typography>

          {loading ? (
            <SkeletonN n={3} />
          ) : predefined.length > 0 ? (
            <Box display="flex" flexWrap="wrap">
              {predefined.map((d) => (
                <FormControlLabel
                  key={d.name}
                  value={`p+${d.name}`}
                  control={<Radio color="primary" />}
                  label={RadioLabel(d.name)}
                />
              ))}
            </Box>
          ) : (
            <Typography variant="body2" color="textSecondary">
              {T('dashboard.noPredefinedFound')}
            </Typography>
          )}
        </Space>
      </RadioGroup>
    </Paper>
  )
}

export default LoadFrom
