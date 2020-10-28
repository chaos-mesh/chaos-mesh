import { AutocompleteMultipleField, LabelField, SelectField, TextField } from 'components/FormField'
import { Box, Button, MenuItem } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { Kind, Spec } from '../data/target'

import AdvancedOptions from 'components/AdvancedOptions'
import PublishIcon from '@material-ui/icons/Publish'
import React from 'react'
import { RootState } from 'store'
import Scope from './Scope'
import T from 'components/T'
import basicData from '../data/basic'
import { useSelector } from 'react-redux'

interface TargetGeneratedProps {
  kind?: Kind | ''
  data: Spec
  onSubmit: (values: Record<string, any>) => void
}

const TargetGenerated: React.FC<TargetGeneratedProps> = ({ kind, data, onSubmit }) => {
  const initialValues = Object.entries(data).reduce((acc, [k, v]) => {
    if (v instanceof Object && v.field) {
      acc[k] = v.value
    } else {
      acc[k] = v
    }

    return acc
  }, {} as Record<string, any>)

  const parseDataToFormFields = () => {
    const rendered = Object.entries(data)
      .filter(([_, v]) => v instanceof Object && v.field)
      .map(([k, v]) => {
        switch (v.field) {
          case 'text':
            return <TextField key={k} id={k} name={k} label={v.label} helperText={v.helperText} {...v.inputProps} />
          case 'number':
            return (
              <TextField
                key={k}
                type="number"
                id={k}
                name={k}
                label={v.label}
                helperText={v.helperText}
                {...v.inputProps}
              />
            )
          case 'select':
            return (
              <SelectField key={k} id={k} name={k} label={v.label} helperText={v.helperText}>
                {v.items!.map((option: string) => (
                  <MenuItem key={option} value={option}>
                    {option}
                  </MenuItem>
                ))}
              </SelectField>
            )
          case 'label':
            return <LabelField key={k} id={k} name={k} label={v.label} helperText={v.helperText} isKV={v.isKV} />
          case 'autocomplete':
            return (
              <AutocompleteMultipleField
                key={k}
                id={k}
                name={k}
                label={v.label}
                helperText={v.helperText}
                options={v.items!}
              />
            )
          default:
            return null
        }
      })
      .filter((d) => d)

    return <>{rendered.map((d) => d)}</>
  }

  const { namespaces } = useSelector((state: RootState) => state.experiments)

  return (
    <Formik initialValues={initialValues} onSubmit={onSubmit}>
      {({ values, setFieldValue }) => {
        const beforeTargetOpen = () => setFieldValue('target', basicData.scope)
        const afterTargetClose = () => setFieldValue('target', undefined)

        return (
          <Form>
            {parseDataToFormFields()}
            {kind === 'NetworkChaos' && (
              <AdvancedOptions
                title={T('newE.target.network.target.title')}
                beforeOpen={beforeTargetOpen}
                afterClose={afterTargetClose}
              >
                {values.target && (
                  <Scope
                    namespaces={namespaces}
                    scope="target"
                    podsPreviewTitle={T('newE.target.network.target.podsPreview')}
                    podsPreviewDesc={T('newE.target.network.target.podsPreviewHelper')}
                  />
                )}
              </AdvancedOptions>
            )}
            <Box mt={6} textAlign="right">
              <Button type="submit" variant="contained" color="primary" startIcon={<PublishIcon />}>
                {T('common.submit')}
              </Button>
            </Box>
          </Form>
        )
      }}
    </Formik>
  )
}

export default TargetGenerated
