import { Box, Button, Link, Typography } from '@material-ui/core'
import ConfirmDialog, { ConfirmDialogHandles } from 'components-mui/ConfirmDialog'
import { useEffect, useRef } from 'react'
import { useHistory, useLocation } from 'react-router-dom'

import RBACGenerator from 'components/RBACGenerator'
import T from 'components/T'
import Token from 'components/Token'

interface AuthProps {
  open: boolean
}

const Auth: React.FC<AuthProps> = ({ open }) => {
  const history = useHistory()
  const { pathname } = useLocation()

  const confirmRef = useRef<ConfirmDialogHandles>(null)
  const confirmRefRBAC = useRef<ConfirmDialogHandles>(null)

  useEffect(() => {
    confirmRef.current!.setOpen(open)
  }, [open])

  const handleSubmitCallback = () => {
    confirmRef.current!.setOpen(false)

    history.replace('/authed')
    setTimeout(() => history.replace(pathname))
  }

  const openGenerator = () => confirmRefRBAC.current!.setOpen(true)
  const closeGenerator = () => confirmRefRBAC.current!.setOpen(false)

  return (
    <ConfirmDialog
      ref={confirmRef}
      title={T('settings.addToken.prompt')}
      dialogProps={{
        disableBackdropClick: true,
        disableEscapeKeyDown: true,
        PaperProps: {
          style: { width: 512 },
        },
      }}
    >
      <Box mb={3}>
        <Typography variant="body2" color="textSecondary">
          {T('settings.addToken.prompt2')}{' '}
          <Link style={{ cursor: 'pointer' }} onClick={openGenerator}>
            {T('settings.addToken.prompt3')}
          </Link>
        </Typography>
      </Box>
      <Token onSubmitCallback={handleSubmitCallback} />
      <ConfirmDialog
        ref={confirmRefRBAC}
        title={T('settings.addToken.generator')}
        dialogProps={{
          PaperProps: {
            style: { width: 750, maxWidth: 'unset' }, // max-width: 600
          },
        }}
      >
        <RBACGenerator />
        <Box textAlign="right">
          <Button onClick={closeGenerator}>{T('common.close')}</Button>
        </Box>
      </ConfirmDialog>
    </ConfirmDialog>
  )
}

export default Auth
