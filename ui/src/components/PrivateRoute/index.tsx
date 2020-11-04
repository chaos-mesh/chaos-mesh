import * as Yup from 'yup'

import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { setAlert, setAlertOpen } from 'slices/globalStatus'

import React from 'react'
import { RootState } from 'store'
import { Route } from 'react-router-dom'
import T from 'components/T'
import TextField from 'components/FormField/TextField'
import { useIntl } from 'react-intl'
import { useSelector } from 'react-redux'
import { useStoreDispatch } from 'store'
import { useTokenHandler } from 'lib/token'

const PrivateRoute: React.FC = ({ children, ...props }) => {
  const intl = useIntl()
  const dispatch = useStoreDispatch()

  const tokenIntercepterNumber = useSelector((state: RootState) => state.globalStatus.tokenIntercepterNumber)
  const hasLoggedIn = tokenIntercepterNumber > -1 || window.localStorage.getItem('chaos-mesh-token')

  const [open, setOpen] = React.useState(true)

  const tokenSubmitHandler = useTokenHandler()

  const handleClose = () => {
    setOpen(false)
  }

  const sumbitHandler = (values: { token: string }) => {
    const { token } = values
    tokenSubmitHandler(token)
    dispatch(
      setAlert({
        type: 'success',
        message: intl.formatMessage({ id: 'common.submitSuccessfully' }),
      })
    )
    dispatch(setAlertOpen(true))
    handleClose()
  }

  return (
    <Route
      {...props}
      render={(location) =>
        hasLoggedIn ? (
          children
        ) : (
          <>
            {children}
            <Dialog
              open={open}
              onClose={handleClose}
              disableBackdropClick
              disableEscapeKeyDown
              PaperProps={{ variant: 'outlined' }}
              maxWidth="md"
              fullWidth
              style={{ bottom: '20vh' }}
            >
              <DialogTitle>{T('account.tokenInputHelper')}</DialogTitle>
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
                      <DialogContent>
                        <DialogContentText>{T('account.tokenTips')}</DialogContentText>
                        <TextField
                          autoFocus
                          margin="dense"
                          name="token"
                          label={intl.formatMessage({ id: 'account.token' })}
                          fullWidth
                          error={!!errors.token}
                          helperText={errors.token}
                        />
                      </DialogContent>
                      <DialogActions>
                        <Button type="submit" color="primary" disabled={!!errors.token || !values.token}>
                          {T('common.submit')}
                        </Button>
                      </DialogActions>
                    </Form>
                  )
                }}
              </Formik>
            </Dialog>
          </>
        )
      }
    />
  )
}

export default PrivateRoute
