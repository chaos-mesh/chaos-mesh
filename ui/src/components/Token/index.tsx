import { Box, Button } from '@material-ui/core'
import { Form, Formik, FormikHelpers } from 'formik'

import LS from 'lib/localStorage'
import React from 'react'
import T from 'components/T'
import { TextField } from 'components/FormField'
import api from 'api'
import { setTokens } from 'slices/globalStatus'
import { useStoreDispatch } from 'store'

function validateName(value: string) {
  let error

  if (value === '') {
    error = (T('settings.addToken.nameValidation') as unknown) as string
  }

  return error
}

function validateToken(value: string) {
  let error

  if (value === '') {
    error = (T('settings.addToken.tokenValidation') as unknown) as string
  }

  return error
}

export interface TokenFormValues {
  name: string
  token: string
}

interface TokenProps {
  onSubmitCallback?: (values: TokenFormValues) => void
}

const Token: React.FC<TokenProps> = ({ onSubmitCallback }) => {
  const dispatch = useStoreDispatch()

  const saveToken = (values: TokenFormValues) => {
    let tokens = []
    const previous = LS.get('token')

    if (previous) {
      tokens = JSON.parse(previous)
    }

    tokens.push(values)

    dispatch(setTokens(tokens))
  }

  const submitToken = (values: TokenFormValues, { resetForm }: FormikHelpers<TokenFormValues>) => {
    api.auth.token(values.token)

    saveToken(values)

    typeof onSubmitCallback === 'function' && onSubmitCallback(values)

    resetForm()
  }

  return (
    <Formik initialValues={{ name: '', token: '' }} onSubmit={submitToken}>
      {({ errors, touched }) => (
        <Form>
          <TextField
            name="name"
            label={T('settings.addToken.name')}
            validate={validateName}
            helperText={errors.name && touched.name ? errors.name : T('settings.addToken.nameHelper')}
            error={errors.name && touched.name ? true : false}
          />
          <TextField
            name="token"
            label={T('settings.addToken.token')}
            multiline
            rows={12}
            validate={validateToken}
            helperText={errors.token && touched.token ? errors.token : T('settings.addToken.tokenHelper')}
            error={errors.token && touched.token ? true : false}
          />
          <Box textAlign="right">
            <Button type="submit" variant="contained" color="primary">
              {T('common.submit')}
            </Button>
          </Box>
        </Form>
      )}
    </Formik>
  )
}

export default Token
