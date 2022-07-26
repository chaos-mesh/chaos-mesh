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
import { Box, Button } from '@mui/material'
import Cookies from 'js-cookie'
import _ from 'lodash'
import { useIntl } from 'react-intl'
import { useNavigate } from 'react-router-dom'

import PaperTop from '@ui/mui-extends/esm/PaperTop'

import { useStoreDispatch, useStoreSelector } from 'store'

import { setConfirm } from 'slices/globalStatus'

import i18n from 'components/T'

import LS from 'lib/localStorage'

const Token = () => {
  const navigate = useNavigate()
  const intl = useIntl()

  const { tokens, tokenName } = useStoreSelector((state) => state.globalStatus)
  const tokenDesc =
    tokenName === 'gcp' ? (
      <Box display="flex" alignItems="center">
        {i18n('settings.addToken.gcp')}
        <GoogleIcon sx={{ ml: 1 }} />
      </Box>
    ) : (
      tokenName + ': ' + _.truncate(tokens[0].token)
    )
  const dispatch = useStoreDispatch()

  const handleRemoveToken = () =>
    dispatch(
      setConfirm({
        title: i18n('common.logout', intl),
        description: i18n('common.logoutDesc', intl),
        handle: handleRemoveTokenConfirm,
      })
    )

  const handleRemoveTokenConfirm = () => {
    if (tokenName === 'gcp') {
      Cookies.remove('access_token')
      Cookies.remove('refresh_token')
      Cookies.remove('expiry')
    } else {
      LS.remove('token')
      LS.remove('token-name')
    }

    navigate(0)
  }

  return (
    <PaperTop title={i18n('settings.addToken.token')} subtitle={tokenDesc}>
      <Button
        variant="outlined"
        size="small"
        color="secondary"
        sx={{ width: 64, height: 32 }}
        onClick={handleRemoveToken}
      >
        {i18n('common.logout')}
      </Button>
    </PaperTop>
  )
}

export default Token
