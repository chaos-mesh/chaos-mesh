import { SelectField, TextField, LabelField } from 'components/FormField'

import AdvancedOptions from 'components/AdvancedOptions'
import { Experiment } from 'components/NewExperiment/types'
import { MenuItem } from '@material-ui/core'
import React from 'react'
import { useFormikContext } from 'formik'

interface BasicStepProps {
  namespaces: string[]
}

const BasicStep: React.FC<BasicStepProps> = ({ namespaces }) => {
  const { errors, touched, handleChange, setFieldValue } = useFormikContext<Experiment>()

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
        helperText="The experiment name"
        autoFocus
        error={errors.name && touched.name ? true : false}
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

      <AdvancedOptions isOpen>
        <LabelField id="labels" name="labels" label="Labels" isKV />
        <LabelField id="annotations" name="annotations" label="Annotations" isKV />
      </AdvancedOptions>
    </>
  )
}

export default BasicStep
