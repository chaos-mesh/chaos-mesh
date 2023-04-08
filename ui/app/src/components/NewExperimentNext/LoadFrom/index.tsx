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
import { Box, Divider, FormControlLabel, Radio, RadioGroup, Typography } from '@mui/material'
import {
  useGetArchives,
  useGetArchivesSchedules,
  useGetArchivesSchedulesUid,
  useGetArchivesUid,
  useGetExperiments,
  useGetExperimentsUid,
  useGetSchedules,
  useGetSchedulesUid,
} from 'openapi'
import { TypesArchiveDetail, TypesExperimentDetail, TypesScheduleDetail } from 'openapi/index.schemas'
import { useEffect, useState } from 'react'
import { useIntl } from 'react-intl'

import Paper from '@ui/mui-extends/esm/Paper'
import SkeletonN from '@ui/mui-extends/esm/SkeletonN'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch } from 'store'

import { setAlert } from 'slices/globalStatus'

import i18n from 'components/T'

import { PreDefinedValue, getDB } from 'lib/idb'

import RadioLabel from './RadioLabel'

interface LoadFromProps {
  callback?: (data: any) => void
  inSchedule?: boolean
  inWorkflow?: boolean
}

const LoadFrom: React.FC<LoadFromProps> = ({ callback, inSchedule, inWorkflow }) => {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [scheduleUUID, setScheduleUUID] = useState('')
  const [experimentUUID, setExperimentUUID] = useState('')
  const [archiveUUID, setArchiveUUID] = useState('')
  const [scheduleArchiveUUID, setScheduleArchiveUUID] = useState('')
  const [predefined, setPredefined] = useState<PreDefinedValue[]>([])
  const [radio, setRadio] = useState('')

  const { data: experiments, isLoading: loading1 } = useGetExperiments()
  const { data: schedules, isLoading: loading2 } = useGetSchedules(undefined, { query: { enabled: inSchedule } })
  const { data: archives, isLoading: loading3 } = (inSchedule ? useGetArchivesSchedules : useGetArchives)()
  const loading = loading1 || loading2 || loading3
  function afterLoad(data: TypesScheduleDetail | TypesExperimentDetail | TypesArchiveDetail) {
    callback && callback(data.kube_object)

    dispatch(
      setAlert({
        type: 'success',
        message: i18n('confirm.success.load', intl),
      })
    )
  }
  useGetSchedulesUid(scheduleUUID, {
    query: {
      enabled: !!scheduleUUID,
      onSuccess(data) {
        afterLoad(data)

        setScheduleUUID('')
      },
    },
  })
  useGetExperimentsUid(experimentUUID, {
    query: {
      enabled: !!experimentUUID,
      onSuccess(data) {
        afterLoad(data)

        setExperimentUUID('')
      },
    },
  })
  useGetArchivesSchedulesUid(scheduleArchiveUUID, {
    query: {
      enabled: !!scheduleArchiveUUID,
      onSuccess(data) {
        afterLoad(data)

        setScheduleArchiveUUID('')
      },
    },
  })
  useGetArchivesUid(archiveUUID, {
    query: {
      enabled: !!archiveUUID,
      onSuccess(data) {
        afterLoad(data)

        setArchiveUUID('')
      },
    },
  })

  useEffect(() => {
    const fetchPredefined = async () => {
      let _predefined = await (await getDB()).getAll('predefined')

      if (!inSchedule) {
        _predefined = _predefined.filter((d) => d.kind !== 'Schedule')
      }

      setPredefined(_predefined)
    }

    fetchPredefined()
  }, [inSchedule, inWorkflow])

  const onRadioChange = (e: any) => {
    const [type, uuid] = e.target.value.split('+')

    if (type === 'p') {
      const experiment = predefined?.filter((p) => p.name === uuid)[0].yaml

      callback && callback(experiment)

      dispatch(
        setAlert({
          type: 'success',
          message: i18n('confirm.success.load', intl),
        })
      )

      return
    }

    switch (type) {
      case 's':
        setScheduleUUID(uuid)

        break
      case 'e':
        setExperimentUUID(uuid)

        break
      case 'a':
        if (inSchedule) {
          setScheduleArchiveUUID(uuid)
        } else {
          setArchiveUUID(uuid)
        }

        break
    }

    setRadio(e.target.value)
  }

  return (
    <Paper>
      <RadioGroup value={radio} onChange={onRadioChange}>
        <Space>
          {inSchedule && (
            <>
              <Typography>{i18n('schedules.title')}</Typography>

              {loading ? (
                <SkeletonN n={3} />
              ) : schedules && schedules.length > 0 ? (
                <Box display="flex" flexWrap="wrap">
                  {schedules.map((d) => (
                    <FormControlLabel
                      key={d.uid}
                      value={`s+${d.uid}`}
                      control={<Radio color="primary" />}
                      label={RadioLabel(d.name!, d.uid)}
                    />
                  ))}
                </Box>
              ) : (
                <Typography variant="body2" color="textSecondary">
                  {i18n('schedules.notFound')}
                </Typography>
              )}

              <Divider />
            </>
          )}

          {!inSchedule && (
            <>
              <Typography>{i18n('experiments.title')}</Typography>

              {loading ? (
                <SkeletonN n={3} />
              ) : experiments && experiments.length > 0 ? (
                <Box display="flex" flexWrap="wrap">
                  {experiments.map((d) => (
                    <FormControlLabel
                      key={d.uid}
                      value={`e+${d.uid}`}
                      control={<Radio color="primary" />}
                      label={RadioLabel(d.name!, d.uid)}
                    />
                  ))}
                </Box>
              ) : (
                <Typography variant="body2" color="textSecondary">
                  {i18n('experiments.notFound')}
                </Typography>
              )}

              <Divider />
            </>
          )}

          <Typography>{i18n('archives.title')}</Typography>

          {loading ? (
            <SkeletonN n={3} />
          ) : archives && archives.length > 0 ? (
            <Box display="flex" flexWrap="wrap">
              {archives.map((d) => (
                <FormControlLabel
                  key={d.uid}
                  value={`a+${d.uid}`}
                  control={<Radio color="primary" />}
                  label={RadioLabel(d.name!, d.uid)}
                />
              ))}
            </Box>
          ) : (
            <Typography variant="body2" color="textSecondary">
              {i18n('archives.notFound')}
            </Typography>
          )}

          <Divider />

          <Typography>{i18n('dashboard.predefined')}</Typography>

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
              {i18n('dashboard.noPredefinedFound')}
            </Typography>
          )}
        </Space>
      </RadioGroup>
    </Paper>
  )
}

export default LoadFrom
