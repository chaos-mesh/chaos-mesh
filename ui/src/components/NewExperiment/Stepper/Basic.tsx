import { SelectField, TextField } from 'components/FormField'

import { Experiment } from 'components/NewExperiment/types'
import { MenuItem } from '@material-ui/core'
import React from 'react'
import { useFormikContext } from 'formik'

interface BasicStepProps {
  namespaces: string[]
}

const BasicStep: React.FC<BasicStepProps> = ({ namespaces }) => {
  const { errors, handleChange, setFieldValue } = useFormikContext<Experiment>()

  const handleBasicNamespaceChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    handleChange(e)

    setFieldValue('scope.namespace_selectors', [e.target.value])
  }

  return (
    <>
      <TextField
        id="name"
        name="name"
        label="Name"
        helperText="Please input an experiment name"
        error={errors.name ? true : false}
      />

      <SelectField
        id="namespace"
        name="namespace"
        label="Namespace"
        helperText="Select the experiment's namespace"
        onChange={handleBasicNamespaceChange}
      >
        {namespaces.map((n) => (
          <MenuItem key={n} value={n}>
            {n}
          </MenuItem>
        ))}
      </SelectField>
    </>
  )
}

export default BasicStep
