/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { Box, Button, Divider, IconButton, Link, Typography } from '@mui/material'
import { useEffect, useState } from 'react'

import ConfirmDialog from '@ui/mui-extends/esm/ConfirmDialog'
import GoogleIcon from '@mui/icons-material/Google'
import RBACGenerator from 'components/RBACGenerator'
import Space from '@ui/mui-extends/esm/Space'
import T from 'components/T'
import Token from 'components/Token'
import { useNavigate } from 'react-router-dom'
import { useStoreSelector } from 'store'

interface AuthProps {
  open: boolean
  setOpen: React.Dispatch<React.SetStateAction<boolean>>
}

const Auth: React.FC<AuthProps> = ({ open, setOpen }) => {
  const navigate = useNavigate()

  const { gcpSecurityMode } = useStoreSelector((state) => state.globalStatus)

  const [tokenGenOpen, setTokenGenOpen] = useState(false)

  useEffect(() => {
    setOpen(open)
  }, [open, setOpen])

  const handleSubmitCallback = () => navigate(0)
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
      {gcpSecurityMode && (
        <>
          <Divider sx={{ mt: 6, mb: 3, color: 'text.secondary', typography: 'body2' }}>
            {T('settings.addToken.or')}
          </Divider>
          <Box textAlign="center">
            <IconButton color="primary" onClick={handleAuthGCP}>
              <GoogleIcon />
            </IconButton>
          </Box>
        </>
      )}

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
