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
import { Typography } from '@mui/material'
import { Form, Formik, getIn } from 'formik'
import { useEffect, useState } from 'react'

import Space from '@ui/mui-extends/esm/Space'

import { useStoreSelector } from 'store'

import { LabelField, Submit, TextField } from 'components/FormField'
import MoreOptions from 'components/MoreOptions'

import typesData from '../data/types'

const validate = (values: any) => {
  let errors = {}

  const { cpu, memory } = values.stressors

  if (cpu.workers <= 0 && memory.workers <= 0) {
    const message = 'The CPU or Memory workers must have at least one greater than 0'

    errors = {
      stressors: {
        cpu: {
          workers: message,
        },
        memory: {
          workers: message,
        },
      },
    }
  }

  return errors
}

interface StressProps {
  onSubmit: (values: Record<string, any>) => void
}

const Stress: React.FC<StressProps> = ({ onSubmit }) => {
  const { spec } = useStoreSelector((state) => state.experiments)

  const initialValues = typesData.StressChaos.spec!

  const [init, setInit] = useState(initialValues)

  useEffect(() => {
    setInit({
      ...initialValues,
      ...spec,
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [spec])

  return (
    <Formik enableReinitialize initialValues={init} onSubmit={onSubmit} validate={validate}>
      {({ errors }) => (
        <Form>
          <Space>
            <Typography>CPU</Typography>
            <TextField
              type="number"
              name="stressors.cpu.workers"
              label="Workers"
              helperText={
                getIn(errors, 'stressors.cpu.workers') ? getIn(errors, 'stressors.cpu.workers') : 'CPU workers'
              }
              error={getIn(errors, 'stressors.cpu.workers') ? true : false}
              inputProps={{ min: 0 }}
            />
            <TextField type="number" name="stressors.cpu.load" label="Load" helperText="CPU load" />
            <LabelField
              name="stressors.cpu.options"
              label="Options of CPU stressors"
              helperText="Type and end with Enter to generate the stress-ng options"
            />

            <Typography>Memory</Typography>
            <TextField
              type="number"
              name="stressors.memory.workers"
              label="Workers"
              helperText={
                getIn(errors, 'stressors.memory.workers') ? getIn(errors, 'stressors.memory.workers') : 'Memory workers'
              }
              error={getIn(errors, 'stressors.memory.workers') ? true : false}
              inputProps={{ min: 0 }}
            />
            <TextField
              name="stressors.memory.size"
              label="Size"
              helperText="Memory size specifies the memory size to be occupied or a percentage of the total memory size"
            />
            <LabelField
              name="stressors.memory.options"
              label="Options of Memory stressors"
              helperText="Type and end with Enter to generate the memStress options"
            />
          </Space>

          <MoreOptions>
            <TextField
              name="stressngStressors"
              label="Options of stress-ng"
              helperText="The options of stress-ng, treated as a string"
            />
            <LabelField
              name="containerNames"
              label="Container Name"
              helperText="Optional. Type and end with Enter to generate the container names. If it's empty, the first container will be injected"
            />
          </MoreOptions>

          <Submit />
        </Form>
      )}
    </Formik>
  )
}

export default Stress
