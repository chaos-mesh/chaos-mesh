import * as Yup from 'yup'

import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  MenuItem,
} from '@material-ui/core'
import { Form, Formik } from 'formik'
import React, { useEffect, useState } from 'react'
import { setAlert, setAlertOpen, setHasPrivilege, setIsPrivilegedToken, setIsValidToken } from 'slices/globalStatus'
import { useNameSpaceRegistry, useTokenRegistry } from 'lib/auth'

import { RootState } from 'store'
import { Route } from 'react-router-dom'
import T from 'components/T'
import TextField from 'components/FormField/TextField'
import { getNamespaces } from 'slices/experiments'
import { useIntl } from 'react-intl'
import { useSelector } from 'react-redux'
import { useStoreDispatch } from 'store'

const PrivateRoute: React.FC = ({ children, ...props }) => {
  const intl = useIntl()
  const dispatch = useStoreDispatch()

  const tokenIntercepterNumber = useSelector((state: RootState) => state.globalStatus.tokenInterceptorNumber)
  const hasPrivilege = useSelector((state: RootState) => state.globalStatus.hasPrivilege)
  const isValidToken = useSelector((state: RootState) => state.globalStatus.isValidToken)
  const isPrivilegedToken = useSelector((state: RootState) => state.globalStatus.isPrivilegedToken)
  const hasLoggedIn = tokenIntercepterNumber > -1 || window.localStorage.getItem('chaos-mesh-token')

  const { namespaces } = useSelector((state: RootState) => state.experiments)

  const tokenSubmitHandler = useTokenRegistry()
  const nsSubmitHandler = useNameSpaceRegistry()

  const loginHandler = (values: { token: string }) => {
    const { token } = values
    tokenSubmitHandler(token)
    dispatch(
      setAlert({
        type: 'success',
        message: intl.formatMessage({ id: 'common.submitSuccessfully' }),
      })
    )
    dispatch(setAlertOpen(true))
    dispatch(setIsValidToken(true))
    dispatch(setHasPrivilege(true))
    window.location.reload()
  }

  const nsHandler = (values: { namespace: string }) => {
    const { namespace } = values
    nsSubmitHandler(namespace)
    dispatch(
      setAlert({
        type: 'success',
        message: intl.formatMessage({ id: 'common.submitSuccessfully' }),
      })
    )
    dispatch(setAlertOpen(true))
    dispatch(setHasPrivilege(true))
    window.location.reload()
  }

  useEffect(() => {
    dispatch(getNamespaces())
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <Route
      {...props}
      render={(location) =>
        hasLoggedIn && isValidToken && isPrivilegedToken ? (
          hasPrivilege ? (
            children
          ) : (
            <>
              {children}
              <Dialog
                open={true}
                disableBackdropClick
                disableEscapeKeyDown
                PaperProps={{ variant: 'outlined' }}
                maxWidth="md"
                fullWidth
                style={{ bottom: '20vh' }}
              >
                <DialogTitle>{T('account.namespaceInputHelper')}</DialogTitle>
                <Formik
                  initialValues={{ namespace: '' }}
                  validationSchema={Yup.object({
                    namespace: Yup.string().required(intl.formatMessage({ id: 'account.namespaceValidation' })),
                  })}
                  onSubmit={nsHandler}
                >
                  {({ errors, values }) => {
                    return (
                      <Form>
                        <DialogContent>
                          <DialogContentText>{T('account.namespaceTips')}</DialogContentText>
                          <TextField
                            select
                            margin="dense"
                            name="namespace"
                            label={intl.formatMessage({ id: 'k8s.namespaces' })}
                            fullWidth
                          >
                            {namespaces.map((ns) => (
                              <MenuItem key={ns} value={ns}>
                                {ns}
                              </MenuItem>
                            ))}
                          </TextField>
                        </DialogContent>
                        <DialogActions>
                          <Button type="submit" color="primary" disabled={!values.namespace}>
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
        ) : (
          <>
            {children}
            <Dialog
              open={true}
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
                onSubmit={loginHandler}
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
