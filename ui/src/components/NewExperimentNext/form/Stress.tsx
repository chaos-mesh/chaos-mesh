import { Form, Formik, getIn } from 'formik'
import { LabelField, Submit, TextField } from 'components/FormField'
import { useEffect, useState } from 'react'

import AdvancedOptions from 'components/AdvancedOptions'
import Space from 'components-mui/Space'
import { Typography } from '@material-ui/core'
import targetData from '../data/target'
import { useStoreSelector } from 'store'

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
  const { target } = useStoreSelector((state) => state.experiments)

  const initialValues = targetData.StressChaos.spec!

  const [init, setInit] = useState(initialValues)

  useEffect(() => {
    setInit({
      ...initialValues,
      ...target['stress_chaos'],
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [target])

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
              helperText="Type string and end with a space to generate the stress-ng options"
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
            <TextField name="stressors.memory.size" label="Size" helperText="Memory size" />
            <LabelField
              name="stressors.memory.options"
              label="Options of Memory stressors"
              helperText="Type string and end with a space to generate the stress-ng options"
            />
          </Space>

          <AdvancedOptions>
            <TextField
              name="stressng_stressors"
              label="Options of stress-ng"
              helperText="The options of stress-ng, treated as a string"
            />
            <TextField
              name="container_name"
              label="Container Name"
              helperText="Optional. Fill the container name you want to inject stress in"
            />
          </AdvancedOptions>

          <Submit />
        </Form>
      )}
    </Formik>
  )
}

export default Stress
