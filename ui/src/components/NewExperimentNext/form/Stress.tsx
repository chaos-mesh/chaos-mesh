import { Form, Formik } from 'formik'
import { LabelField, TextField } from 'components/FormField'

import AdvancedOptions from 'components/AdvancedOptions'
import React from 'react'
import { Typography } from '@material-ui/core'
import targetData from '../data/target'

interface StressProps {
  onSubmit: (values: Record<string, any>) => void
}

const Stress: React.FC<StressProps> = ({ onSubmit }) => {
  return (
    <Formik initialValues={targetData.StressChaos.spec!} onSubmit={onSubmit}>
      <Form>
        <Typography gutterBottom>CPU</Typography>
        <TextField
          type="number"
          id="stressors.cpu.workers"
          name="stressors.cpu.workers"
          label="Workers"
          helperText="CPU workers"
        />
        <TextField type="number" id="stressors.cpu.load" name="stressors.cpu.load" label="Load" helperText="CPU load" />
        <LabelField
          id="stressors.cpu.options"
          name="stressors.cpu.options"
          label="Options of CPU stressors"
          helperText="Type string and end with a space to generate the stress-ng options"
        />

        <Typography gutterBottom>Memory</Typography>
        <TextField
          type="number"
          id="stressors.memory.workers"
          name="stressors.memory.workers"
          label="Workers"
          helperText="Memory workers"
        />
        <LabelField
          id="stressors.memory.options"
          name="stressors.memory.options"
          label="Options of Memory stressors"
          helperText="Type string and end with a space to generate the stress-ng options"
        />

        <AdvancedOptions>
          <TextField
            id="stressng_stressors"
            name="stressng_stressors"
            label="Options of stress-ng"
            helperText="The options of stress-ng, treated as a string"
          />
          <TextField
            id="container_name"
            name="container_name"
            label="Container Name"
            helperText="Optional. Fill the container name you want to inject stress in"
          />
        </AdvancedOptions>
      </Form>
    </Formik>
  )
}

export default Stress
