import { SelectField, TextField } from 'components/FormField'

import { MenuItem } from '@material-ui/core'
import React from 'react'
import { StepperFormProps } from '../types'

interface BasicStepProps {
  formProps: StepperFormProps
  namespaces: string[]
}

const BasicStep: React.FC<BasicStepProps> = ({ formProps, namespaces }) => {
  const { values, handleChange } = formProps

  const handleBasicNamespaceChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    handleChange(e)

    formProps.setFieldValue('scope.namespace_selectors', [e.target.value])
  }

  return (
    <>
      <TextField
        id="name"
        label="Name"
        helperText="Please input an experiment name"
        value={values.name}
        onChange={handleChange}
      />

      <SelectField
        id="namespace"
        name="namespace"
        label="Namespace"
        helperText="Select the experiment's namespace"
        value={values.namespace}
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
