import { Box, Button, MenuItem } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { LabelField, SelectField, TextField } from 'components/FormField'

import PublishIcon from '@material-ui/icons/Publish'
import React from 'react'
import { Spec } from '../data/target'
import T from 'components/T'

interface TargetGeneratedProps {
  data: Spec
  onSubmit: (values: Record<string, any>) => void
}

const TargetGenerated: React.FC<TargetGeneratedProps> = ({ data, onSubmit }) => {
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
            return null
          default:
            return null
        }
      })
      .filter((d) => d)

    return <>{rendered.map((d) => d)}</>
  }

  return (
    <Formik initialValues={initialValues} onSubmit={onSubmit}>
      <Form>
        {parseDataToFormFields()}
        <Box mt={6} textAlign="right">
          <Button type="submit" variant="contained" color="primary" startIcon={<PublishIcon />}>
            {T('common.submit')}
          </Button>
        </Box>
      </Form>
    </Formik>
  )
}

export default TargetGenerated
