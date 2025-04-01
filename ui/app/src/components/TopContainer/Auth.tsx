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
import GoogleIcon from '@mui/icons-material/Google'
import { Box, Button, Divider, IconButton, Link, Typography, createSvgIcon } from '@mui/material'
import { Stale } from 'api/queryUtils'
import { useGetCommonConfig } from 'openapi'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import ConfirmDialog from '@ui/mui-extends/esm/ConfirmDialog'
import Space from '@ui/mui-extends/esm/Space'

import RBACGenerator from 'components/RBACGenerator'
import i18n from 'components/T'
import Token from 'components/Token'

const OpenIdIcon = createSvgIcon(
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 25.573 25.573">
    <g>
      <polygon points="12.036,24.589 12.036,3.296 15.391,0.983 15.391,22.74" />
      <path d="M11.11,7.926v2.893c0,0-6.632,0.521-7.058,5.556c0,0-0.93,4.396,7.058,5.785v2.43c0,0-11.226-1.155-11.109-8.331C0.001,16.258-0.115,8.968,11.11,7.926z" />
      <path d="M16.2,7.926v2.702c0,0,2.142-0.029,3.934,1.463l-1.964,0.807l7.403,1.855V8.967l-2.527,1.43C23.046,10.397,20.889,8.13,16.2,7.926z" />
    </g>
  </svg>,
  'OpenId'
)

interface AuthProps {
  open: boolean
}

const Auth: React.FC<AuthProps> = ({ open }) => {
  const navigate = useNavigate()

  const { data: config } = useGetCommonConfig({
    query: {
      enabled: false,
      staleTime: Stale.DAY,
    },
  })

  const [tokenGenOpen, setTokenGenOpen] = useState(false)

  const handleSubmitCallback = () => navigate(0)
  const handleAuthGCP = () => (window.location.href = '/api/auth/gcp/redirect')
  const handleAuthOIDC = () => (window.location.href = '/api/auth/oidc/redirect')

  return (
    <ConfirmDialog
      open={open}
      title={i18n('settings.addToken.prompt')}
      dialogProps={{
        disableEscapeKeyDown: true,
        PaperProps: {
          style: { width: 512 },
        },
      }}
    >
      <Space>
        <Typography variant="body2" color="textSecondary">
          {i18n('settings.addToken.prompt2')}
          <Link sx={{ cursor: 'pointer' }} onClick={() => setTokenGenOpen(true)}>
            {i18n('settings.addToken.prompt3')}
          </Link>
        </Typography>
        <Token onSubmitCallback={handleSubmitCallback} />
      </Space>
      {config?.gcp_security_mode && (
        <>
          <Divider sx={{ mt: 6, mb: 3, color: 'text.secondary', typography: 'body2' }}>
            {i18n('settings.addToken.or')}
          </Divider>
          <Box textAlign="center">
            <IconButton color="primary" onClick={handleAuthGCP}>
              <GoogleIcon />
            </IconButton>
          </Box>
        </>
      )}
      {config?.oidc_security_mode && (
        <>
          <Divider sx={{ mt: 6, mb: 3, color: 'text.secondary', typography: 'body2' }}>
            {i18n('settings.addToken.or')}
          </Divider>
          <Box textAlign="center">
            <IconButton color="primary" onClick={handleAuthOIDC} title="OIDC">
              <OpenIdIcon />
            </IconButton>
          </Box>
        </>
      )}

      <ConfirmDialog
        open={tokenGenOpen}
        title={i18n('settings.addToken.generator')}
        dialogProps={{
          PaperProps: {
            style: { width: 750, maxWidth: 'unset' }, // max-width: 600
          },
        }}
      >
        <RBACGenerator />
        <Box mt={3} textAlign="right">
          <Button onClick={() => setTokenGenOpen(false)}>{i18n('common.close')}</Button>
        </Box>
      </ConfirmDialog>
    </ConfirmDialog>
  )
}

export default Auth
