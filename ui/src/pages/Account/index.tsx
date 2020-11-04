import * as Yup from 'yup'

import { Box, Button, Container, Paper } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { setAlert, setAlertOpen } from 'slices/globalStatus'

import PaperTop from 'components/PaperTop'
import React from 'react'
import T from 'components/T'
import TextField from 'components/FormField/TextField'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'
import { useTokenHandler } from 'lib/token'

const Account = () => {
  const tokenSubmitHandler = useTokenHandler()
  const dispatch = useStoreDispatch()
  const intl = useIntl()

  const sumbitHandler = (values: { token: string }) => {
    const { token } = values
    tokenSubmitHandler(token)
    dispatch(
      setAlert({
        type: 'success',
        message: intl.formatMessage({ id: 'common.updateSuccessfully' }),
      })
    )
    dispatch(setAlertOpen(true))
  }

  return (
    <Paper variant="outlined" style={{ height: '100%' }}>
      <PaperTop title={T('account.title')} />
      <Container>
        <Formik
          initialValues={{ token: '' }}
          validationSchema={Yup.object({
            token: Yup.string().required(intl.formatMessage({ id: 'account.tokenValidation' })),
          })}
          onSubmit={sumbitHandler}
        >
          {({ errors, values }) => {
            return (
              <Form>
                <Box p={6} width={400} maxWidth="100%">
                  <Box mb={4}>
                    <TextField
                      name="token"
                      variant="outlined"
                      margin="dense"
                      fullWidth
                      label={T('account.token')}
                      error={!!errors.token}
                      helperText={errors.token || intl.formatMessage({ id: 'account.tokenChangeHelper' })}
                    ></TextField>
                  </Box>
                  <Box>
                    <Button type="submit" variant="outlined" color="primary" disabled={!!errors.token || !values.token}>
                      {T('common.submit')}
                    </Button>
                  </Box>
                </Box>
              </Form>
            )
          }}
        </Formik>
      </Container>
    </Paper>
  )
}

export default Account
