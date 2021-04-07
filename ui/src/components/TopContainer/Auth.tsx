import { useHistory, useLocation } from 'react-router-dom'

import ConfirmDialog from 'components-mui/ConfirmDialog'
import React from 'react'
import T from 'components/T'
import Token from 'components/Token'

interface AuthProps {
  open: boolean
  setOpen: (open: boolean) => void
}

const Auth: React.FC<AuthProps> = ({ open, setOpen }) => {
  const history = useHistory()
  const { pathname } = useLocation()

  const handleSubmitCallback = () => {
    setOpen(false)

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
