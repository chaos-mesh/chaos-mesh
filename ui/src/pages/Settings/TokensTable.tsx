import { Button, Checkbox, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core'
import ConfirmDialog, { ConfirmDialogHandles } from 'components-mui/ConfirmDialog'
import React, { useRef, useState } from 'react'
import { setTokenName, setTokens } from 'slices/globalStatus'
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

  const [selected, setSelected] = useState({
    tokenName: '',
    title: '',
    description: '',
  })
  const confirmRef = useRef<ConfirmDialogHandles>(null)

  const handleUseToken = (token: TokenFormValues) => () => {
    dispatch(setTokenName(token.name))
    api.auth.token(token.token)
  }

  const handleRemoveToken = (token: TokenFormValues) => () => {
    setSelected({
      tokenName: token.name,
      title: `${intl.formatMessage({ id: 'common.delete' })} ${token.name}`,
      description: intl.formatMessage({ id: 'settings.addToken.deleteDesc' }),
    })

    confirmRef.current!.setOpen(true)
  }

  const handleRemoveTokenConfirm = () => {
    const current = tokens.filter(({ name }) => name !== selected.tokenName)

    if (current.length) {
      confirmRef.current!.setOpen(false)

      dispatch(setTokens(current))

      if (selected.tokenName === tokenName) {
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
    <>
      <TableContainer component={PaperContainer}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell />
              <TableCell>{T('settings.addToken.name')}</TableCell>
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

      <ConfirmDialog
        ref={confirmRef}
        title={selected.title}
        description={selected.description}
        onConfirm={handleRemoveTokenConfirm}
      />
    </>
  )
}

export default TokensTable
