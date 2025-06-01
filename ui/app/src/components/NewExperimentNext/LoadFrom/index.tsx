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
import Paper from '@/mui-extends/Paper'
import SkeletonN from '@/mui-extends/SkeletonN'
import Space from '@/mui-extends/Space'
import {
  useGetArchives,
  useGetArchivesSchedules,
  useGetArchivesSchedulesUid,
  useGetArchivesUid,
  useGetExperiments,
  useGetExperimentsUid,
  useGetSchedules,
  useGetSchedulesUid,
} from '@/openapi'
import { TypesArchiveDetail, TypesExperimentDetail, TypesScheduleDetail } from '@/openapi/index.schemas'
import { useStoreDispatch } from '@/store'
import { useComponentActions } from '@/zustand/component'
import { Box, Divider, FormControlLabel, Radio, RadioGroup, Typography } from '@mui/material'
import { useEffect, useState } from 'react'
import { useIntl } from 'react-intl'

import i18n from '@/components/T'

import RadioLabel from './RadioLabel'

interface LoadFromProps {
  callback?: (data: any) => void
  inSchedule?: boolean
  inWorkflow?: boolean
}

const LoadFrom: ReactFCWithChildren<LoadFromProps> = ({ callback, inSchedule, inWorkflow }) => {
  const intl = useIntl()

  const { setAlert } = useComponentActions()
  const dispatch = useStoreDispatch()

  const [metaInfo, setMetaInfo] = useState<{
    id: string
    type: string
  }>({
    id: '',
    type: '',
  })
  const [radio, setRadio] = useState('')

  const { data: experiments, isLoading: loading1 } = useGetExperiments()
  const { data: schedules, isLoading: loading2 } = useGetSchedules(undefined, { query: { enabled: inSchedule } })
  const { data: archives, isLoading: loading3 } = (inSchedule ? useGetArchivesSchedules : useGetArchives)()
  const loading = loading1 || loading2 || loading3

  const { data: scheduleData } = useGetSchedulesUid(metaInfo.id, {
    query: {
      enabled: metaInfo.type === 'schedule',
    },
  })
  const { data: experimentData } = useGetExperimentsUid(metaInfo.id, {
    query: {
      enabled: metaInfo.type === 'experiment',
    },
  })
  const { data: scheduleArchiveData } = useGetArchivesSchedulesUid(metaInfo.id, {
    query: {
      enabled: metaInfo.type === 'scheduleArchive',
    },
  })
  const { data: archiveData } = useGetArchivesUid(metaInfo.id, {
    query: {
      enabled: metaInfo.type === 'archive',
    },
  })

  useEffect(() => {
    function afterLoad(data: TypesScheduleDetail | TypesExperimentDetail | TypesArchiveDetail) {
      if (callback) {
        callback(data.kube_object)
      }

      dispatch(
        setAlert({
          type: 'success',
          message: i18n('confirm.success.load', intl),
        }),
      )
    }

    if (scheduleData) {
      afterLoad(scheduleData)
    }

    if (experimentData) {
      afterLoad(experimentData)
    }

    if (scheduleArchiveData) {
      afterLoad(scheduleArchiveData)
    }

    if (archiveData) {
      afterLoad(archiveData)
    }

    setMetaInfo({
      id: '',
      type: '',
    })
  }, [scheduleData, experimentData, scheduleArchiveData, archiveData])

  const onRadioChange = (e: any) => {
    const [type, uuid] = e.target.value.split('+')

    switch (type) {
      case 's':
        setMetaInfo({
          id: uuid,
          type: 'schedule',
        })

        break
      case 'e':
        setMetaInfo({
          id: uuid,
          type: 'experiment',
        })

        break
      case 'a':
        if (inSchedule) {
          setMetaInfo({
            id: uuid,
            type: 'scheduleArchive',
          })
        } else {
          setMetaInfo({
            id: uuid,
            type: 'archive',
          })
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
        </Space>
      </RadioGroup>
    </Paper>
  )
}

export default LoadFrom
