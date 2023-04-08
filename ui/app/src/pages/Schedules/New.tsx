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
import { Grid } from '@mui/material'
import { usePostSchedules } from 'openapi'
import { useIntl } from 'react-intl'
import { useNavigate } from 'react-router-dom'

import { useStoreDispatch } from 'store'

import { resetNewExperiment } from 'slices/experiments'
import { setAlert } from 'slices/globalStatus'

import NewExperiment from 'components/NewExperimentNext'
import i18n from 'components/T'

const New = () => {
  const navigate = useNavigate()
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const { mutateAsync } = usePostSchedules()

  const onSubmit = (parsedValues: any) => {
    mutateAsync({ data: parsedValues })
      .then(() => {
        dispatch(
          setAlert({
            type: 'success',
            message: i18n('confirm.success.create', intl),
          })
        )

        dispatch(resetNewExperiment())

        navigate('/schedules')
      })
      .catch(console.error)
  }

  return (
    <Grid container>
      <Grid item xs={12} lg={8}>
        <NewExperiment inSchedule={true} onSubmit={onSubmit} />
      </Grid>
    </Grid>
  )
}

export default New
