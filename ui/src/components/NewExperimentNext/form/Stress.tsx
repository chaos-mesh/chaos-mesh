import { Box, Button } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { LabelField, TextField } from 'components/FormField'
import React, { useEffect, useState } from 'react'

import AdvancedOptions from 'components/AdvancedOptions'
import PublishIcon from '@material-ui/icons/Publish'
import { RootState } from 'store'
import T from 'components/T'
import { Typography } from '@material-ui/core'
import targetData from '../data/target'
import { useSelector } from 'react-redux'

interface StressProps {
  onSubmit: (values: Record<string, any>) => void
}

const Stress: React.FC<StressProps> = ({ onSubmit }) => {
  const { target } = useSelector((state: RootState) => state.experiments)

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
    <Formik enableReinitialize initialValues={init} onSubmit={onSubmit}>
      <Form>
        <Typography gutterBottom>CPU</Typography>
        <TextField type="number" name="stressors.cpu.workers" label="Workers" helperText="CPU workers" />
        <TextField type="number" name="stressors.cpu.load" label="Load" helperText="CPU load" />
        <LabelField
          name="stressors.cpu.options"
          label="Options of CPU stressors"
          helperText="Type string and end with a space to generate the stress-ng options"
        />

        <Typography gutterBottom>Memory</Typography>
        <TextField type="number" name="stressors.memory.workers" label="Workers" helperText="Memory workers" />
        <LabelField
          name="stressors.memory.options"
          label="Options of Memory stressors"
          helperText="Type string and end with a space to generate the stress-ng options"
        />

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

        <Box mt={6} textAlign="right">
          <Button type="submit" variant="contained" color="primary" startIcon={<PublishIcon />}>
            {T('common.submit')}
          </Button>
        </Box>
      </Form>
    </Formik>
  )
}

export default Stress
