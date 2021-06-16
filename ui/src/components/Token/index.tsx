import { Form, Formik, FormikHelpers } from 'formik'
import { Submit, TextField } from 'components/FormField'
import { setAlert, setTokenName, setTokens } from 'slices/globalStatus'
import { useStoreDispatch, useStoreSelector } from 'store'

import Space from 'components-mui/Space'
import T from 'components/T'
import api from 'api'
import { useIntl } from 'react-intl'
import { validateName } from 'lib/formikhelpers'

function validateToken(value: string) {
  let error

  if (value === '') {
    error = T('settings.addToken.tokenValidation') as unknown as string
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
          message: T('settings.addToken.duplicateDesc', intl),
        })
      )

      return
    }

    api.auth.token(values.token)

    function restSteps() {
      saveToken(values)

      typeof onSubmitCallback === 'function' && onSubmitCallback(values)

      resetForm()
    }

    // Test the validity of the token in advance
    api.experiments
      .state()
      .then(restSteps)
      .catch((error) => {
        const data = error.response?.data

        if (data && data.code === 'error.api.invalid_request' && data.message.includes('Unauthorized')) {
          setFieldError('token', 'Please check the validity of the token')

          api.auth.resetToken()

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
              label={T('common.name')}
              validate={validateName(T('settings.addToken.nameValidation') as unknown as string)}
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
            <Submit />
          </Space>
        </Form>
      )}
    </Formik>
  )
}

export default Token
