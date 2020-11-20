import Token, { TokenFormValues } from 'components/Token'
import { useHistory, useLocation } from 'react-router-dom'

import ConfirmDialog from 'components-mui/ConfirmDialog'
import React from 'react'
import T from 'components/T'
import { setTokenName } from 'slices/globalStatus'
import { useStoreDispatch } from 'store'

interface AuthProps {
  open: boolean
  setOpen: (open: boolean) => void
}

const Auth: React.FC<AuthProps> = ({ open, setOpen }) => {
  const history = useHistory()
  const { pathname } = useLocation()

  const dispatch = useStoreDispatch()

  const handleSubmitCallback = (values: TokenFormValues) => {
    setOpen(false)

    dispatch(setTokenName(values.name))

    history.replace('/authed')
    setTimeout(() => history.replace(pathname))
  }

  return (
    <ConfirmDialog
      open={open}
      setOpen={setOpen}
      title={T('settings.addToken.prompt')}
      dialogProps={{
        disableBackdropClick: true,
        disableEscapeKeyDown: true,
        PaperProps: {
          variant: 'outlined',
          style: { width: 500, minWidth: 300 },
        },
      }}
    >
      <Token onSubmitCallback={handleSubmitCallback} />
    </ConfirmDialog>
  )
}

export default Auth
