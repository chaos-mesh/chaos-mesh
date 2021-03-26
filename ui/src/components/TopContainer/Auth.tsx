import { Box, Button, Link, Typography } from '@material-ui/core'
import Token, { TokenFormValues } from 'components/Token'
import { useEffect, useState } from 'react'
import { useHistory, useLocation } from 'react-router-dom'

import ConfirmDialog from 'components-mui/ConfirmDialog'
import RBACGenerator from 'components/RBACGenerator'
import T from 'components/T'
import { setTokenName } from 'slices/globalStatus'
import { useStoreDispatch } from 'store'

interface AuthProps {
  open: boolean
}

const Auth: React.FC<AuthProps> = ({ open }) => {
  const history = useHistory()
  const { pathname } = useLocation()

  const dispatch = useStoreDispatch()

  const [title, setTitle] = useState<string | JSX.Element>('')
  const [genTitle, setGenTitle] = useState<string | JSX.Element>('')

  useEffect(() => {
    setTitle(open ? T('settings.addToken.prompt') : '')
  }, [open])

  const handleSubmitCallback = (values: TokenFormValues) => {
    setTitle('')

    dispatch(setTokenName(values.name))

    history.replace('/authed')
    setTimeout(() => history.replace(pathname))
  }

  const openGenerator = () => setGenTitle(T('settings.addToken.generator'))
  const closeGenerator = () => setGenTitle('')

  return (
    <ConfirmDialog
      title={title}
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
      {genTitle && (
        <ConfirmDialog
          title={genTitle}
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
      )}
    </ConfirmDialog>
  )
}

export default Auth
