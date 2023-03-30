/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { applyAPIAuthentication, resetAPIAuthentication } from 'api/interceptors'
import { Form, Formik, FormikHelpers } from 'formik'
import { getExperimentsState } from 'openapi'
import { useIntl } from 'react-intl'

import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch, useStoreSelector } from 'store'

import { setAlert, setTokenName, setTokens } from 'slices/globalStatus'

import { Submit, TextField } from 'components/FormField'
import i18n from 'components/T'

import { validateName } from 'lib/formikhelpers'

function validateToken(value: string) {
  let error

  if (value === '') {
    error = i18n('settings.addToken.tokenValidation') as unknown as string
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
  const intl = useIntl()

  const { tokens } = useStoreSelector((state) => state.globalStatus)
  const dispatch = useStoreDispatch()

  const saveToken = (values: TokenFormValues) => {
    dispatch(setTokens([...tokens, values]))
    dispatch(setTokenName(values.name))
  }

  const submitToken = (values: TokenFormValues, { setFieldError, resetForm }: FormikHelpers<TokenFormValues>) => {
    if (tokens.some((token) => token.name === values.name)) {
      dispatch(
        setAlert({
          type: 'warning',
          message: i18n('settings.addToken.duplicateDesc', intl),
        })
      )

      return
    }

    applyAPIAuthentication(values.token)

    function restSteps() {
      saveToken(values)

      typeof onSubmitCallback === 'function' && onSubmitCallback(values)

      resetForm()
    }

    // Test the validity of the token in advance
    getExperimentsState()
      .then(restSteps)
      .catch((error) => {
        const data = error.response?.data

        if (data && data.code === 'error.api.invalid_request' && data.message.includes('Unauthorized')) {
          setFieldError('token', 'Please check the validity of the token')

          resetAPIAuthentication()

          return
        }

        restSteps()
      })
  }

  return (
    <Formik initialValues={{ name: '', token: '' }} onSubmit={submitToken}>
      {({ errors, touched }) => (
        <Form>
          <Space>
            <TextField
              name="name"
              label={i18n('common.name')}
              validate={validateName(i18n('settings.addToken.nameValidation') as unknown as string)}
              helperText={errors.name && touched.name ? errors.name : i18n('settings.addToken.nameHelper')}
              error={errors.name && touched.name ? true : false}
            />
            <TextField
              name="token"
              label={i18n('settings.addToken.token')}
              multiline
              rows={12}
              validate={validateToken}
              helperText={errors.token && touched.token ? errors.token : i18n('settings.addToken.tokenHelper')}
              error={errors.token && touched.token ? true : false}
            />
          </Space>
          <Submit fullWidth />
        </Form>
      )}
    </Formik>
  )
}

export default Token
