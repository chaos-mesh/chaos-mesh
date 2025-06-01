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
import PaperTop from '@/mui-extends/PaperTop'
import { usePostExperiments } from '@/openapi'
import { useComponentActions } from '@/zustand/component'
import { useExperimentActions, useExperimentStore } from '@/zustand/experiment'
import { useSettingStore } from '@/zustand/setting'
import DoneAllIcon from '@mui/icons-material/DoneAll'
import { Box, Typography } from '@mui/material'
import { useIntl } from 'react-intl'
import { useNavigate } from 'react-router'

import { Submit } from '@/components/FormField'
import { type ExperimentKind } from '@/components/NewExperiment/types'
import i18n from '@/components/T'

import { parseSubmit } from '@/lib/formikhelpers'

interface Step3Props {
  onSubmit?: (parsedValues: any) => void
  inSchedule?: boolean
}

const Step3: ReactFCWithChildren<Step3Props> = ({ onSubmit, inSchedule }) => {
  const navigate = useNavigate()
  const intl = useIntl()

  const { setAlert } = useComponentActions()

  const { step1, step2, kindAction, env, basic, spec } = useExperimentStore()
  const { debugMode, useNewPhysicalMachine } = useSettingStore()
  const reset = useExperimentActions().reset

  const { mutateAsync } = usePostExperiments()

  const submitExperiment = () => {
    const parsedValues = parseSubmit(
      env,
      kindAction[0] as ExperimentKind,
      {
        ...basic,
        spec: {
          ...basic.spec,
          ...spec,
        },
      },
      { inSchedule, useNewPhysicalMachine },
    )

    if (import.meta.env.DEV || debugMode) {
      console.debug('Here is the experiment you are going to submit:', JSON.stringify(parsedValues, null, 2))
    }

    if (!debugMode) {
      if (onSubmit) {
        onSubmit(parsedValues)
      } else {
        mutateAsync({
          data: parsedValues,
        })
          .then(() => {
            setAlert({
              type: 'success',
              message: i18n('confirm.success.create', intl),
            })

            reset()

            navigate('/experiments')
          })
          .catch((error) => {
            if (import.meta.env.DEV) {
              console.error('Error submitting experiment:', error.response)
            }

            setAlert({
              type: 'error',
              message: error.response.data.message,
            })
          })
      }
    }
  }

  return (
    <>
      {step1 && step2 && (
        <Paper>
          <PaperTop title={i18n('common.submit')} boxProps={{ mb: 6 }} />
          <Box textAlign="center">
            <DoneAllIcon fontSize="large" />
            <Typography>{i18n('newE.complete')}</Typography>
          </Box>
          <Submit onClick={submitExperiment} />
        </Paper>
      )}
    </>
  )
}

export default Step3
