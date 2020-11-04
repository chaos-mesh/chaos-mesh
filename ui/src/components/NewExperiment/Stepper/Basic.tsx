import { LabelField, SelectField, TextField } from 'components/FormField'

import AdvancedOptions from 'components/AdvancedOptions'
import { Experiment } from 'components/NewExperiment/types'
import { MenuItem } from '@material-ui/core'
import React from 'react'
import T from 'components/T'
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
        label={T('newE.basic.name')}
        helperText={T('newE.basic.nameHelper')}
        error={errors.name && touched.name ? true : false}
      />

      <SelectField
        id="namespace"
        name="namespace"
        label={T('newE.basic.namespace')}
        helperText={T('newE.basic.namespaceHelper')}
        onChange={handleBasicNamespaceChange}
      >
        {namespaces.map((n) => (
          <MenuItem key={n} value={n}>
            {n}
          </MenuItem>
        ))}
      </SelectField>

      <AdvancedOptions isOpen>
        <LabelField id="labels" name="labels" label={T('k8s.labels')} isKV />
        <LabelField id="annotations" name="annotations" label={T('k8s.annotations')} isKV />
      </AdvancedOptions>
    </>
  )
}

export default BasicStep
