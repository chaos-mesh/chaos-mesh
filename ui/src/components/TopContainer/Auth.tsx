import { Box, Button, Link, Typography } from '@material-ui/core'
import { useEffect, useState } from 'react'

import ConfirmDialog from 'components-mui/ConfirmDialog'
import RBACGenerator from 'components/RBACGenerator'
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
      <Box mb={3}>
        <Typography variant="body2" color="textSecondary">
          {T('settings.addToken.prompt2')}{' '}
          <Link style={{ cursor: 'pointer' }} onClick={() => setTokenGenOpen(true)}>
            {T('settings.addToken.prompt3')}
          </Link>
        </Typography>
      </Box>
      <Token onSubmitCallback={handleSubmitCallback} />
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
