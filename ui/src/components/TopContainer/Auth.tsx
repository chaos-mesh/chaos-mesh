import { Box, Button, Divider, IconButton, Link, Typography } from '@material-ui/core'
import { useEffect, useState } from 'react'

import ConfirmDialog from 'components-mui/ConfirmDialog'
import GoogleIcon from '@material-ui/icons/Google'
import RBACGenerator from 'components/RBACGenerator'
import Space from 'components-mui/Space'
import T from 'components/T'
import Token from 'components/Token'
import { useHistory } from 'react-router-dom'

interface AuthProps {
  open: boolean
  setOpen: React.Dispatch<React.SetStateAction<boolean>>
}

const Auth: React.FC<AuthProps> = ({ open, setOpen }) => {
  const history = useHistory()

  const [tokenGenOpen, setTokenGenOpen] = useState(false)

  useEffect(() => {
    setOpen(open)
  }, [open, setOpen])

  const handleSubmitCallback = () => history.go(0)
  const handleAuthGCP = () => (window.location.href = '/api/auth/gcp/redirect')

  return (
    <ConfirmDialog
      open={open}
      title={T('settings.addToken.prompt')}
      dialogProps={{
        disableEscapeKeyDown: true,
        PaperProps: {
          style: { width: 512 },
        },
      }}
    >
      <Space>
        <Typography variant="body2" color="textSecondary">
          {T('settings.addToken.prompt2')}
          <Link sx={{ cursor: 'pointer' }} onClick={() => setTokenGenOpen(true)}>
            {T('settings.addToken.prompt3')}
          </Link>
        </Typography>
        <Token onSubmitCallback={handleSubmitCallback} />
      </Space>
      <Divider sx={{ mt: 6, mb: 3, color: 'text.secondary', typography: 'body2' }}>{T('settings.addToken.or')}</Divider>
      <Box textAlign="center">
        <IconButton color="primary" onClick={handleAuthGCP}>
          <GoogleIcon />
        </IconButton>
      </Box>

      <ConfirmDialog
        open={tokenGenOpen}
        title={T('settings.addToken.generator')}
        dialogProps={{
          PaperProps: {
            style: { width: 750, maxWidth: 'unset' }, // max-width: 600
          },
        }}
      >
        <RBACGenerator />
        <Box mt={3} textAlign="right">
          <Button onClick={() => setTokenGenOpen(false)}>{T('common.close')}</Button>
        </Box>
      </ConfirmDialog>
    </ConfirmDialog>
  )
}

export default Auth
