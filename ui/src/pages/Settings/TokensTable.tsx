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
import { Button, Checkbox, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core'
import { setConfirm, setConfirmOpen, setTokenName, setTokens } from 'slices/globalStatus'
import { useStoreDispatch, useStoreSelector } from 'store'

import LS from 'lib/localStorage'
import PaperContainer from 'components-mui/PaperContainer'
import T from 'components/T'
import { TokenFormValues } from 'components/Token'
import api from 'api'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'

const TokensTable = () => {
  const intl = useIntl()
  const history = useHistory()

  const { tokens, tokenName } = useStoreSelector((state) => state.globalStatus)
  const dispatch = useStoreDispatch()

  const handleUseToken = (token: TokenFormValues) => () => {
    dispatch(setTokenName(token.name))
    api.auth.token(token.token)
  }

  const handleRemoveToken = (token: TokenFormValues) => () =>
    dispatch(
      setConfirm({
        title: `${T('common.delete', intl)} ${token.name}`,
        description: T('common.deleteDesc', intl),
        handle: handleRemoveTokenConfirm(token.name),
      })
    )

  const handleRemoveTokenConfirm = (n: string) => () => {
    const current = tokens.filter(({ name }) => name !== n)

    if (current.length) {
      dispatch(setConfirmOpen(false))

      dispatch(setTokens(current))

      if (n === tokenName) {
        api.auth.resetToken()
        handleUseToken(current[0])()
      }
    } else {
      LS.remove('token')
      LS.remove('token-name')
      history.go(0)
    }
  }

  return (
    <TableContainer component={PaperContainer}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell />
            <TableCell>{T('common.name')}</TableCell>
            <TableCell>{T('settings.addToken.token')}</TableCell>
            <TableCell>{T('common.status')}</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {tokens.map((token) => {
            const key = `${token.name}:${token.token}`

            return (
              <TableRow key={key}>
                <TableCell padding="checkbox">
                  <Checkbox indeterminate checked={true} onChange={handleRemoveToken(token)} />
                </TableCell>
                <TableCell>{token.name}</TableCell>
                <TableCell>{'*'.repeat(12)}</TableCell>
                <TableCell>
                  <Button
                    onClick={handleUseToken(token)}
                    variant="outlined"
                    color="primary"
                    size="small"
                    disabled={token.name === tokenName}
                  >
                    {token.name === tokenName ? T('common.using') : T('common.use')}
                  </Button>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </TableContainer>
  )
}

export default TokensTable
